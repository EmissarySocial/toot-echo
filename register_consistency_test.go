package tootecho

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

/******************************************
 * Registration Consistency
 *
 * Every line of Register names the same endpoint four times -- an Echo method, a
 * route constant, a handler field, and a scope constant. Nothing makes the four
 * agree, and a mismatch still compiles: it simply registers the wrong path, or
 * guards a route with another route's scope. These tests read Register and check
 * that the four halves line up.
 ******************************************/

// registration matches one line of Register, capturing its four halves. GET routes register through
// the "get" wrapper (GET plus HEAD), so that name counts as GET.
var registration = regexp.MustCompile(
	`(single_result|paged_result)\(api, (?:e\.)?([A-Za-z]+), route\.(\w+), api\.(\w+), scope\.(\w+)`)

// methodForPrefix maps the verb that begins an endpoint name to its HTTP method.
var methodForPrefix = []struct {
	prefix string
	method string
}{
	{"Get", "GET"}, {"Post", "POST"}, {"Put", "PUT"},
	{"Patch", "PATCH"}, {"Delete", "DELETE"},
}

// readRegistrations returns every registration line found in register.go.
func readRegistrations(t *testing.T) [][]string {

	t.Helper()

	source, err := os.ReadFile("register.go")

	if err != nil {
		t.Fatalf("unable to read register.go: %v", err)
	}

	result := make([][]string, 0)

	for _, line := range strings.Split(string(source), "\n") {
		if match := registration.FindStringSubmatch(line); match != nil {
			if match[2] == "get" {
				match[2] = "GET"
			}
			result = append(result, match)
		}
	}

	if len(result) < 100 {
		t.Fatalf("expected to find the whole API, found only %d registrations", len(result))
	}

	return result
}

func TestRegistration_RouteHandlerAndScopeAgree(t *testing.T) {

	for _, match := range readRegistrations(t) {

		routeName, handlerName, scopeName := match[3], match[4], match[5]

		if routeName != handlerName {
			t.Errorf("route.%s is registered against api.%s", routeName, handlerName)
		}

		if routeName != scopeName {
			t.Errorf("route.%s is guarded by scope.%s", routeName, scopeName)
		}
	}
}

func TestRegistration_MethodMatchesTheEndpointName(t *testing.T) {

	for _, match := range readRegistrations(t) {

		method, routeName := match[2], match[3]

		for _, each := range methodForPrefix {

			if !strings.HasPrefix(routeName, each.prefix) {
				continue
			}

			if method != each.method {
				t.Errorf("%s is registered as e.%s, but its name says %s",
					routeName, method, each.method)
			}

			break
		}
	}
}

// Registering two handlers on one method and path is silent in Echo: the second
// replaces the first, so one endpoint disappears and another answers in its place.
func TestRegistration_NoPathIsRegisteredTwice(t *testing.T) {

	seen := make(map[string]string)

	for _, match := range readRegistrations(t) {

		key := match[2] + " " + match[3]

		if previous, ok := seen[key]; ok {
			t.Errorf("%s is registered by both %s and %s", key, previous, match[3])
		}

		seen[key] = match[3]
	}
}
