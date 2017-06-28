package utility

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"os"
	"path/filepath"
)

type FilePathHelper struct {
	Current string
}

func NewFilePathHelper(p string) FilePathHelper {
	wp := FilePathHelper{Current: p}
	return wp
}

func (cd *FilePathHelper) AddWorkingDirPath(f string) string {
	return filepath.Join(cd.Current, f)
}

func (cd *FilePathHelper) AddBackupDirPath(f string) string {
	return filepath.Join(cd.Current, config.BACKUP_DIR, f)
}

func (cd *FilePathHelper) AddStagingDirPath(f string) string {
	return filepath.Join(cd.Current, config.STAGING_DIR, f)
}

func (cd *FilePathHelper) WorkingToBackup(fileName string) error {
	return os.Rename(cd.AddWorkingDirPath(fileName), cd.AddBackupDirPath(fileName))
}

func (cd *FilePathHelper) BackupToWorking(fileName string) error {
	return os.Rename(cd.AddBackupDirPath(fileName), cd.AddWorkingDirPath(fileName))
}
func (cd *FilePathHelper) StagingToWorking(fileName string) error {
	fmt.Printf("moving %v -> %v\n", cd.AddStagingDirPath(fileName), cd.AddWorkingDirPath(fileName))
	err := os.Rename(cd.AddStagingDirPath(fileName), cd.AddWorkingDirPath(fileName))
	if err != nil {
		fmt.Printf("Error stw: %v\n", err.Error())
	}
	return err
}
