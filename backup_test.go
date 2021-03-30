package cosmovisor_test

import (
	"fmt"
	"github.com/provenance-io/cosmovisor"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tree(t *testing.T, path string) string {
	t.Helper()

	s := ""
	filepath.Walk(path, func(p string, i fs.FileInfo, e error) error {
		d := ""
		if i.IsDir() {
			d = "/"
		}
		s += fmt.Sprintf("%s%s\n", p, d)
		return nil
	})
	return s
}

func (s *upgradeTestSuite) TestBackupDataDir() {
	home := copyTestData(s.T(), "validate")
	data := fmt.Sprintf("%s/%s", home, "data")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", DataDir: data}
	info := &cosmovisor.UpgradeInfo{Name: "chain2"}

	err := cosmovisor.DoUpgrade(cfg, info)
	s.Require().NoError(err)
	// Backup dir should now exist.
	backupDir := cfg.BackupDir(info.Name)
	s.Require().DirExists(backupDir)
	// Verify copied files exist.
	appDbName := fmt.Sprintf("%s/%s", backupDir, "data/application.db")
	s.Require().FileExists(appDbName)
	appBz, err := ioutil.ReadFile(appDbName)
	s.Require().NoError(err)
	s.Require().Equal(string(appBz), "test\n")
	stateDbName := fmt.Sprintf("%s/%s", backupDir, "data/modulesDir/state.db")
	s.Require().FileExists(stateDbName)
	stateBz, err := ioutil.ReadFile(stateDbName)
	s.Require().NoError(err)
	s.Require().Equal(string(stateBz), "test\n")
	// Verify keep file exists.
	keep := fmt.Sprintf("%s/%s", backupDir, ".keep")
	s.Require().FileExists(keep)
}

func (s *upgradeTestSuite) TestNoBackupDataDir() {
	home := copyTestData(s.T(), "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd", DataDir: ""}
	info := &cosmovisor.UpgradeInfo{Name: "chain2"}

	err := cosmovisor.DoUpgrade(cfg, info)
	s.Require().NoError(err)
	// Backup dir should not exist.
	upgradeDir := cfg.BackupDir(info.Name)
	s.Require().NoDirExists(upgradeDir)
	// Verify keep file exists.
	keep := fmt.Sprintf("%s/%s", upgradeDir, ".keep")
	s.Require().NoFileExists(keep)
}

func (s *upgradeTestSuite) TestTouchFile() {
	f, err := ioutil.TempFile("", "")
	defer os.Remove(f.Name())
	s.Require().NoError(err)
	// Verify that the touch updated the chtimes correctly.
	now := time.Now().Local()
	touchTime, err := cosmovisor.TouchFile(f.Name())
	s.Require().NoError(err)
	s.Require().True(touchTime.After(now))
	// Make sure touchTime is equal to mod time.
	info, err := os.Stat(f.Name())
	s.Require().NoError(err)
	s.Require().Equal(touchTime, info.ModTime())
}
