package tootecho

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/benpate/toot"
	"github.com/benpate/toot/object"
	"github.com/benpate/toot/txn"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// newMultipartUpload builds a real multipart/form-data POST body carrying one file field
// ("file") and one plain text field ("description"), the same shape a media upload sends.
func newMultipartUpload(t *testing.T, fieldName string, filename string, content string, extraFields map[string]string) (*bytes.Buffer, string) {

	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)

	for key, value := range extraFields {
		require.NoError(t, writer.WriteField(key, value))
	}

	require.NoError(t, writer.Close())

	return body, writer.FormDataContentType()
}

// TestBindBody_MultipartFileReachesTheHandler confirms the whole path end-to-end: a real
// multipart POST, through Echo's router and this library's binder, into a registered
// PostMedia handler whose input has a *multipart.FileHeader field.
func TestBindBody_MultipartFileReachesTheHandler(t *testing.T) {

	var receivedContent string
	var receivedFilename string
	var receivedDescription string

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	api.PostMedia = func(_ testToken, input txn.PostMedia) (object.MediaAttachment, error) {

		require.NotNil(t, input.File)
		receivedFilename = input.File.Filename
		receivedDescription = input.Description

		file, err := input.File.Open()
		require.NoError(t, err)
		defer file.Close()

		content, err := io.ReadAll(file)
		require.NoError(t, err)
		receivedContent = string(content)

		return object.MediaAttachment{ID: "test-id"}, nil
	}

	e := echo.New()
	Register(e, api)

	body, contentType := newMultipartUpload(t, "file", "photo.jpg", "fake image bytes", map[string]string{
		"description": "A test photo",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v2/media", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	req.Header.Set(echo.HeaderAuthorization, "Bearer test")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "photo.jpg", receivedFilename)
	require.Equal(t, "fake image bytes", receivedContent)
	require.Equal(t, "A test photo", receivedDescription)
}

// TestBindBody_EmptyBracketArrayReachesTheHandler replays the exact request shape the official
// Mastodon client sends for an array field: "media_ids[]=<id>" once per attachment, never an
// indexed key.
func TestBindBody_EmptyBracketArrayReachesTheHandler(t *testing.T) {

	var receivedStatus string
	var receivedMediaIDs []string

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	api.PostStatus = func(_ testToken, input txn.PostStatus) (object.Status, error) {
		receivedStatus = input.Status
		receivedMediaIDs = input.MediaIDs
		return object.Status{ID: "test-id"}, nil
	}

	e := echo.New()
	Register(e, api)

	form := url.Values{}
	form.Set("status", "a post with two photos")
	form.Add("media_ids[]", "111")
	form.Add("media_ids[]", "222")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/statuses", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set(echo.HeaderAuthorization, "Bearer test")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "a post with two photos", receivedStatus)
	require.Equal(t, []string{"111", "222"}, receivedMediaIDs)
}

// TestExpandEmptyBracketArrays covers the rewrite in isolation: the array convention real
// clients use, an already-indexed key (left untouched, so an existing caller keeps working),
// a mix of both in one request, and a plain non-array field.
func TestExpandEmptyBracketArrays(t *testing.T) {

	in := url.Values{
		"media_ids[]":     {"111", "222"},
		"account_ids[0]":  {"a"},
		"description":     {"hello"},
		"already_indexed": {"x", "y"}, // no brackets at all -- must pass through as-is
	}

	out := expandEmptyBracketArrays(in)

	require.Equal(t, []string{"111"}, out["media_ids[0]"])
	require.Equal(t, []string{"222"}, out["media_ids[1]"])
	require.NotContains(t, out, "media_ids[]")

	require.Equal(t, []string{"a"}, out["account_ids[0]"])
	require.Equal(t, []string{"hello"}, out["description"])
	require.Equal(t, []string{"x", "y"}, out["already_indexed"])
}

// TestBindMultipartFiles_LeavesUnrelatedOrMissingFieldsAlone confirms the helper only ever
// touches a *multipart.FileHeader field with a matching upload, and is a no-op otherwise --
// no upload at all, a field with no form tag, and a tag nothing was uploaded under.
func TestBindMultipartFiles_LeavesUnrelatedOrMissingFieldsAlone(t *testing.T) {

	type input struct {
		File        *multipart.FileHeader `form:"file"`
		Thumbnail   *multipart.FileHeader `form:"thumbnail"`
		Description string                `form:"description"`
		untagged    *multipart.FileHeader //nolint:unused // exercises "no form tag" below
	}

	// No files uploaded at all
	empty := input{Description: "kept"}
	require.NoError(t, bindMultipartFiles(&empty, nil))
	require.Nil(t, empty.File)
	require.Equal(t, "kept", empty.Description)

	// A file under a tag this struct doesn't have
	one := input{}
	require.NoError(t, bindMultipartFiles(&one, map[string][]*multipart.FileHeader{
		"avatar": {{Filename: "x.png"}},
	}))
	require.Nil(t, one.File)
	require.Nil(t, one.Thumbnail)

	// Only the field whose tag matches gets set; a second declared field is untouched
	two := input{}
	require.NoError(t, bindMultipartFiles(&two, map[string][]*multipart.FileHeader{
		"file": {{Filename: "photo.jpg"}},
	}))
	require.NotNil(t, two.File)
	require.Equal(t, "photo.jpg", two.File.Filename)
	require.Nil(t, two.Thumbnail)

	// Not a pointer, and not a struct: both must be quietly ignored, not panic
	require.NoError(t, bindMultipartFiles(input{}, map[string][]*multipart.FileHeader{"file": {{}}}))
	notAStruct := "hello"
	require.NoError(t, bindMultipartFiles(&notAStruct, map[string][]*multipart.FileHeader{"file": {{}}}))
}

// TestBindBody_UnsentFieldsStayNil replays the profile edit the iOS app sends -- display_name, an empty
// note, and bot=false, with no discoverable -- and confirms the omitted field decodes as nil.
func TestBindBody_UnsentFieldsStayNil(t *testing.T) {

	var received txn.PatchAccount_UpdateCredentials

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	api.PatchAccount_UpdateCredentials = func(_ testToken, input txn.PatchAccount_UpdateCredentials) (object.Account, error) {
		received = input
		return object.Account{ID: "test-id"}, nil
	}

	e := echo.New()
	Register(e, api)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.WriteField("note", ""))
	require.NoError(t, writer.WriteField("bot", "false"))
	require.NoError(t, writer.WriteField("display_name", "Test Demo"))
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/update_credentials", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	req.Header.Set(echo.HeaderAuthorization, "Bearer test")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	require.NotNil(t, received.DisplayName)
	require.Equal(t, "Test Demo", *received.DisplayName)

	require.NotNil(t, received.Note, "a note that was sent empty is still sent")
	require.Equal(t, "", *received.Note)

	require.NotNil(t, received.Bot)
	require.False(t, *received.Bot)

	require.Nil(t, received.Discoverable, "a field the client left out must stay nil")
	require.Nil(t, received.Locked)
}

// TestBindBody_ProfileImagesReachTheHandler uploads an avatar and a header the way the iOS app does
// and confirms both files reach the handler, alongside a text field.
func TestBindBody_ProfileImagesReachTheHandler(t *testing.T) {

	var received txn.PatchAccount_UpdateCredentials
	var avatarContent string

	api := toot.New(func(*http.Request) (testToken, error) {
		return testToken{}, nil
	})

	api.PatchAccount_UpdateCredentials = func(_ testToken, input txn.PatchAccount_UpdateCredentials) (object.Account, error) {
		received = input

		if input.Avatar != nil {
			file, err := input.Avatar.Open()
			require.NoError(t, err)
			defer func() { require.NoError(t, file.Close()) }()

			content, err := io.ReadAll(file)
			require.NoError(t, err)
			avatarContent = string(content)
		}

		return object.Account{ID: "test-id"}, nil
	}

	e := echo.New()
	Register(e, api)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.WriteField("display_name", "With Pictures"))

	avatar, err := writer.CreateFormFile("avatar", "me.png")
	require.NoError(t, err)
	_, err = avatar.Write([]byte("avatar bytes"))
	require.NoError(t, err)

	header, err := writer.CreateFormFile("header", "banner.jpg")
	require.NoError(t, err)
	_, err = header.Write([]byte("header bytes"))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/update_credentials", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	req.Header.Set(echo.HeaderAuthorization, "Bearer test")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "With Pictures", *received.DisplayName)
	require.NotNil(t, received.Avatar)
	require.Equal(t, "me.png", received.Avatar.Filename)
	require.Equal(t, "avatar bytes", avatarContent)
	require.NotNil(t, received.Header)
	require.Equal(t, "banner.jpg", received.Header.Filename)
}
