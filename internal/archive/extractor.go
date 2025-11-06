// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

const (
	defaultDirPerm    = 0755 // Default directory permissions
	defaultFilePerm   = 0644 // Default file permissions
	unixPermMask      = 0777 // Unix permission mask for tar headers
	chunkedThreshold  = 2    // multiplier for bufferSize to decide chunked extraction
	errChanMultiplier = 2    // multiplier for error channel buffer size
)

// fileExtractionWork represents work for concurrent file extraction.
// For chunked extraction, multiple work items are sent for the same file,
// with eof set to true on the last chunk.
type fileExtractionWork struct {
	targetPath string
	data       []byte
	mode       os.FileMode
	eof        bool // indicates if this is the last chunk for the file
}

// extractFileWorker processes file extraction work from the work channel.
// It handles both single-chunk (traditional) and multi-chunk (large file) extractions.
//
//nolint:cyclop // necessary for proper error handling and resource cleanup
func (e *Extractor) extractFileWorker(
	ctx context.Context,
	workChan <-chan fileExtractionWork,
	errChan chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	logger.Debug("Worker started")

	openFiles := make(map[string]io.WriteCloser)

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Worker cancelled via context")
			// Close any open files
			for path, file := range openFiles {
				_ = file.Close()

				delete(openFiles, path)
			}

			return
		case work, ok := <-workChan:
			if !ok {
				logger.Debug("Worker finished: work channel closed")
				// Close any remaining open files
				for path, file := range openFiles {
					_ = file.Close()

					delete(openFiles, path)
				}

				return
			}

			logger.Debugf("Worker processing file chunk: %s, eof: %v", work.targetPath, work.eof)

			err := e.extractFileChunk(work.targetPath, work.data, work.mode, work.eof, openFiles)
			if err != nil {
				logger.Debugf("Worker error for %s: %v", work.targetPath, err)

				// Close all open files on error
				for path, file := range openFiles {
					_ = file.Close()

					delete(openFiles, path)
				}

				select {
				case errChan <- err:
				case <-ctx.Done():
				}

				return
			}

			logger.Debugf("Worker completed file chunk: %s", work.targetPath)
		}
	}
}

// extractFileChunk extracts a chunk of file data, handling both single and multi-chunk files.
// For the first chunk, it opens the file. For subsequent chunks, it writes to the open file.
// On EOF, it closes the file and sets permissions.
//
//nolint:funlen // complex file chunk handling with error management
func (e *Extractor) extractFileChunk(
	targetPath string,
	data []byte,
	mode os.FileMode,
	eof bool,
	openFiles map[string]io.WriteCloser,
) error {
	targetPath = filepath.Clean(targetPath)
	logger.Debugf("Processing file chunk for: %s, size: %d, eof: %v", targetPath, len(data), eof)

	file, exists := openFiles[targetPath]
	if !exists {
		// First chunk: create parent directory and open file
		logger.Debugf("Opening file for first chunk: %s", targetPath)

		err := e.fs.MkdirAll(filepath.Dir(targetPath), defaultDirPerm) // #nosec G301
		if err != nil {
			logger.Debugf("MkdirAll failed for %s: %v", targetPath, err)

			return fmt.Errorf("mkdirall failed: %w", err)
		}

		file, err = e.fs.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultFilePerm) // #nosec G302
		if err != nil {
			logger.Debugf("OpenFile failed for %s: %v", targetPath, err)

			return fmt.Errorf("open file failed: %w", err)
		}

		openFiles[targetPath] = file
		logger.Debugf("Opened file: %s", targetPath)
	}

	// Write chunk data
	logger.Debugf("Writing chunk to file: %s", targetPath)

	_, err := file.Write(data)
	if err != nil {
		logger.Debugf("Write failed for %s: %v", targetPath, err)

		_ = file.Close()

		delete(openFiles, targetPath)

		return fmt.Errorf("write failed: %w", err)
	}

	if eof {
		// Last chunk: close file and set permissions
		logger.Debugf("Closing file after last chunk: %s", targetPath)

		err = file.Close()

		delete(openFiles, targetPath)

		if err != nil {
			logger.Debugf("Close failed for %s: %v", targetPath, err)

			return fmt.Errorf("close failed: %w", err)
		}

		// Set correct permissions with timeout to prevent hanging
		logger.Debugf("Setting permissions for file: %s", targetPath)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second, //nolint:mnd // 5 second timeout for chmod operation
		)
		defer cancel()

		done := make(chan error, 1)

		go func() {
			done <- e.fs.Chmod(targetPath, mode)
		}()

		select {
		case err := <-done:
			if err != nil {
				logger.Debugf("Chmod failed for %s: %v", targetPath, err)

				return fmt.Errorf("chmod failed: %w", err)
			}

			logger.Debugf("Set permissions for file: %s", targetPath)
		case <-ctx.Done():
			logger.Warnf("Chmod timed out for file: %s", targetPath)
		}

		logger.Debugf("File extraction completed for: %s", targetPath)
	}

	return nil
}

// NewExtractor creates a new Extractor with the given dependencies.
func NewExtractor(fs filesystem.FileSystem, processor Processor) *Extractor {
	numWorkers := runtime.NumCPU()

	return &Extractor{
		fs:           fs,
		processor:    processor,
		maxFiles:     20000,             //nolint:mnd // 20000 is the standard limit for Go archives
		maxTotalSize: 500 * 1024 * 1024, //nolint:mnd // 500MB limit (Go archives are ~196MB)
		maxFileSize:  50 * 1024 * 1024,  //nolint:mnd // 50MB per file limit (largest Go file is ~20.6MB)
		bufferSize:   32 * 1024 * 1024,  //nolint:mnd // 32MB buffer size for improved I/O performance
		numWorkers:   numWorkers,        // number of concurrent workers based on CPU cores
	}
}

// Extract extracts the tar.gz archive to the specified destination directory.
// It validates paths to prevent directory traversal attacks and limits the number of files.
// The file count limit is set to 20,000 to accommodate legitimate Go archives while preventing zip bomb attacks.
//
//nolint:cyclop,funlen // complex archive extraction with security validations
func (e *Extractor) Extract(archivePath, destDir string) error {
	extractionStart := time.Now()
	archivePath = filepath.Clean(archivePath)
	destDir = filepath.Clean(destDir)

	logger.Debugf("Extract archive: %s", archivePath)
	logger.Debugf("Extract destination: %s", destDir)

	// Set up concurrent file extraction
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := e.Validate(archivePath, destDir)
	if err != nil {
		cancel()

		return err
	}

	file, err := e.fs.Open(archivePath)
	if err != nil {
		cancel()

		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "opening archive file",
			Err:         err,
		}
	}

	defer func() { _ = file.Close() }()

	gzipStart := time.Now()

	gzipReader, err := e.processor.NewGzipReader(file)
	if err != nil {
		cancel()

		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "creating gzip reader",
			Err:         err,
		}
	}

	logger.Debugf("Gzip reader creation took: %v", time.Since(gzipStart))

	defer func() { _ = gzipReader.Close() }()

	tarReader := e.processor.NewTarReader(gzipReader)

	fileCount := 0
	totalSize := int64(0)

	defer func() {
		logger.Debugf("Total extraction time: %v for %d files (%d bytes)", time.Since(extractionStart), fileCount, totalSize)
	}()

	workChan := make(
		chan fileExtractionWork,
		e.numWorkers*2, //nolint:mnd // buffer size is numWorkers * 2 for optimal throughput
	)
	errChan := make(chan error, e.numWorkers*errChanMultiplier) // buffer size to prevent blocking on multiple errors

	var waitGroup sync.WaitGroup

	defer close(workChan)
	defer waitGroup.Wait()

	// Start workers
	for range e.numWorkers {
		waitGroup.Add(1)

		go e.extractFileWorker(ctx, workChan, errChan, &waitGroup)
	}

	loopStart := time.Now()

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			cancel()

			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "reading tar header",
				Err:         err,
			}
		}

		fileCount++
		if fileCount > e.maxFiles {
			cancel()

			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "validating file count",
				Err:         fmt.Errorf("archive contains too many files: %w", ErrTooManyFiles),
			}
		}

		// Check for zip bomb: extremely large files or excessive total size
		if header.Size > e.maxFileSize {
			cancel()

			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "validating file size",
				Err: fmt.Errorf("archive contains file too large: %s (%d bytes): %w",
					header.Name, header.Size, ErrFileTooLarge),
			}
		}

		totalSize += header.Size
		if totalSize > e.maxTotalSize {
			cancel()

			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "validating total size",
				Err:         fmt.Errorf("archive total size too large: %d bytes: %w", totalSize, ErrFileTooLarge),
			}
		}

		err = e.processTarEntryConcurrent(ctx, tarReader, header, destDir, archivePath, workChan)
		if err != nil {
			cancel() // Cancel workers

			return err
		}
	}

	logger.Debugf("Tar processing loop took: %v", time.Since(loopStart))

	logger.Debug("Starting worker synchronization")
	// Wait for workers to finish
	waitGroup.Wait()
	logger.Debug("Worker synchronization completed")

	// Check for any worker errors
	select {
	case workerErr := <-errChan:
		return workerErr
	default:
		// No errors
	}

	logger.Debugf("Successfully extracted archive: %s", archivePath)

	return nil
}

// Validate checks if the archive file exists and is a regular file.
// It returns an error if the archive path does not exist or is not a regular file.
func (e *Extractor) Validate(archivePath, destDir string) error {
	logger.Debugf("Extractor.Validate: validating archive: %s", archivePath)
	logger.Debugf("Extractor.Validate: destination: %s", destDir)

	info, err := e.fs.Stat(archivePath)
	if err != nil {
		logger.Debugf("Extractor.Validate: Stat failed for %s: %v", archivePath, err)

		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating archive",
			Err: &ValidationError{
				FilePath: archivePath,
				Criteria: "file existence",
				Err:      err,
			},
		}
	}

	if !info.Mode().IsRegular() {
		logger.Debugf("Extractor.Validate: %s is not a regular file", archivePath)

		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating archive",
			Err: &ValidationError{
				FilePath: archivePath,
				Criteria: "regular file type",
				Err:      ErrArchiveNotRegular,
			},
		}
	}

	logger.Debugf("Extractor.Validate: validation successful for %s", archivePath)

	return nil
}

// validateAndConstructTargetPath validates the header name and constructs a safe target path.
// It performs multiple layers of path validation to prevent path traversal attacks.
func (e *Extractor) validateAndConstructTargetPath(header *tar.Header, destDir, archivePath string) (string, error) {
	// Validate the header name
	err := validateHeaderName(header.Name)
	if err != nil {
		return "", &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating header name",
			Err:         err,
		}
	}

	// Construct target path safely
	cleanDestDir := filepath.Clean(destDir)
	// gosec G305 is triggered by filepath.Join, but we have validated the path thoroughly above
	// The path is safe because:
	// 1. header.Name is validated to not contain .. or be absolute
	// 2. targetPath is checked to be within cleanDestDir
	// 3. ValidatePath ensures no traversal
	targetPath := cleanDestDir + string(filepath.Separator) + header.Name
	targetPath = filepath.Clean(targetPath)

	// Validate that the target path is within the destination directory
	if !strings.HasPrefix(targetPath, cleanDestDir+string(filepath.Separator)) && targetPath != cleanDestDir {
		return "", &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating target path within destination",
			Err:         fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath),
		}
	}

	// Additional validation to prevent path traversal
	rel, err := filepath.Rel(cleanDestDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating relative path",
			Err:         fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath),
		}
	}

	// Ensure the target path is safe by checking it doesn't escape the destination directory
	if !strings.HasPrefix(targetPath, cleanDestDir) {
		return "", &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating target path safety",
			Err:         fmt.Errorf("invalid file path in archive: %s: %w", targetPath, ErrInvalidPath),
		}
	}

	// Final safety check: ensure the path is validated before use
	err = ValidatePath(targetPath, cleanDestDir)
	if err != nil {
		return "", &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "final path validation",
			Err:         err,
		}
	}

	return targetPath, nil
}

// processTarEntryConcurrent processes a single tar entry with concurrent file extraction support.
// It performs path validation and delegates to extractEntryConcurrent for actual extraction.
// Directories, symlinks, and hard links are processed sequentially, while regular files are processed concurrently.
func (e *Extractor) processTarEntryConcurrent(
	ctx context.Context,
	tarReader TarReader,
	header *tar.Header,
	destDir, archivePath string,
	workChan chan<- fileExtractionWork,
) error {
	logger.Debugf("Processing tar entry: %s", header.Name)

	targetPath, err := e.validateAndConstructTargetPath(header, destDir, archivePath)
	if err != nil {
		return err
	}

	// gosec G305 is triggered by filepath.Join, but we have validated the path thoroughly above
	// The path is safe because:
	// 1. header.Name is validated to not contain .. or be absolute
	// 2. targetPath is checked to be within cleanDestDir
	// 3. ValidatePath ensures no traversal
	// 4. Symlinks in the path are resolved and validated to stay within cleanDestDir
	err = e.extractEntryConcurrent(ctx, tarReader, header, targetPath, destDir, destDir, archivePath, workChan)
	if err != nil {
		return err
	}

	return nil
}

// extractDirectory creates a directory with the specified permissions.
func (e *Extractor) extractDirectory(targetPath string, mode os.FileMode) error {
	// Create directory permissively, then set correct permissions
	err := e.fs.MkdirAll(targetPath, defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("mkdirall failed: %w", err)
	}

	err = e.fs.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	return nil
}

// extractRegularFile extracts a regular file from the tar reader using buffered I/O for better performance.
func (e *Extractor) extractRegularFile(tarReader TarReader, targetPath string, mode os.FileMode, size int64) error {
	targetPath = filepath.Clean(targetPath)

	// Ensure parent directory exists
	err := e.fs.MkdirAll(filepath.Dir(targetPath), defaultDirPerm) // #nosec G301
	if err != nil {
		return fmt.Errorf("mkdirall failed: %w", err)
	}

	// Create file permissively, then set correct permissions
	file, err := e.fs.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultFilePerm) // #nosec G302
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}

	// Use buffered copy with configurable buffer for better I/O performance
	buffer := make([]byte, e.bufferSize)

	copyStart := time.Now()
	_, err = io.CopyBuffer(file, tarReader, buffer)
	copyDuration := time.Since(copyStart)
	logger.Debugf("File copy took: %v for file size %d bytes", copyDuration, size)

	if err != nil {
		_ = file.Close()

		return fmt.Errorf("copy buffer failed: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	// Set correct permissions
	err = e.fs.Chmod(targetPath, mode)
	if err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	return nil
}

// extractSymlink creates a symlink after validating the linkname.
func (e *Extractor) extractSymlink(targetPath, linkname, baseDir, destDir, archivePath string) error {
	// Validate the linkname to prevent symlink attacks
	err := e.validateLinkname(linkname, baseDir, destDir)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating symlink target",
			Err:         err,
		}
	}

	// Create symlink
	err = e.fs.Symlink(linkname, targetPath)
	if err != nil {
		return fmt.Errorf("symlink failed: %w", err)
	}

	return nil
}

// extractHardLink creates a hard link after validating the linkname.
func (e *Extractor) extractHardLink(targetPath, linkname, baseDir, destDir, archivePath string) error {
	// Validate the linkname to prevent hard link attacks
	err := e.validateLinkname(linkname, baseDir, destDir)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "validating linkname for hard link",
			Err:         err,
		}
	}

	// Create hard link
	err = e.fs.Link(linkname, targetPath)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "extracting hard link",
			Err:         err,
		}
	}

	return nil
}

// extractEntryConcurrent extracts a single entry from the tar archive with concurrent file extraction support.
// It handles directories, regular files, symlinks, and hard links, preserving permissions from the tar header.
// Files and directories are created permissively then chmod to the correct permissions from header.Mode & 0777.
// Regular files are processed concurrently, while other entry types are processed sequentially.
func (e *Extractor) extractEntryConcurrent(
	ctx context.Context,
	tarReader TarReader,
	header *tar.Header,
	targetPath, baseDir, destDir, archivePath string,
	workChan chan<- fileExtractionWork,
) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return e.extractDirectoryConcurrent(header, targetPath, destDir, archivePath)

	case tar.TypeReg:
		return e.extractRegularFileConcurrent(ctx, tarReader, header, targetPath, destDir, archivePath, workChan)

	case tar.TypeSymlink:
		return e.extractSymlinkConcurrent(header, targetPath, baseDir, destDir, archivePath)

	case tar.TypeLink:
		return e.extractHardLinkConcurrent(header, targetPath, baseDir, destDir, archivePath)

	default:
		// Skip unsupported entry types (e.g., character devices, block devices)
		return nil
	}
}

// extractDirectoryConcurrent extracts a directory entry concurrently.
func (e *Extractor) extractDirectoryConcurrent(header *tar.Header, targetPath, destDir, archivePath string) error {
	mode := os.FileMode(header.Mode & unixPermMask) // #nosec G115

	err := e.extractDirectory(targetPath, mode)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "extracting directory",
			Err:         err,
		}
	}

	return nil
}

// extractRegularFileConcurrent extracts a regular file entry concurrently.
// For large files, it uses sequential extraction to avoid overhead.
// For smaller files, it uses chunked extraction for concurrency benefits.
func (e *Extractor) extractRegularFileConcurrent(
	ctx context.Context,
	tarReader TarReader,
	header *tar.Header,
	targetPath, destDir, archivePath string,
	workChan chan<- fileExtractionWork,
) error {
	mode := os.FileMode(header.Mode & unixPermMask) // #nosec G115
	size := header.Size

	// Fallback to sequential extraction for large files
	if size > int64(e.bufferSize*chunkedThreshold) {
		return e.extractRegularFile(tarReader, targetPath, mode, size)
	}

	// For smaller files, use chunked extraction
	return e.extractFileInChunks(ctx, tarReader, targetPath, mode, size, destDir, archivePath, workChan)
}

// extractSymlinkConcurrent extracts a symlink entry concurrently.
func (e *Extractor) extractSymlinkConcurrent(
	header *tar.Header,
	targetPath, baseDir, destDir, archivePath string,
) error {
	err := e.extractSymlink(targetPath, header.Linkname, baseDir, destDir, archivePath)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "extracting symlink",
			Err:         err,
		}
	}

	return nil
}

// extractHardLinkConcurrent extracts a hard link entry concurrently.
func (e *Extractor) extractHardLinkConcurrent(
	header *tar.Header,
	targetPath, baseDir, destDir, archivePath string,
) error {
	err := e.extractHardLink(targetPath, header.Linkname, baseDir, destDir, archivePath)
	if err != nil {
		return &ExtractionError{
			ArchivePath: archivePath,
			Destination: destDir,
			Context:     "extracting hard link",
			Err:         err,
		}
	}

	return nil
}

// extractFileInChunks reads a file in chunks and sends each chunk to workers.
// This enables concurrent extraction for smaller files.
func (e *Extractor) extractFileInChunks(
	ctx context.Context,
	tarReader TarReader,
	targetPath string,
	mode os.FileMode,
	size int64,
	destDir, archivePath string,
	workChan chan<- fileExtractionWork,
) error {
	buffer := make([]byte, e.bufferSize)
	totalRead := int64(0)
	readCount := 0

	for totalRead < size {
		select {
		case <-ctx.Done():
			return ctx.Err() //nolint:wrapcheck // context error does not need wrapping
		default:
		}

		remaining := size - totalRead

		chunkSize := int64(e.bufferSize)
		if remaining < chunkSize {
			chunkSize = remaining
		}

		readCount++

		bytesRead, err := io.ReadFull(tarReader, buffer[:chunkSize])
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			return &ExtractionError{
				ArchivePath: archivePath,
				Destination: destDir,
				Context:     "reading file chunk",
				Err:         err,
			}
		}

		totalRead += int64(bytesRead)
		isEOF := totalRead >= size

		// Allocate new slice to prevent data corruption from buffer reuse
		data := make([]byte, bytesRead)
		copy(data, buffer[:bytesRead])

		select {
		case workChan <- fileExtractionWork{
			targetPath: targetPath,
			data:       data,
			mode:       mode,
			eof:        isEOF,
		}:
		case <-ctx.Done():
			return ctx.Err() //nolint:wrapcheck // context error does not need wrapping
		}

		if isEOF {
			break
		}
	}

	return nil
}

// validateEvalSymlinks performs common symlink evaluation and validation logic.
// It calls EvalSymlinks, cleans the result, checks the destDir prefix, and returns
// appropriate SecurityError with the provided validationType for EvalSymlinks errors
// and "symlink chain destination check" for invalid path cases.
func (e *Extractor) validateEvalSymlinks(targetPath, destDir, validationType string) error {
	evaled, err := e.fs.EvalSymlinks(targetPath)
	if err != nil {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    validationType,
			Err:           err,
		}
	}

	evaled = filepath.Clean(evaled)
	if !strings.HasPrefix(evaled, destDir+string(filepath.Separator)) && evaled != destDir {
		return &SecurityError{
			AttemptedPath: targetPath,
			Validation:    "symlink chain destination check",
			Err:           ErrInvalidPath,
		}
	}

	return nil
}

// validateSymlinkChain validates that a symlink chain resolves within the destination directory.
func (e *Extractor) validateSymlinkChain(resolved, destDir string) error {
	return e.validateEvalSymlinks(resolved, destDir, "symlink chain resolution")
}

// validateResolvedPath resolves any symlinks in the target path and validates
// that the resolved path stays within the destination directory.
func (e *Extractor) validateResolvedPath(targetPath, destDir string) error {
	return e.validateEvalSymlinks(targetPath, destDir, "resolved path destination check")
}

// validateLinkname validates the linkname for symlinks and hard links to prevent symlink attacks.
// It ensures the linkname does not contain absolute paths, ".." sequences, backslashes, or null bytes,
// resolves the path relative to the base directory and checks that the resolved path stays within the destination directory.
// Additionally, if the resolved path exists and is a symlink, it resolves any symlink chains and verifies
// the final resolved path is within the destination directory. It also prevents links to sensitive system files.
//
//nolint:lll,cyclop,funlen
func (e *Extractor) validateLinkname(linkname, baseDir, destDir string) error {
	if filepath.IsAbs(linkname) {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "absolute path prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(linkname, "..") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "parent directory reference prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(linkname, "\\") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "backslash prevention",
			Err:           ErrInvalidPath,
		}
	}

	if strings.Contains(linkname, "\x00") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "null byte prevention",
			Err:           ErrInvalidPath,
		}
	}

	resolved := filepath.Join(baseDir, linkname)
	resolved = filepath.Clean(resolved)

	// Check if resolved is within destDir
	if !strings.HasPrefix(resolved, destDir+string(filepath.Separator)) && resolved != destDir {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "linkname destination check",
			Err:           ErrInvalidPath,
		}
	}

	// Additional validation using filepath.Rel
	rel, err := filepath.Rel(destDir, resolved)
	if err != nil || strings.HasPrefix(rel, "..") {
		return &SecurityError{
			AttemptedPath: linkname,
			Validation:    "relative path validation",
			Err:           ErrInvalidPath,
		}
	}

	// Prevent links to sensitive system files
	sensitivePaths := []string{"/etc", "/usr", "/bin", "/sbin", "/dev", "/proc", "/sys", "/root", "/home"}
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(resolved, sensitive) {
			return &SecurityError{
				AttemptedPath: linkname,
				Validation:    "sensitive system path prevention",
				Err:           ErrInvalidPath,
			}
		}
	}

	// If the resolved path exists and is a symlink, resolve the symlink chain
	info, err := e.fs.Lstat(resolved)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		err = e.validateSymlinkChain(resolved, destDir)
		if err != nil {
			return &SecurityError{
				AttemptedPath: linkname,
				Validation:    "symlink chain validation",
				Err:           err,
			}
		}
	}

	return nil
}
