package install

import (
	"fmt"

	"github.com/nicholas-fedor/goUpdater/internal/archive"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

func (a *archiveServiceImpl) Validate(archivePath, destDir string) error {
	err := a.extractor.Validate(archivePath, destDir)
	if err != nil {
		logger.Debugf("failed to validate archive: %v", err)

		return fmt.Errorf("failed to validate archive: %w", err)
	}

	return nil
}

func (a *archiveServiceImpl) Extract(archivePath, destDir string) error {
	err := a.extractor.Extract(archivePath, destDir)
	if err != nil {
		logger.Debugf("archiveServiceImpl.Extract: extractor.Extract returned err: %v", err)

		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

func (a *archiveServiceImpl) ExtractVersion(archivePath string) string {
	return archive.ExtractVersion(archivePath)
}
