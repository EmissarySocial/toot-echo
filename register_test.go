package tootecho

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benpate/toot"
	"github.com/benpate/toot/object"
	"github.com/benpate/toot/txn"
	"github.com/labstack/echo/v4"
)

// testToken satisfies toot.ScopesGetter with every scope granted, so that these
// tests exercise ROUTING rather than authorization.
type testToken struct{}

// Scopes grants the top-level read and write scopes.
func (testToken) Scopes() []string {
	return []string{"read", "write"}
}

// testAPI returns an API whose handlers each record which one was called.
func testAPI(called *string) toot.API[testToken] {

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	api.PostAccount = func(testToken, txn.PostAccount) (object.Token, error) {
		*called = "PostAccount"
		return object.Token{}, nil
	}

	api.PostAccount_Follow = func(testToken, txn.PostAccount_Follow) (object.Relationship, error) {
		*called = "PostAccount_Follow"
		return object.Relationship{}, nil
	}

	api.GetAccount_Relationships = func(testToken, txn.GetAccount_Relationships) ([]object.Relationship, error) {
		*called = "GetAccount_Relationships"
		return nil, nil
	}

	return api
}

// Every endpoint must reach its own handler. Registering two handlers on one path
// is silent in Echo -- the second simply replaces the first -- so a mis-typed
// route constant hijacks an unrelated endpoint instead of failing loudly.
func TestRegister_EachRouteReachesItsOwnHandler(t *testing.T) {

	testCases := []struct {
		method   string
		path     string
		expected string
	}{
		{http.MethodPost, "/api/v1/accounts", "PostAccount"},
		{http.MethodPost, "/api/v1/accounts/123/follow", "PostAccount_Follow"},
		{http.MethodGet, "/api/v1/accounts/relationships", "GetAccount_Relationships"},
	}

	for _, testCase := range testCases {

		called := ""
		e := echo.New()
		Register(e, testAPI(&called))

		request := httptest.NewRequest(testCase.method, testCase.path, strings.NewReader(""))
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

		recorder := httptest.NewRecorder()
		e.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("%s %s: expected 200, got %d (%s)",
				testCase.method, testCase.path, recorder.Code, recorder.Body.String())
			continue
		}

		if called != testCase.expected {
			t.Errorf("%s %s: expected handler %q, got %q",
				testCase.method, testCase.path, testCase.expected, called)
		}
	}
}

// A paged result writes its pagination links into the RESPONSE header.
func TestRegister_PagedResultSetsLinkHeader(t *testing.T) {

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	api.GetAccount_Followers = func(testToken, txn.GetAccount_Followers) ([]object.Account, toot.PageInfo, error) {
		return []object.Account{}, toot.PageInfo{MinID: "AAA", MaxID: "ZZZ"}, nil
	}

	e := echo.New()
	Register(e, api)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/accounts/123/followers?limit=20", nil))

	// Links are absolute, and keep the request's own limit
	expected := `<http://example.com/api/v1/accounts/123/followers?max_id=ZZZ&limit=20>; rel="next", ` +
		`<http://example.com/api/v1/accounts/123/followers?min_id=AAA&limit=20>; rel="prev"`

	if got := recorder.Header().Get("Link"); got != expected {
		t.Errorf("expected Link %q, got %q", expected, got)
	}
}

// Middleware handed to Register must actually run. The parameter existed but was
// never forwarded, so a caller's middleware was accepted and silently discarded.
func TestRegister_CallerMiddlewareRuns(t *testing.T) {

	ran := false

	marker := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ran = true
			return next(ctx)
		}
	}

	called := ""
	e := echo.New()
	Register(e, testAPI(&called), marker)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/accounts/relationships", nil))

	if !ran {
		t.Error("middleware passed to Register never ran")
	}
}

// A browser client has to be able to read an error body, not just a success body.
func TestRegister_CORSHeaderOnErrorResponse(t *testing.T) {

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, errors.New("not authorized")
	})

	api.GetAccount_Relationships = func(testToken, txn.GetAccount_Relationships) ([]object.Relationship, error) {
		return nil, nil
	}

	e := echo.New()
	Register(e, api)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/accounts/relationships", nil))

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected CORS header on the error response, got %q", got)
	}
}

// Filters in the request (here only_media and exclude_replies) must reach the next page's link,
// or the next page comes back unfiltered.
func TestRegister_PagedResultKeepsFiltersOnLinks(t *testing.T) {

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	api.GetAccount_Statuses = func(testToken, txn.GetAccount_Statuses) ([]object.Status, toot.PageInfo, error) {
		return []object.Status{}, toot.PageInfo{MaxID: "ZZZ"}, nil
	}

	e := echo.New()
	Register(e, api)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/accounts/123/statuses?only_media=true&exclude_replies=true&max_id=OLD", nil))

	expected := `<http://example.com/api/v1/accounts/123/statuses?max_id=ZZZ&exclude_replies=true&only_media=true>; rel="next"`

	if got := recorder.Header().Get("Link"); got != expected {
		t.Errorf("expected Link %q, got %q", expected, got)
	}
}

// The favourites and bookmarks lists are paged too, so their Link header is written and the cursor round-trips.
func TestRegister_FavouritesAndBookmarksSetLinkHeader(t *testing.T) {

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	var receivedMaxID string

	api.GetFavourites = func(_ testToken, input txn.GetFavourites) ([]object.Status, toot.PageInfo, error) {
		receivedMaxID = input.MaxID
		return []object.Status{}, toot.PageInfo{MaxID: "1700"}, nil
	}

	api.GetBookmarks = func(testToken, txn.GetBookmarks) ([]object.Status, toot.PageInfo, error) {
		return []object.Status{}, toot.PageInfo{MaxID: "900"}, nil
	}

	e := echo.New()
	Register(e, api)

	for path, expected := range map[string]string{
		"/api/v1/favourites?limit=20": `<http://example.com/api/v1/favourites?max_id=1700&limit=20>; rel="next"`,
		"/api/v1/bookmarks?limit=20":  `<http://example.com/api/v1/bookmarks?max_id=900&limit=20>; rel="next"`,
	} {
		recorder := httptest.NewRecorder()
		e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		if got := recorder.Header().Get("Link"); got != expected {
			t.Errorf("%s: expected Link %q, got %q", path, expected, got)
		}
	}

	// A client following the link sends the cursor back
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/favourites?max_id=1700&limit=20", nil))

	if receivedMaxID != "1700" {
		t.Errorf("expected max_id 1700 to reach the handler, got %q", receivedMaxID)
	}
}
