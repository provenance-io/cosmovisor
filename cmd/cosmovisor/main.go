package main

import (
	"fmt"
	"os"

	"github.com/provenance-io/cosmovisor/version"

	"github.com/provenance-io/cosmovisor"
)

func main() {
	if os.Getenv("DAEMON_INFO") != "" {
		fmt.Fprintf(os.Stderr, "%s\n", version.BuildInfo())
		return
	}

	if err := Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

// Run is the main loop, but returns an error
func Run(args []string) error {
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}

	doUpgrade, err := cosmovisor.LaunchProcess(cfg, args, os.Stdout, os.Stderr)
	// if RestartAfterUpgrade, we launch after a successful upgrade (only condition LaunchProcess returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, os.Stdout, os.Stderr)
	}
	return err
}
