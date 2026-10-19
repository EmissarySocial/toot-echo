package tootecho

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benpate/toot"
	"github.com/benpate/toot/object"
	"github.com/benpate/toot/txn"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// userToken lets a test tell who the handler thinks is asking.
type userToken struct{ User string }

func (userToken) Scopes() []string { return []string{"read", "write"} }

// A public route never refuses a request, but a valid token on it still tells the handler who is asking.
func TestRegister_PublicRouteUsesTokenWhenValid(t *testing.T) {

	var seen userToken

	api := toot.New(func(request *http.Request) (userToken, error) {

		if request.Header.Get("Authorization") == "Bearer good" {
			return userToken{User: "sherif"}, nil
		}

		return userToken{}, errors.New("not a valid token")
	})

	api.GetAccount = func(token userToken, _ txn.GetAccount) (object.Account, error) {
		seen = token
		return object.Account{ID: "x"}, nil
	}

	e := echo.New()
	Register(e, api)

	request := func(header string) int {

		seen = userToken{User: "unset"}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/123", nil)

		if header != "" {
			req.Header.Set(echo.HeaderAuthorization, header)
		}

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	require.Equal(t, http.StatusOK, request("Bearer good"))
	require.Equal(t, "sherif", seen.User, "a valid token identifies the caller")

	require.Equal(t, http.StatusOK, request(""))
	require.Equal(t, "", seen.User, "no token means an anonymous caller, and the request still works")

	require.Equal(t, http.StatusOK, request("Bearer expired"))
	require.Equal(t, "", seen.User, "a bad token on a public route is ignored, not an error")
}
