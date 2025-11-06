package install

import (
	"fmt"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// Validate checks if the archive can be extracted to the destination directory.
func (a *ArchiveServiceImpl) Validate(archivePath, destDir string) error {
	err := a.extractor.Validate(archivePath, destDir)
	if err != nil {
		logger.Debugf("failed to validate archive: %v", err)

		return fmt.Errorf("failed to validate archive: %w", err)
	}

	return nil
}

// Extract extracts the archive to the destination directory.
func (a *ArchiveServiceImpl) Extract(archivePath, destDir string) error {
	err := a.extractor.Extract(archivePath, destDir)
	if err != nil {
		logger.Debugf("ArchiveServiceImpl.Extract: extractor.Extract returned err: %v", err)

		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// ExtractVersion extracts the version from the archive path.
func (a *ArchiveServiceImpl) ExtractVersion(archivePath string) string {
	return archive.ExtractVersion(archivePath)
}
