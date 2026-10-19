package tootecho

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benpate/toot"
	"github.com/benpate/toot/object"
	"github.com/benpate/toot/txn"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// pushToken grants the push scope that the push subscription routes require.
type pushToken struct{}

// Scopes grants the push scope.
func (pushToken) Scopes() []string {
	return []string{"push"}
}

// postPushSubscription sends a push subscription request and returns what the handler received.
func postPushSubscription(t *testing.T, query string, body string) (txn.PostPushSubscription, int) {

	t.Helper()

	var received txn.PostPushSubscription

	api := toot.New(func(*http.Request) (pushToken, error) {
		return pushToken{}, nil
	})

	api.PostPushSubscription = func(_ pushToken, input txn.PostPushSubscription) (object.WebPushSubscription, error) {
		received = input
		return object.WebPushSubscription{}, nil
	}

	e := echo.New()
	Register(e, api)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/push/subscription?"+query, strings.NewReader(body))

	if body != "" {
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	}

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	return received, recorder.Code
}

// TestQueryParams_FillFormFieldsOnWrites confirms a write that sends its parameters in the query string,
// with an empty body, reaches the handler with them filled in, as Mastodon's own server accepts.
func TestQueryParams_FillFormFieldsOnWrites(t *testing.T) {

	// The request Ice Cubes sends to register for push notifications
	query := "subscription%5Bendpoint%5D=https://relay.example/push/abc" +
		"&subscription%5Bkeys%5D%5Bp256dh%5D=p256dh-key" +
		"&subscription%5Bkeys%5D%5Bauth%5D=auth-key" +
		"&data%5Balerts%5D%5Bmention%5D=true&data%5Balerts%5D%5Bfollow%5D=true"

	received, status := postPushSubscription(t, query, "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "https://relay.example/push/abc", received.Subscription.Endpoint)
	require.Equal(t, "p256dh-key", received.Subscription.Keys.P256dh)
	require.Equal(t, "auth-key", received.Subscription.Keys.Auth)
	require.True(t, received.Data.Alerts.Mention)
	require.True(t, received.Data.Alerts.Follow)
	require.False(t, received.Data.Alerts.Reblog)
}

// TestQueryParams_BodyWinsOverQuery confirms that when both carry a value, the body's is used.
func TestQueryParams_BodyWinsOverQuery(t *testing.T) {

	received, status := postPushSubscription(t,
		"subscription%5Bendpoint%5D=https://from-query.example",
		"subscription%5Bendpoint%5D=https://from-body.example&subscription%5Bkeys%5D%5Bauth%5D=body-auth",
	)

	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "https://from-body.example", received.Subscription.Endpoint)
	require.Equal(t, "body-auth", received.Subscription.Keys.Auth)
}
