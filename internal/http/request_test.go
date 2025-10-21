// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		method    string
		url       string
		body      io.Reader
		headers   map[string]string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid GET request",
			method:  http.MethodGet,
			url:     "https://example.com",
			body:    nil,
			headers: map[string]string{"User-Agent": "test"},
			wantErr: false,
		},
		{
			name:    "valid POST request with body",
			method:  http.MethodPost,
			url:     "https://example.com",
			body:    strings.NewReader("test body"),
			headers: map[string]string{"Content-Type": "application/json"},
			wantErr: false,
		},
		{
			name:      "invalid URL",
			method:    http.MethodGet,
			url:       "invalid-url",
			body:      nil,
			headers:   nil,
			wantErr:   true,
			errString: "invalid URI",
		},
		{
			name:    "empty headers",
			method:  http.MethodGet,
			url:     "https://example.com",
			body:    nil,
			headers: nil,
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewRequest(context.Background(), testCase.method, testCase.url, testCase.body, testCase.headers)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errString != "" {
					assert.Contains(t, err.Error(), testCase.errString)
				}

				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, testCase.method, got.Method)
				assert.Equal(t, testCase.url, got.URL.String())

				for key, value := range testCase.headers {
					assert.Equal(t, value, got.Header.Get(key))
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		headers   map[string]string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid GET request",
			url:     "https://example.com",
			headers: map[string]string{"Accept": "application/json"},
			wantErr: false,
		},
		{
			name:      "invalid URL",
			url:       "invalid-url",
			headers:   nil,
			wantErr:   true,
			errString: "invalid URI",
		},
		{
			name:    "empty headers",
			url:     "https://example.com",
			headers: nil,
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := Get(context.Background(), testCase.url, testCase.headers)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errString != "" {
					assert.Contains(t, err.Error(), testCase.errString)
				}

				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, http.MethodGet, got.Method)
				assert.Equal(t, testCase.url, got.URL.String())

				for key, value := range testCase.headers {
					assert.Equal(t, value, got.Header.Get(key))
				}
			}
		})
	}
}

func TestPost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		body      io.Reader
		headers   map[string]string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid POST request with body",
			url:     "https://example.com",
			body:    strings.NewReader("test data"),
			headers: map[string]string{"Content-Type": "application/json"},
			wantErr: false,
		},
		{
			name:    "valid POST request without body",
			url:     "https://example.com",
			body:    nil,
			headers: map[string]string{"Accept": "application/json"},
			wantErr: false,
		},
		{
			name:      "invalid URL",
			url:       "invalid-url",
			body:      strings.NewReader("data"),
			headers:   nil,
			wantErr:   true,
			errString: "invalid URI",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := Post(context.Background(), testCase.url, testCase.body, testCase.headers)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errString != "" {
					assert.Contains(t, err.Error(), testCase.errString)
				}

				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, http.MethodPost, got.Method)
				assert.Equal(t, testCase.url, got.URL.String())

				for key, value := range testCase.headers {
					assert.Equal(t, value, got.Header.Get(key))
				}
			}
		})
	}
}

func TestPut(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		body      io.Reader
		headers   map[string]string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid PUT request with body",
			url:     "https://example.com",
			body:    strings.NewReader("updated data"),
			headers: map[string]string{"Content-Type": "application/json"},
			wantErr: false,
		},
		{
			name:    "valid PUT request without body",
			url:     "https://example.com",
			body:    nil,
			headers: map[string]string{"Accept": "application/json"},
			wantErr: false,
		},
		{
			name:      "invalid URL",
			url:       "invalid-url",
			body:      strings.NewReader("data"),
			headers:   nil,
			wantErr:   true,
			errString: "invalid URI",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := Put(context.Background(), testCase.url, testCase.body, testCase.headers)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errString != "" {
					assert.Contains(t, err.Error(), testCase.errString)
				}

				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, http.MethodPut, got.Method)
				assert.Equal(t, testCase.url, got.URL.String())

				for key, value := range testCase.headers {
					assert.Equal(t, value, got.Header.Get(key))
				}
			}
		})
	}
}

func TestPatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		body      io.Reader
		headers   map[string]string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid PATCH request with body",
			url:     "https://example.com",
			body:    strings.NewReader("patch data"),
			headers: map[string]string{"Content-Type": "application/json"},
			wantErr: false,
		},
		{
			name:    "valid PATCH request without body",
			url:     "https://example.com",
			body:    nil,
			headers: map[string]string{"Accept": "application/json"},
			wantErr: false,
		},
		{
			name:      "invalid URL",
			url:       "invalid-url",
			body:      strings.NewReader("data"),
			headers:   nil,
			wantErr:   true,
			errString: "invalid URI",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := Patch(context.Background(), testCase.url, testCase.body, testCase.headers)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errString != "" {
					assert.Contains(t, err.Error(), testCase.errString)
				}

				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, http.MethodPatch, got.Method)
				assert.Equal(t, testCase.url, got.URL.String())

				for key, value := range testCase.headers {
					assert.Equal(t, value, got.Header.Get(key))
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		headers   map[string]string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid DELETE request",
			url:     "https://example.com",
			headers: map[string]string{"Accept": "application/json"},
			wantErr: false,
		},
		{
			name:    "valid DELETE request without headers",
			url:     "https://example.com",
			headers: nil,
			wantErr: false,
		},
		{
			name:      "invalid URL",
			url:       "invalid-url",
			headers:   nil,
			wantErr:   true,
			errString: "invalid URI",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := Delete(context.Background(), testCase.url, testCase.headers)

			if testCase.wantErr {
				require.Error(t, err)

				if testCase.errString != "" {
					assert.Contains(t, err.Error(), testCase.errString)
				}

				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, http.MethodDelete, got.Method)
				assert.Equal(t, testCase.url, got.URL.String())

				for key, value := range testCase.headers {
					assert.Equal(t, value, got.Header.Get(key))
				}
			}
		})
	}
}
