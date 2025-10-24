package archive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	// Note: This function uses real filesystem, so cannot be tested in unit tests.
	// Testing is covered by Extractor.Validate tests.
	t.Skip("Cannot test Validate function without filesystem operations")
}

func TestExtract(t *testing.T) {
	t.Parallel()

	// Note: This function uses real filesystem, so cannot be tested in unit tests.
	// Testing is covered by Extractor.Extract tests.
	t.Skip("Cannot test Extract function without filesystem operations")
}

func TestExtractVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "valid version with linux-amd64",
			filename: "go1.21.0.linux-amd64.tar.gz",
			want:     "go1.21.0",
		},
		{
			name:     "valid version with darwin-amd64",
			filename: "/path/to/go1.20.0.darwin-amd64.tar.gz",
			want:     "go1.20.0",
		},
		{
			name:     "valid version with suffix",
			filename: "go1.25.2.linux-amd64.tar.gz",
			want:     "go1.25.2",
		},
		{
			name:     "no go prefix",
			filename: "invalid-filename",
			want:     "invalid-filename",
		},
		{
			name:     "empty after go",
			filename: "go",
			want:     "go",
		},
		{
			name:     "no digit after go",
			filename: "goxxx",
			want:     "goxxx",
		},
		{
			name:     "valid semver with high numbers",
			filename: "go1.99.99.linux-amd64.tar.gz",
			want:     "go1.99.99",
		},
		{
			name:     "no extension",
			filename: "go1.21.0.linux-amd64",
			want:     "go1.21.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ExtractVersion(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidatePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		targetPath string
		installDir string
		wantErr    bool
	}{
		{
			name:       "valid path within install dir",
			targetPath: "/install/go/bin/go",
			installDir: "/install",
			wantErr:    false,
		},
		{
			name:       "path equals install dir",
			targetPath: "/install",
			installDir: "/install",
			wantErr:    false,
		},
		{
			name:       "path traversal attempt",
			targetPath: "/install/../outside/go",
			installDir: "/install",
			wantErr:    true,
		},
		{
			name:       "absolute path outside",
			targetPath: "/outside/go",
			installDir: "/install",
			wantErr:    true,
		},
		{
			name:       "relative path with ..",
			targetPath: "/install/go/../../../outside",
			installDir: "/install",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePath(tt.targetPath, tt.installDir)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_validateHeaderName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		headerName string
		wantErr    bool
	}{
		{
			name:       "valid header name",
			headerName: "go/bin/go",
			wantErr:    false,
		},
		{
			name:       "absolute path",
			headerName: "/absolute/path",
			wantErr:    true,
		},
		{
			name:       "parent directory",
			headerName: "../escape",
			wantErr:    true,
		},
		{
			name:       "backslash",
			headerName: "go\\bin\\go",
			wantErr:    true,
		},
		{
			name:       "null byte",
			headerName: "go\x00bin",
			wantErr:    true,
		},
		{
			name:       "valid with subdirs",
			headerName: "go/src/cmd/go/main.go",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateHeaderName(tt.headerName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
