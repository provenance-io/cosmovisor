package cosmovisor_test

import (
	"bufio"
	"io"
	"testing"

	"github.com/provenance-io/cosmovisor"

	"github.com/stretchr/testify/require"
)

func TestWaitForInfo(t *testing.T) {
	cases := map[string]struct {
		write         []string
		expectUpgrade *cosmovisor.UpgradeInfo
		expectErr     bool
	}{
		"no match": {
			write: []string{
				"some",
				"random\ninfo\n",
			},
		},
		"old match name with no info - ignored": {
			write: []string{
				"first line\n",
				"UPGRADE \"myname\" NEEDED at height: 123: \n",
				"next line\n",
			},
		},
		"old match name with info - ignored": {
			write: []string{
				"first line\n",
				"UPGRADE \"take2\" NEEDED at height: 123:   DownloadData here!\n",
				"next line\n",
			},
		},
		"no match consensus failure": {
			write: []string{
				"first line\n",
				"CONSENSUS FAILURE!!! err=\"some random error\" module=consensus\n",
			},
		},
		"match consensus failure no info": {
			write: []string{
				"first line\n",
				`01:00 ERR UPGRADE "myname" NEEDED at height: 123: ` + "\n",
				`err="UPGRADE \"myname\" NEEDED at height: 123: " module=consensus message="CONSENSUS FAILURE!!!"` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "myname",
				Info: "",
			},
		},
		"match consensus failure with info": {
			write: []string{
				"first line\n",
				`01:00 ERR UPGRADE "test" NEEDED at height: 10: /app/plan.json` + "\n",
				`"err="UPGRADE \"test\" NEEDED at height: 10: /app/plan.json" another=thing module=consensus stack="goroutine 91 [running]:\nruntime/debug.Stack(0xc001709a98, 0x1c3cb40, 0xc001df3620)\n\truntime/debug/stack.go:24 +0x9f\ngithub.com/tendermint/tendermint/consensus.(*State).receiveRoutine.func2(0xc001250000, 0x21b4ba0)\n\tgithub.com/tendermint/tendermint@v0.34.8/consensus/state.go:726" message="CONSENSUS FAILURE!!!"` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "test",
				Info: "/app/plan.json",
			},
		},
		"match consensus failure json with no info": {
			write: []string{
				`{"level":"error","time":"2021-03-24T20:33:13Z","message":"UPGRADE \"jsontest\" NEEDED at height: 10: "}` + "\n",
				`{"level":"error","module":"consensus","err":"UPGRADE \"jsontest\" NEEDED at height: 10: ","message":"CONSENSUS FAILURE!!!"}` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "jsontest",
				Info: "",
			},
		},
		"match consensus failure json with info": {
			write: []string{
				`{"level":"error","time":"2021-03-24T20:33:13Z","message":"UPGRADE \"jsontest\" NEEDED at height: 10: /app/plan.json"}` + "\n",
				`{"level":"error","module":"consensus","err":"UPGRADE \"jsontest\" NEEDED at height: 10: /app/plan.json","message":"CONSENSUS FAILURE!!!"}` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "jsontest",
				Info: "/app/plan.json",
			},
		},
		"panic text with no info": {
			write: []string{
				"first line\n",
				`01:00 ERR UPGRADE "test-panic" NEEDED at height: 10: ` + "\n",
				`panic: UPGRADE "test-panic" NEEDED at height: 10: ` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "test-panic",
				Info: "",
			},
		},
		"panic text with info": {
			write: []string{
				"first line\n",
				`01:00 ERR UPGRADE "test-panic" NEEDED at height: 10: /app/plan.json` + "\n",
				`panic: UPGRADE "test-panic" NEEDED at height: 10: /app/plan.json` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "test-panic",
				Info: "/app/plan.json",
			},
		},
		"panic text with info as json": {
			write: []string{
				`01:00 ERR UPGRADE "chain2" NEEDED at height: 49: {"binaries":{"linux/amd64":"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"}}` + "\n",
				`panic: UPGRADE "chain2" NEEDED at height: 49: {"binaries":{"linux/amd64":"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"}}` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "chain2",
				Info: `{"binaries":{"linux/amd64":"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"}}`,
			},
		},
		"consensus failure with info as json": {
			write: []string{
				`01:00 ERR UPGRADE "chain2" NEEDED at height: 49: {"binaries":{"linux/amd64":"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"}}` + "\n",
				`message="CONSENSUS FAILURE!!!" err="UPGRADE \"chain2\" NEEDED at height: 49: {\"binaries\":{\"linux/amd64\":\"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998\"}}"` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "chain2",
				Info: `{"binaries":{"linux/amd64":"https://github.com/cosmos/cosmos-sdk/raw/51249cb93130810033408934454841c98423ed4b/cosmovisor/testdata/repo/zip_binary/autod.zip?checksum=sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"}}`,
			},
		},
		"panic text with info as https": {
			write: []string{
				`01:00 ERR UPGRADE "chain2" NEEDED at height: 49: https://really.cool.network/downloads/v0/download.zip?sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998` + "\n",
				`panic: UPGRADE "chain2" NEEDED at height: 49: https://really.cool.network/downloads/v0/download.zip?sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "chain2",
				Info: `https://really.cool.network/downloads/v0/download.zip?sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998`,
			},
		},
		"consensus failure with info as https": {
			write: []string{
				`01:00 ERR UPGRADE "chain2" NEEDED at height: 49: https://really.cool.network/downloads/v0/download.zip?sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998` + "\n",
				`message="CONSENSUS FAILURE!!!" err="UPGRADE \"chain2\" NEEDED at height: 49: https://really.cool.network/downloads/v0/download.zip?sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998"` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "chain2",
				Info: `https://really.cool.network/downloads/v0/download.zip?sha256:dc48829b4126ae95bc0db316c66d4e9da5f3db95e212665b6080638cca77e998`,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, w := io.Pipe()
			scan := bufio.NewScanner(r)

			// write all info in separate routine
			go func() {
				for _, line := range tc.write {
					n, err := w.Write([]byte(line))
					require.NoError(t, err)
					require.Equal(t, len(line), n)
				}
				w.Close()
			}()

			// now scan the info
			info, err := cosmovisor.WaitForUpdate(scan)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectUpgrade, info)
		})
	}
}
