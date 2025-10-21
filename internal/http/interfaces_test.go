// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockHTTP "github.com/nicholas-fedor/goUpdater/internal/http/mocks"
)

var (
	errInvalidURL  = errors.New("invalid URL")
	errNotFound    = errors.New("404 Not Found")
	errNilResponse = errors.New("nil response")
)

func TestMiddlewareFunc_Wrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    MiddlewareFunc
		next http.RoundTripper
		want http.RoundTripper
	}{
		{
			name: "middleware function returns wrapped round tripper",
			f: func(next http.RoundTripper) http.RoundTripper {
				return &testRoundTripper{wrapped: next}
			},
			next: &testRoundTripper{},
			want: &testRoundTripper{wrapped: &testRoundTripper{}},
		},
		{
			name: "middleware function returns nil",
			f: func(_ http.RoundTripper) http.RoundTripper {
				return nil
			},
			next: &testRoundTripper{},
			want: nil,
		},
		{
			name: "middleware function returns same round tripper",
			f: func(next http.RoundTripper) http.RoundTripper {
				return next
			},
			next: &testRoundTripper{id: "original"},
			want: &testRoundTripper{id: "original"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.f.Wrap(testCase.next)
			require.Equal(t, testCase.want, got)
		})
	}
}

// testRoundTripper is a simple RoundTripper implementation for testing.
type testRoundTripper struct {
	id      string
	wrapped http.RoundTripper
}

func (t *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.wrapped != nil {
		return t.wrapped.RoundTrip(req) //nolint:wrapcheck // test helper
	}

	return &http.Response{StatusCode: http.StatusOK}, nil
}

func TestRequestBuilder_BuildRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		url         string
		body        io.Reader
		mockSetup   func(*mockHTTP.MockRequestBuilder)
		expectedReq *http.Request
		expectedErr error
	}{
		{
			name:   "successful GET request without body",
			method: "GET",
			url:    "https://api.example.com/test",
			body:   nil,
			mockSetup: func(mockRB *mockHTTP.MockRequestBuilder) {
				req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.example.com/test", nil)
				mockRB.EXPECT().
					BuildRequest("GET", "https://api.example.com/test", mock.Anything).
					Return(req, nil).Once()
			},
			expectedReq: &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Scheme: "https", Host: "api.example.com", Path: "/test"},
			},
			expectedErr: nil,
		},
		{
			name:   "successful POST request with body",
			method: "POST",
			url:    "https://api.example.com/data",
			body:   strings.NewReader(`{"key": "value"}`),
			mockSetup: func(mockRB *mockHTTP.MockRequestBuilder) {
				req, _ := http.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"https://api.example.com/data",
					strings.NewReader(`{"key": "value"}`),
				)
				mockRB.EXPECT().
					BuildRequest("POST", "https://api.example.com/data", mock.AnythingOfType("*strings.Reader")).
					Return(req, nil).Once()
			},
			expectedReq: &http.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Scheme: "https", Host: "api.example.com", Path: "/data"},
			},
			expectedErr: nil,
		},
		{
			name:   "request creation failure",
			method: "GET",
			url:    "invalid-url",
			body:   nil,
			mockSetup: func(mockRB *mockHTTP.MockRequestBuilder) {
				mockRB.EXPECT().
					BuildRequest("GET", "invalid-url", mock.Anything).
					Return(nil, errInvalidURL).Once()
			},
			expectedReq: nil,
			expectedErr: errInvalidURL,
		},
		{
			name:   "empty method and URL",
			method: "",
			url:    "",
			body:   nil,
			mockSetup: func(mockRB *mockHTTP.MockRequestBuilder) {
				req, _ := http.NewRequestWithContext(context.Background(), "", "", nil)
				mockRB.EXPECT().
					BuildRequest("", "", mock.Anything).
					Return(req, nil).Once()
			},
			expectedReq: &http.Request{Method: http.MethodGet, URL: &url.URL{}},
			expectedErr: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRB := mockHTTP.NewMockRequestBuilder(t)
			testCase.mockSetup(mockRB)

			gotReq, gotErr := mockRB.BuildRequest(testCase.method, testCase.url, testCase.body)

			if testCase.expectedErr != nil {
				require.Error(t, gotErr)
				require.Contains(t, gotErr.Error(), testCase.expectedErr.Error())
				require.Nil(t, gotReq)
			} else {
				require.NoError(t, gotErr)
				require.NotNil(t, gotReq)
				require.Equal(t, testCase.expectedReq.Method, gotReq.Method)

				if testCase.expectedReq.URL != nil {
					require.Equal(t, testCase.expectedReq.URL.String(), gotReq.URL.String())
				}
			}
		})
	}
}

func TestResponseHandler_HandleResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		response     *http.Response
		mockSetup    func(*mockHTTP.MockResponseHandler)
		expectedData interface{}
		expectedErr  error
	}{
		{
			name:     "successful response handling with JSON data",
			response: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"success": true}`))},
			mockSetup: func(mockRH *mockHTTP.MockResponseHandler) {
				data := map[string]interface{}{"success": true}
				mockRH.EXPECT().HandleResponse(mock.AnythingOfType("*http.Response")).Return(data, nil).Once()
			},
			expectedData: map[string]interface{}{"success": true},
			expectedErr:  nil,
		},
		{
			name:     "successful response handling with string data",
			response: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("OK"))},
			mockSetup: func(mockRH *mockHTTP.MockResponseHandler) {
				mockRH.EXPECT().HandleResponse(mock.AnythingOfType("*http.Response")).Return("OK", nil).Once()
			},
			expectedData: "OK",
			expectedErr:  nil,
		},
		{
			name:     "error response handling",
			response: &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("Not Found"))},
			mockSetup: func(mockRH *mockHTTP.MockResponseHandler) {
				mockRH.EXPECT().
					HandleResponse(mock.AnythingOfType("*http.Response")).
					Return(nil, errNotFound).Once()
			},
			expectedData: nil,
			expectedErr:  errNotFound,
		},
		{
			name:     "nil response handling",
			response: nil,
			mockSetup: func(mockRH *mockHTTP.MockResponseHandler) {
				mockRH.EXPECT().
					HandleResponse((*http.Response)(nil)).
					Return(nil, errNilResponse).Once()
			},
			expectedData: nil,
			expectedErr:  errNilResponse,
		},
		{
			name:     "empty response body",
			response: &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader(""))},
			mockSetup: func(mockRH *mockHTTP.MockResponseHandler) {
				mockRH.EXPECT().HandleResponse(mock.AnythingOfType("*http.Response")).Return("", nil).Once()
			},
			expectedData: "",
			expectedErr:  nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockRH := mockHTTP.NewMockResponseHandler(t)
			testCase.mockSetup(mockRH)

			gotData, gotErr := mockRH.HandleResponse(testCase.response)

			if testCase.expectedErr != nil {
				require.Error(t, gotErr)
				require.Contains(t, gotErr.Error(), testCase.expectedErr.Error())
				require.Nil(t, gotData)
			} else {
				require.NoError(t, gotErr)
				require.Equal(t, testCase.expectedData, gotData)
			}
		})
	}
}

func TestMiddleware_Wrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		next      http.RoundTripper
		mockSetup func(*mockHTTP.MockMiddleware)
		expected  http.RoundTripper
	}{
		{
			name: "middleware wraps round tripper successfully",
			next: &testRoundTripper{id: "original"},
			mockSetup: func(mockMW *mockHTTP.MockMiddleware) {
				wrapped := &testRoundTripper{id: "wrapped"}
				mockMW.EXPECT().
					Wrap(mock.AnythingOfType("*http.testRoundTripper")).
					Return(wrapped).Once()
			},
			expected: &testRoundTripper{id: "wrapped"},
		},
		{
			name: "middleware returns nil",
			next: &testRoundTripper{},
			mockSetup: func(mockMW *mockHTTP.MockMiddleware) {
				mockMW.EXPECT().
					Wrap(mock.AnythingOfType("*http.testRoundTripper")).
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			name: "middleware returns same round tripper",
			next: &testRoundTripper{id: "same"},
			mockSetup: func(mockMW *mockHTTP.MockMiddleware) {
				mockMW.EXPECT().
					Wrap(mock.AnythingOfType("*http.testRoundTripper")).
					Return(&testRoundTripper{id: "same"}).Once()
			},
			expected: &testRoundTripper{id: "same"},
		},
		{
			name: "nil next round tripper",
			next: nil,
			mockSetup: func(mockMW *mockHTTP.MockMiddleware) {
				mockMW.EXPECT().
					Wrap(nil).
					Return(&testRoundTripper{id: "from-nil"}).Once()
			},
			expected: &testRoundTripper{id: "from-nil"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockMW := mockHTTP.NewMockMiddleware(t)
			testCase.mockSetup(mockMW)

			got := mockMW.Wrap(testCase.next)
			require.Equal(t, testCase.expected, got)
		})
	}
}
