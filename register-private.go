package tootecho

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/benpate/derp"
	"github.com/benpate/rosetta/list"
	"github.com/benpate/toot"
	"github.com/benpate/toot/scope"
	"github.com/go-playground/form/v4"
	"github.com/labstack/echo/v4"
)

// formDecoder decodes url.Values (parsed form/multipart bodies) into Go structs,
// including nested structs and arrays via bracket notation (e.g. "subscription[endpoint]",
// "keywords_attributes[0][keyword]") -- something echo.DefaultBinder cannot do (see getInputs).
// A single shared *form.Decoder is safe for concurrent use and reuses its internal struct
// cache across requests, so it's created once here rather than per-request.
//
// go-playground/form defaults to dot-separated struct namespaces (e.g. "subscription.endpoint"),
// but Mastodon (like Rails) uses brackets for struct fields too, not just array/map indices
// (e.g. "subscription[endpoint]", "keywords_attributes[0][keyword]"). Reconfigure the
// namespace delimiters to match what real clients actually send.
var formDecoder = newFormDecoder()

func newFormDecoder() *form.Decoder {
	decoder := form.NewDecoder()
	decoder.SetNamespacePrefix("[")
	decoder.SetNamespaceSuffix("]")
	return decoder
}

// echoMethod represents an e.GET, e.POST, e.PUT, e.DELETE method that registers
// a new echo.HandlerFunc with the echo router.
type echoMethod func(string, echo.HandlerFunc, ...echo.MiddlewareFunc) *echo.Route

// commonWrapper is a function that can be used to handle both single results and paged results.
type commonWrapper[AuthToken toot.ScopesGetter, Input any, Output any] func(echo.Context, AuthToken, Input) (Output, error)

// single_result inserts a new echo.HandlerFunc into the echo router
// that returns a single result object (or an array without any paging metadata)
func single_result[AuthToken toot.ScopesGetter, Input any, Output any](api toot.API[AuthToken], fn echoMethod, path string, handler toot.APIFunc_SingleResult[AuthToken, Input, Output], requiredScope string, middleware ...echo.MiddlewareFunc) {

	// Do not register empty handlers
	if handler == nil {
		return
	}

	// Wrap the handler in a function that `getResult` can use
	wrapper := func(_ echo.Context, authToken AuthToken, input Input) (Output, error) {
		// Call the handler and return outputs to the caller
		return handler(authToken, input)
	}

	any_result(api, fn, path, wrapper, requiredScope, middleware...)
}

// register inserts a new echo.HandlerFunc into the echo router
// that returns a paged result object.
func paged_result[AuthToken toot.ScopesGetter, Input any, Output any](api toot.API[AuthToken], fn echoMethod, path string, handler toot.APIFunc_PagedResult[AuthToken, Input, Output], requiredScope string, middleware ...echo.MiddlewareFunc) {

	// Do not register empty handlers
	if handler == nil {
		return
	}

	// Wrap the handler in a function that `getResult` can use
	wrapper := func(ctx echo.Context, authToken AuthToken, input Input) (Output, error) {

		// Call the actual handler
		output, pageInfo, err := handler(authToken, input)

		// Apply paging headers to the response. These must be written into the
		// RESPONSE header: http.Request.Response is only populated for a client
		// following a redirect, and is always nil here on the server side.
		//
		// RULE: PageInfo.SetHeader prefixes the Link header with its "path"
		// argument verbatim, so it must be an absolute URL. Clients follow the
		// Link header as-is, and a bare path fails with "unsupported URL".
		//
		// The query string goes along too, so that each link keeps the request's filters.
		baseURL := ctx.Scheme() + "://" + ctx.Request().Host + ctx.Request().URL.RequestURI()
		pageInfo.SetHeader(ctx.Response().Header(), baseURL)

		// Return outputs to the caller
		return output, err
	}

	any_result(api, fn, path, wrapper, requiredScope, middleware...)
}

// any_result should not be called directly.  It is used by `single_result`
// and `paged_result` to inserts a new echo.HandlerFunc into the echo router.
// It requires a `commonWrapper` function to handle the actual request
func any_result[AuthToken toot.ScopesGetter, Input any, Output any](api toot.API[AuthToken], fn echoMethod, path string, wrapper commonWrapper[AuthToken, Input, Output], requiredScope string, middleware ...echo.MiddlewareFunc) {

	const location = "toot-echo.any_result"

	// Create a new echo.HandlerFunc that 1) parses inputs, 2) calls the actual handler, and
	// 3) generates a JSON response.
	tootHandler := func(ctx echo.Context) error {

		// Set the CORS header FIRST, so that a browser client can also read the body
		// of an error response.  Bearer tokens (not cookies) carry authorization here,
		// so a wildcard origin does not expose an authenticated session.
		ctx.Response().Header().Set("Access-Control-Allow-Origin", "*")

		// Parse inputs from the request
		authToken, input, err := getInputs[AuthToken, Input](ctx, api, requiredScope)

		if err != nil {
			return derp.Wrap(err, location, "Error parsing inputs")
		}

		// Call the actual handler to map the Inputs to Outputs
		result, err := wrapper(ctx, authToken, input)

		if err != nil {
			return derp.Wrap(err, location, "Error executing API call")
		}

		// Return the API result to the caller as JSON
		if err := ctx.JSON(http.StatusOK, result); err != nil {
			return derp.Wrap(err, location, "Error writing response body")
		}

		// Woot.
		return nil
	}

	// WithHost runs first, so that every handler sees a corrected Host header, and
	// the caller's own middleware runs after it.
	handlerMiddleware := append([]echo.MiddlewareFunc{WithHost}, middleware...)

	// Register the new echo.HandlerFunc with the echo Router
	fn(path, tootHandler, handlerMiddleware...)
}

func getInputs[AuthToken toot.ScopesGetter, Input any](ctx echo.Context, api toot.API[AuthToken], requiredScope string) (AuthToken, Input, error) {

	const location = "toot-echo.getInputs"

	var input Input
	var authToken AuthToken

	// If the request is not public (at least one scope is required)
	// then try to authorize the request.
	// A public request is open to anyone, so it is never refused; but when the
	// caller did send a valid token, the handler still gets to know who is asking.
	// Otherwise, an empty AuthToken will be passed to the handler.
	if requiredScope == scope.Public {

		if api.Authorize != nil {

			if token, err := api.Authorize(ctx.Request()); err == nil {
				authToken = token
			}
		}

	} else {

		var err error
		authToken, err = api.Authorize(ctx.Request())

		if err != nil {
			return authToken, input, derp.Wrap(err, location, "Request is not authorized. LOL.")
		}

		// Verify the scopes required for this API call
		if !verifyScope(authToken.Scopes(), requiredScope) {
			return authToken, input, derp.Unauthorized(location, "Request is not authorized.", requiredScope, authToken.Scopes())
		}
	}

	// Collect input arguments from the Request.
	//
	// echo.DefaultBinder handles query params and headers correctly, but its form-body binding
	// only matches flat keys against a field's tag string -- it cannot populate a nested struct
	// or a variable-length array of structs (Mastodon sends both: e.g. "subscription[endpoint]",
	// "keywords_attributes[0][keyword]"). A nested/array field silently stays zero-valued instead
	// of erroring, which is worse than a hard failure.
	//
	// So: query params and headers still go through Echo's binder (no known issues there), but
	// the request BODY -- where nested/array fields actually show up -- is decoded with
	// go-playground/form instead, which understands bracket notation for both nested structs
	// and arrays.
	//
	// Path params need their own follow-up step (decodePathParamStrings, below): Echo's
	// BindPathParams assigns each param:"..." field the *raw* path segment exactly as captured
	// by the router, with no percent-decoding (confirmed directly against Echo -- a path segment
	// of "%2F" comes back as the literal three characters, not "/"). That's fine for the common
	// case of opaque, unescaped IDs, but Mastodon account IDs can legitimately be full URLs (e.g.
	// Emissary uses a User's ActivityPub actor URL as its Mastodon account ID), and any correct
	// client percent-encodes such a value before putting it in a path segment.
	binder := echo.DefaultBinder{}

	if err := binder.BindPathParams(ctx, &input); err != nil {
		return authToken, input, derp.Wrap(err, location, "Unable to read path parameters")
	}

	if err := decodePathParamStrings(&input); err != nil {
		return authToken, input, derp.Wrap(err, location, "Unable to decode path parameters")
	}

	if err := binder.BindQueryParams(ctx, &input); err != nil {
		return authToken, input, derp.Wrap(err, location, "Unable to read query parameters")
	}

	if err := bindBody(ctx, &input); err != nil {
		return authToken, input, derp.Wrap(err, location, "Unable to read request body")
	}

	if err := binder.BindHeaders(ctx, &input); err != nil {
		return authToken, input, derp.Wrap(err, location, "Error reading headers")
	}

	// Return success
	return authToken, input, nil
}

// decodePathParamStrings percent-decodes every param:"..."-tagged string field on i, in place.
// See the comment above BindPathParams in getInputs for why this is needed: Echo assigns each
// such field the raw, still-percent-encoded path segment, which is wrong whenever that segment
// is meant to carry a value like a URL rather than an opaque token. This decodes every
// param-tagged string field, not just "id", since any of them could carry the same kind of
// value (e.g. a second ID in a nested resource route).
func decodePathParamStrings(i interface{}) error {

	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for index := 0; index < t.NumField(); index++ {

		field := t.Field(index)

		if _, hasParamTag := field.Tag.Lookup("param"); !hasParamTag {
			continue
		}

		fieldValue := v.Field(index)

		if fieldValue.Kind() != reflect.String || !fieldValue.CanSet() {
			continue
		}

		decoded, err := url.PathUnescape(fieldValue.String())

		if err != nil {
			return err
		}

		fieldValue.SetString(decoded)
	}

	return nil
}

// bindBody reads the request body into i. Form and multipart-form bodies go through
// formDecoder (which understands nested structs and arrays); everything else (JSON, XML,
// or no body at all) falls back to echo's own body binder.
func bindBody(ctx echo.Context, i interface{}) error {

	req := ctx.Request()

	if req.ContentLength == 0 {
		return nil
	}

	base, _, _ := strings.Cut(req.Header.Get(echo.HeaderContentType), ";")

	switch strings.TrimSpace(base) {

	case echo.MIMEApplicationForm:

		// RULE: Request.ParseForm only reads the body for POST/PUT/PATCH -- DELETE (Mastodon's
		// list/filter/keyword "remove" endpoints) needs its body parsed independent of method.
		body, err := io.ReadAll(req.Body)

		if err != nil {
			return err
		}

		values, err := url.ParseQuery(string(body))

		if err != nil {
			return err
		}

		return formDecoder.Decode(i, expandEmptyBracketArrays(values))

	case echo.MIMEMultipartForm:
		if err := req.ParseMultipartForm(32 << 20); err != nil {
			return err
		}

		if err := formDecoder.Decode(i, expandEmptyBracketArrays(req.MultipartForm.Value)); err != nil {
			return err
		}

		// go-playground/form decodes url.Values (text fields) -- it has no concept of an
		// uploaded file, which lives in req.MultipartForm.File instead. A *multipart.FileHeader
		// field (e.g. Mastodon's media upload) needs its own pass to bind those.
		return bindMultipartFiles(i, req.MultipartForm.File)

	default:
		// JSON, XML, or anything else -- echo already handles these correctly.
		binder := echo.DefaultBinder{}
		return binder.BindBody(ctx, i)
	}
}

// expandEmptyBracketArrays rewrites every "field[]" key (each value sent as its own repeated
// key, e.g. "media_ids[]=1&media_ids[]=2") into the indexed "field[0]", "field[1]", ... keys
// formDecoder actually understands. A key with no trailing "[]" passes through untouched.
//
// RULE: real Mastodon clients encode every array field this way; go-playground/form only
// understands an explicit index and errors on an empty one.
func expandEmptyBracketArrays(values url.Values) url.Values {

	result := make(url.Values, len(values))

	for key, fieldValues := range values {

		base, isArray := strings.CutSuffix(key, "[]")

		if !isArray {
			result[key] = fieldValues
			continue
		}

		for index, value := range fieldValues {
			indexedKey := base + "[" + strconv.Itoa(index) + "]"
			result[indexedKey] = append(result[indexedKey], value)
		}
	}

	return result
}

// fileHeaderType is *multipart.FileHeader, cached once rather than reflected on every call.
var fileHeaderType = reflect.TypeOf((*multipart.FileHeader)(nil))

// bindMultipartFiles sets every *multipart.FileHeader field on i (matched by its form tag)
// from the uploaded files a multipart request carried. A missing or unmatched field is left
// untouched, not an error.
func bindMultipartFiles(i interface{}, fileForm map[string][]*multipart.FileHeader) error {

	if len(fileForm) == 0 {
		return nil
	}

	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for index := 0; index < t.NumField(); index++ {

		field := t.Field(index)

		if field.Type != fileHeaderType {
			continue
		}

		tag, _, _ := strings.Cut(field.Tag.Get("form"), ",")

		if tag == "" || tag == "-" {
			continue
		}

		files := fileForm[tag]

		if len(files) == 0 {
			continue
		}

		v.Field(index).Set(reflect.ValueOf(files[0]))
	}

	return nil
}

// verifyScope confirms that the required scope exists in the
// `present` slice.
func verifyScope(present []string, requiredScope string) bool {

	// Always allow public requests
	if requiredScope == scope.Public {
		return true
	}

	// Since we're already authenticated, "private" requests
	// with no additional scope requirements are also allowed
	if requiredScope == scope.Private {
		return true
	}

	// If the required scope contains a colon, see if the user has just the "prefix" scope
	if prefix, suffix := list.Split(requiredScope, ':'); suffix != "" {
		for _, scope := range present {
			if scope == prefix {
				return true
			}
		}
	}

	// Otherwise, search for the full scope in the `present` list
	for _, scope := range present {
		if scope == requiredScope {
			return true
		}
	}

	// No scope was found in the `present` list. This request will be denied.
	return false
}
