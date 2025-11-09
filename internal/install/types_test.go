// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package install

import (
	"context"
	"testing"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/download"
	"github.com/stretchr/testify/assert"
)

func TestNewDownloadServiceImpl(t *testing.T) {
	t.Parallel()

	// Set a short timeout for all tests to prevent hanging
	timeout := 15 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	type args struct {
		downloader *download.Downloader
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Successful construction with nil downloader",
			args: args{
				downloader: nil,
			},
		},
		{
			name: "Successful construction with valid downloader",
			args: args{
				downloader: &download.Downloader{}, // Minimal instance for testing
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			select {
			case <-ctx.Done():
				t.Fatal("Test timed out")
			default:
			}

			got := NewDownloadServiceImpl(testCase.args.downloader)
			assert.NotNil(t, got, "NewDownloadServiceImpl() should return a non-nil instance")
			assert.IsType(t, &DownloadServiceImpl{}, got, "NewDownloadServiceImpl() should return correct type")
		})
	}
}

func TestNewArchiveServiceImpl(t *testing.T) {
	t.Parallel()

	// Set a short timeout for all tests to prevent hanging
	timeout := 15 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	type args struct {
		extractor *archive.Extractor
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Successful construction with nil extractor",
			args: args{
				extractor: nil,
			},
		},
		{
			name: "Successful construction with valid extractor",
			args: args{
				extractor: &archive.Extractor{}, // Minimal instance for testing
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			select {
			case <-ctx.Done():
				t.Fatal("Test timed out")
			default:
			}

			got := NewArchiveServiceImpl(testCase.args.extractor)
			assert.NotNil(t, got, "NewArchiveServiceImpl() should return a non-nil instance")
			assert.IsType(t, &ArchiveServiceImpl{}, got, "NewArchiveServiceImpl() should return correct type")
		})
	}
}

func TestNewDefaultVersionFetcherImpl(t *testing.T) {
	t.Parallel()

	// Set a short timeout for all tests to prevent hanging
	timeout := 15 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	tests := []struct {
		name string
	}{
		{
			name: "Successful construction",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			select {
			case <-ctx.Done():
				t.Fatal("Test timed out")
			default:
			}

			got := NewDefaultVersionFetcherImpl()
			assert.NotNil(t, got, "NewDefaultVersionFetcherImpl() should return a non-nil instance")
			assert.IsType(t, &DefaultVersionFetcherImpl{}, got, "NewDefaultVersionFetcherImpl() should return correct type")
		})
	}
}

func TestNewVersionServiceImpl(t *testing.T) {
	t.Parallel()

	// Set a short timeout for all tests to prevent hanging
	timeout := 15 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	tests := []struct {
		name string
	}{
		{
			name: "Successful construction",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			select {
			case <-ctx.Done():
				t.Fatal("Test timed out")
			default:
			}

			got := NewVersionServiceImpl()
			assert.NotNil(t, got, "NewVersionServiceImpl() should return a non-nil instance")
			assert.IsType(t, &VersionServiceImpl{}, got, "NewVersionServiceImpl() should return correct type")
		})
	}
}
