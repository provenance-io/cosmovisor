package cosmovisor_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/cosmovisor"
)

type processTestSuite struct {
	suite.Suite
}

func TestProcessTestSuite(t *testing.T) {
	suite.Run(t, new(processTestSuite))
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func (s *processTestSuite) TestLaunchProcessLocal() {
	home := copyTestData(s.T(), "validate")
	cfg := &cosmovisor.Config{Home: home, Name: "dummyd"}

	// should run the genesis binary and produce expected output
	var stdout, stderr bytes.Buffer
	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)

	args := []string{"foo", "bar", "1234"}
	doUpgrade, err := cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().True(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Genesis foo bar 1234\nUPGRADE \"chain2\" NEEDED at height: 49: {}\npanic: UPGRADE \"chain2\" NEEDED at height: 49: {}\n", stdout.String())

	// ensure this is upgraded now and produces new output

	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"second", "run", "--verbose"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().False(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Chain 2 is live!\nArgs: second run --verbose\nFinished successfully\n", stdout.String())

	// ended without other upgrade
	s.Require().Equal(cfg.UpgradeBin("chain2"), currentBin)
}

// TestLaunchProcess will try running the script a few times and watch upgrades work properly
// and args are passed through
func (s *processTestSuite) TestLaunchProcessWithDownloads() {
	// this is a fun path
	// genesis -> "chain2" = zip_binary
	// zip_binary -> "chain3" = ref_zipped -> zip_directory
	// zip_directory no upgrade
	home := copyTestData(s.T(), "download")
	cfg := &cosmovisor.Config{Home: home, Name: "autod", AllowDownloadBinaries: true}

	// should run the genesis binary and produce expected output
	var stdout, stderr bytes.Buffer
	currentBin, err := cfg.CurrentBin()
	s.Require().NoError(err)

	s.Require().Equal(cfg.GenesisBin(), currentBin)
	args := []string{"some", "args"}
	doUpgrade, err := cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().True(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal(
		"Preparing auto-download some args\n"+
			`ERROR: UPGRADE "chain2" NEEDED at height: 49: {"binaries":{"any":"https://raw.githubusercontent.com/provenance-io/cosmovisor/main/testdata/repo/zip_binary/autod.zip?checksum=sha256:625f3888456c57b1b1f7706243864497bc7ee18d7e8f30de792bbc6150815d54"}} module=main`+"\n" +
			`ERROR: CONSENSUS FAILURE!!! err="UPGRADE \"chain2\" NEEDED at height: 49: {\"binaries\":{\"any\":\"https://raw.githubusercontent.com/provenance-io/cosmovisor/main/testdata/repo/zip_binary/autod.zip?checksum=sha256:625f3888456c57b1b1f7706243864497bc7ee18d7e8f30de792bbc6150815d54\"}}" module=main`+"\n",
			stdout.String(),
		)

	// ensure this is upgraded now and produces new output
	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain2"), currentBin)
	args = []string{"run", "--fast"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().True(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal(
		"Chain 2 from zipped binary link to referral\n"+
			"Args: run --fast\n"+
			`ERROR: UPGRADE "chain3" NEEDED at height: 936: https://raw.githubusercontent.com/provenance-io/cosmovisor/main/testdata/repo/ref_zipped?checksum=sha256:3d370b9b483c779b6cbaa7dbd266da6cacf9eb8f29b0bfb66e16d4fa8ba02b3a module=main`+"\n"+
			`ERROR: CONSENSUS FAILURE!!! err="UPGRADE \"chain3\" NEEDED at height: 936: https://raw.githubusercontent.com/provenance-io/cosmovisor/main/testdata/repo/ref_zipped?checksum=sha256:3d370b9b483c779b6cbaa7dbd266da6cacf9eb8f29b0bfb66e16d4fa8ba02b3a" module=main`+"\n",
			stdout.String(),
		)

	// ended with one more upgrade
	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain3"), currentBin)
	// make sure this is the proper binary now....
	args = []string{"end", "--halt"}
	stdout.Reset()
	stderr.Reset()
	doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, &stdout, &stderr)
	s.Require().NoError(err)
	s.Require().False(doUpgrade)
	s.Require().Equal("", stderr.String())
	s.Require().Equal("Chain 2 from zipped directory\nArgs: end --halt\n", stdout.String())

	// and this doesn't upgrade
	currentBin, err = cfg.CurrentBin()
	s.Require().NoError(err)
	s.Require().Equal(cfg.UpgradeBin("chain3"), currentBin)
}
