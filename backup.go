package cosmovisor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/otiai10/copy"
)

// BackupData backs up the data directory located at $DAEMON_BACKUP_DATA_DIR to
// $DAEMON_HOME/backups/$plan/data and create keep at $DAEMON_HOME/backups/$plan/.keep
func BackupData(cfg *Config, upgradeInfo *UpgradeInfo) error {
	backupDir := cfg.BackupDir(upgradeInfo.Name)
	// Stamp file for completion tracking.
	backupStamp := fmt.Sprintf("%s/.keep", backupDir)
	// If stamp exists, this plan has executed the backup already.
	if _, err := os.Stat(backupStamp); err == nil {
		return nil
	}
	// Make backup dir if it doesn't exist.
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(backupDir, 0700); err != nil {
			return err
		}
	}
	// Perform the copy from data src -> backup dst.
	if err := copy.Copy(cfg.DataDir, filepath.Join(backupDir, "data")); err != nil {
		return err
	}
	// Touch the stamp file if everything completed.
	if _, err := TouchFile(backupStamp); err != nil {
		return err
	}
	// Success!
	return nil
}

// TouchFile creates a file at the location similar to the POSIX `touch` command.
func TouchFile(file string) (time.Time, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file, ee := os.Create(file)
		if ee != nil {
			return time.Time{}, ee
		}
		defer file.Close()
	}

	currentTime := time.Now().Local()
	if err := os.Chtimes(file, currentTime, currentTime); err != nil {
		return time.Time{}, err
	}
	return currentTime, nil
}
