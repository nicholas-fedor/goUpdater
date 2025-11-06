// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package update

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errNetworkTimeout  = errors.New("network timeout")
	errTestError       = errors.New("test error")
	errErrorWithQuotes = errors.New("error with quotes \"here\"")
	errHashMismatch    = errors.New("hash mismatch")
	errUnderlyingError = errors.New("underlying error")
	errSpecificError   = errors.New("specific error")
	errDifferentError  = errors.New("different error")
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		e    *Error
		want string
	}{
		{
			name: "all fields populated with basic error",
			e: &Error{
				OperationPhase: "download",
				CurrentStep:    "fetching",
				Progress:       "50%",
				Err:            errNetworkTimeout,
			},
			want: "update failed: phase=download step=fetching progress=50%: network timeout",
		},
		{
			name: "nil underlying error",
			e: &Error{
				OperationPhase: "install",
				CurrentStep:    "extracting",
				Progress:       "75%",
				Err:            nil,
			},
			want: "update failed: phase=install step=extracting progress=75%: <nil>",
		},
		{
			name: "empty strings",
			e: &Error{
				OperationPhase: "",
				CurrentStep:    "",
				Progress:       "",
				Err:            errTestError,
			},
			want: "update failed: phase= step= progress=: test error",
		},
		{
			name: "special characters in fields",
			e: &Error{
				OperationPhase: "phase with spaces & symbols",
				CurrentStep:    "step\nwith\tnewlines",
				Progress:       "100% complete!",
				Err:            errErrorWithQuotes,
			},
			want: "update failed: phase=phase with spaces & symbols step=step\nwith\tnewlines " +
				"progress=100% complete!: error with quotes \"here\"",
		},
		{
			name: "long progress string",
			e: &Error{
				OperationPhase: "verification",
				CurrentStep:    "checksum",
				Progress:       "verifying file integrity with SHA256 hash comparison",
				Err:            errHashMismatch,
			},
			want: "update failed: phase=verification step=checksum " +
				"progress=verifying file integrity with SHA256 hash comparison: hash mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.e.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	baseErr := errUnderlyingError
	wrappedErr := &Error{
		OperationPhase: "test",
		CurrentStep:    "unwrap",
		Progress:       "testing",
		Err:            baseErr,
	}

	tests := []struct {
		name     string
		e        *Error
		wantErr  error
		wantBool bool
	}{
		{
			name:     "unwrap with underlying error",
			e:        wrappedErr,
			wantErr:  baseErr,
			wantBool: true,
		},
		{
			name:     "unwrap with nil error",
			e:        &Error{Err: nil},
			wantErr:  nil,
			wantBool: false,
		},
		{
			name: "unwrap empty error",
			e: &Error{
				OperationPhase: "",
				CurrentStep:    "",
				Progress:       "",
				Err:            nil,
			},
			wantErr:  nil,
			wantBool: false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.e.Unwrap()
			assert.Equal(t, testCase.wantErr, err)
			assert.Equal(t, testCase.wantBool, err != nil)
		})
	}
}

func TestError_Unwrap_Compatibility(t *testing.T) {
	t.Parallel()

	baseErr := errSpecificError
	customErr := &customError{msg: "custom error"}

	tests := []struct {
		name        string
		wrappedErr  *Error
		targetErr   error
		shouldMatch bool
	}{
		{
			name: "errors.Is compatibility with basic error",
			wrappedErr: &Error{
				Err: baseErr,
			},
			targetErr:   baseErr,
			shouldMatch: true,
		},
		{
			name: "errors.Is compatibility with custom error type",
			wrappedErr: &Error{
				Err: customErr,
			},
			targetErr:   customErr,
			shouldMatch: true,
		},
		{
			name: "errors.Is no match",
			wrappedErr: &Error{
				Err: baseErr,
			},
			targetErr:   errDifferentError,
			shouldMatch: false,
		},
		{
			name: "errors.As compatibility",
			wrappedErr: &Error{
				Err: customErr,
			},
			targetErr:   &customError{},
			shouldMatch: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.name == "errors.As compatibility" {
				var target *customError

				result := errors.As(testCase.wrappedErr, &target)
				assert.Equal(t, testCase.shouldMatch, result)

				if result {
					assert.Equal(t, customErr.msg, target.msg)
				}
			} else {
				result := errors.Is(testCase.wrappedErr, testCase.targetErr)
				assert.Equal(t, testCase.shouldMatch, result)
			}
		})
	}
}

// customError is a test error type for testing errors.As compatibility.
type customError struct {
	msg string
}

func (c *customError) Error() string {
	return c.msg
}
