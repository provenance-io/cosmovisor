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
		"consensus failure structured logging": {
			write: []string{
				`Jun  7 11:28:40 query-node-us-east1-0 cosmovisor[245614]: {"level":"error","time":"2021-06-07T11:28:40Z","message":"UPGRADE \"citrine\" NEEDED at height: 1582700: https://github.com/provenance-io/provenance/releases/download/v1.4.1/plan-v1.4.1.json"}` + "\n",
				`Jun  7 11:28:40 query-node-us-east1-0 cosmovisor[245614]: {"level":"error","module":"consensus","err":"UPGRADE \"citrine\" NEEDED at height: 1582700: https://github.com/provenance-io/provenance/releases/download/v1.4.1/plan-v1.4.1.json","stack":"goroutine 179 [running]:\nruntime/debug.Stack(0xc0018bbb48, 0x1d51c40, 0xc004550e90)\n\truntime/debug/stack.go:24 +0x9f\ngithub.com/tendermint/tendermint/consensus.(*State).receiveRoutine.func2(0xc0010f8a80, 0x22e04f0)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:726 +0x5b\npanic(0x1d51c40, 0xc004550e90)\n\truntime/panic.go:965 +0x1b9\ngithub.com/cosmos/cosmos-sdk/x/upgrade.BeginBlocker(0x7fffe0a60dd1, 0xc, 0xc000c1dec0, 0x252eb58, 0xc0010240c0, 0x2566ed8, 0xc000fb28e0, 0xc000d628d0, 0x254e888, 0xc00011c150, ...)\n\tgithub.com/cosmos/cosmos-sdk@v0.42.4/x/upgrade/abci.go:70 +0x11cb\ngithub.com/cosmos/cosmos-sdk/x/upgrade.AppModule.BeginBlock(...)\n\tgithub.com/cosmos/cosmos-sdk@v0.42.4/x/upgrade/module.go:127\ngithub.com/cosmos/cosmos-sdk/types/module.(*Manager).BeginBlock(0xc000b9b0a0, 0x254e888, 0xc00011c150, 0x25668e8, 0xc001d970c0, 0xb, 0x0, 0xc003a64df0, 0xd, 0x18266c, ...)\n\tgithub.com/cosmos/cosmos-sdk@v0.42.4/types/module/module.go:338 +0x1b8\ngithub.com/provenance-io/provenance/app.(*App).BeginBlocker(...)\n\tgithub.com/provenance-io/provenance/app/app.go:639\ngithub.com/cosmos/cosmos-sdk/baseapp.(*BaseApp).BeginBlock(0xc0010ff860, 0xc002715e80, 0x20, 0x20, 0xb, 0x0, 0xc003a64df0, 0xd, 0x18266c, 0x75030f2, ...)\n\tgithub.com/cosmos/cosmos-sdk@v0.42.4/baseapp/abci.go:179 +0x638\ngithub.com/tendermint/tendermint/abci/client.(*localClient).BeginBlockSync(0xc000cb5740, 0xc002715e80, 0x20, 0x20, 0xb, 0x0, 0xc003a64df0, 0xd, 0x18266c, 0x75030f2, ...)\n\tgithub.com/tendermint/tendermint@v0.34.10/abci/client/local_client.go:274 +0xfa\ngithub.com/tendermint/tendermint/proxy.(*appConnConsensus).BeginBlockSync(0xc00113ab40, 0xc002715e80, 0x20, 0x20, 0xb, 0x0, 0xc003a64df0, 0xd, 0x18266c, 0x75030f2, ...)\n\tgithub.com/tendermint/tendermint@v0.34.10/proxy/app_conn.go:81 +0x75\ngithub.com/tendermint/tendermint/state.execBlockOnProxyApp(0x254f5a8, 0xc000101aa0, 0x255bb28, 0xc00113ab40, 0xc0013fde00, 0x2566a88, 0xc00113a790, 0x1, 0xc004ab2840, 0x20, ...)\n\tgithub.com/tendermint/tendermint@v0.34.10/state/execution.go:307 +0x51b\ngithub.com/tendermint/tendermint/state.(*BlockExecutor).ApplyBlock(0xc000c2e380, 0xb, 0x0, 0x0, 0x0, 0xc001e1b250, 0xd, 0x1, 0x18266b, 0xc004ab2840, ...)\n\tgithub.com/tendermint/tendermint@v0.34.10/state/execution.go:140 +0x168\ngithub.com/tendermint/tendermint/consensus.(*State).finalizeCommit(0xc0010f8a80, 0x18266c)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:1635 +0xb48\ngithub.com/tendermint/tendermint/consensus.(*State).tryFinalizeCommit(0xc0010f8a80, 0x18266c)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:1546 +0x428\ngithub.com/tendermint/tendermint/consensus.(*State).enterCommit.func1(0xc0010f8a80, 0xc000000000, 0x18266c)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:1481 +0x8e\ngithub.com/tendermint/tendermint/consensus.(*State).enterCommit(0xc0010f8a80, 0x18266c, 0x0)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:1519 +0x6be\ngithub.com/tendermint/tendermint/consensus.(*State).addVote(0xc0010f8a80, 0xc0077d7400, 0xc000f7ff80, 0x28, 0x22e28a0, 0xc0018c1c08, 0x11e2eb9)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:2132 +0xde5\ngithub.com/tendermint/tendermint/consensus.(*State).tryAddVote(0xc0010f8a80, 0xc0077d7400, 0xc000f7ff80, 0x28, 0xc003f95100, 0xc00681d8c0, 0xc0279e9a3049885e)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:1930 +0x56\ngithub.com/tendermint/tendermint/consensus.(*State).handleMsg(0xc0010f8a80, 0x2508380, 0xc0023de3e8, 0xc000f7ff80, 0x28)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:838 +0x8cd\ngithub.com/tendermint/tendermint/consensus.(*State).receiveRoutine(0xc0010f8a80, 0x0)\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:762 +0x3f2\ncreated by github.com/tendermint/tendermint/consensus.(*State).OnStart\n\tgithub.com/tendermint/tendermint@v0.34.10/consensus/state.go:378 +0x8c5\n","time":"2021-06-07T11:28:40Z","message":"CONSENSUS FAILURE!!!"}` + "\n",
			},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "citrine",
				Info: "https://github.com/provenance-io/provenance/releases/download/v1.4.1/plan-v1.4.1.json",
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
