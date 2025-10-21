// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		resp *http.Response
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid response with 200 status",
			args: args{
				resp: &http.Response{StatusCode: http.StatusOK},
			},
			wantErr: false,
		},
		{
			name: "valid response with 201 status",
			args: args{
				resp: &http.Response{StatusCode: http.StatusCreated},
			},
			wantErr: false,
		},
		{
			name: "invalid response with 400 status",
			args: args{
				resp: &http.Response{StatusCode: http.StatusBadRequest},
			},
			wantErr: true,
		},
		{
			name: "invalid response with 500 status",
			args: args{
				resp: &http.Response{StatusCode: http.StatusInternalServerError},
			},
			wantErr: true,
		},
		{
			name: "nil response",
			args: args{
				resp: nil,
			},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateResponse(testCase.args.resp)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckStatusCode(t *testing.T) {
	t.Parallel()

	type args struct {
		resp     *http.Response
		expected int
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "status code matches expected",
			args: args{
				resp:     &http.Response{StatusCode: http.StatusOK},
				expected: 200,
			},
			wantErr: false,
		},
		{
			name: "status code does not match expected",
			args: args{
				resp:     &http.Response{StatusCode: http.StatusNotFound},
				expected: 200,
			},
			wantErr: true,
		},
		{
			name: "check different status codes",
			args: args{
				resp:     &http.Response{StatusCode: http.StatusCreated},
				expected: 201,
			},
			wantErr: false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := CheckStatusCode(testCase.args.resp, testCase.args.expected)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReadBody(t *testing.T) {
	t.Parallel()

	type args struct {
		resp *http.Response
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "successfully reads response body",
			args: args{
				resp: &http.Response{
					Body: io.NopCloser(strings.NewReader("test data")),
				},
			},
			want:    []byte("test data"),
			wantErr: false,
		},
		{
			name: "handles empty body",
			args: args{
				resp: &http.Response{
					Body: io.NopCloser(strings.NewReader("")),
				},
			},
			want:    []byte(""),
			wantErr: false,
		},
		{
			name: "handles read error",
			args: args{
				resp: &http.Response{
					Body: &errorReader{err: io.ErrUnexpectedEOF},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReadBody(testCase.args.resp)
			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

// errorReader is a helper for testing read errors.
type errorReader struct {
	err error
}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, e.err
}

func (e *errorReader) Close() error {
	return nil
}

func TestParseJSON(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		resp    *http.Response
		v       interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "successfully parses valid JSON",
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"name":"test","value":42}`)),
			},
			v:       &testStruct{},
			want:    &testStruct{Name: "test", Value: 42},
			wantErr: false,
		},
		{
			name: "handles invalid JSON",
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`invalid json`)),
			},
			v:       &testStruct{},
			want:    &testStruct{},
			wantErr: true,
		},
		{
			name: "handles read error",
			resp: &http.Response{
				Body: &errorReader{err: io.ErrUnexpectedEOF},
			},
			v:       &testStruct{},
			want:    &testStruct{},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ParseJSON(testCase.resp, testCase.v)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, testCase.v)
			}
		})
	}
}

func TestParseXML(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name  string `xml:"name"`
		Value int    `xml:"value"`
	}

	tests := []struct {
		name    string
		resp    *http.Response
		v       interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "successfully parses valid XML",
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`<testStruct><name>test</name><value>42</value></testStruct>`)),
			},
			v:       &testStruct{},
			want:    &testStruct{Name: "test", Value: 42},
			wantErr: false,
		},
		{
			name: "handles invalid XML",
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`invalid xml`)),
			},
			v:       &testStruct{},
			want:    &testStruct{},
			wantErr: true,
		},
		{
			name: "handles read error",
			resp: &http.Response{
				Body: &errorReader{err: io.ErrUnexpectedEOF},
			},
			v:       &testStruct{},
			want:    &testStruct{},
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ParseXML(testCase.resp, testCase.v)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want, testCase.v)
			}
		})
	}
}
