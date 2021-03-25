package cosmovisor

import (
	"bufio"
	"regexp"
	"strings"
)

// Trim off whitespace around the info - match least greedy, grab as much space on both sides
// Defined here: https://github.com/cosmos/cosmos-sdk/blob/release/v0.38.2/x/upgrade/abci.go#L38
//  fmt.Sprintf("UPGRADE \"%s\" NEEDED at %s: %s", plan.Name, plan.DueAt(), plan.Info)
// DueAt defined here: https://github.com/cosmos/cosmos-sdk/blob/release/v0.38.2/x/upgrade/internal/types/plan.go#L73-L78
//
//    if !p.Time.IsZero() {
//      return fmt.Sprintf("time: %s", p.Time.UTC().Format(time.RFC3339))
//    }
//    return fmt.Sprintf("height: %d", p.Height)

// Accommodate both json and plain logging formats (json: \"plan\", plain: "plan").
var consensusUpgradeRegex = regexp.MustCompile(`UPGRADE (?:\\|)"(.*?)(?:\\|)" NEEDED at height: (\d+):\s+([^ "]*)`)

// UpgradeInfo is the details from the regexp
type UpgradeInfo struct {
	Name string
	Info string
}

// WaitForUpdate will listen to the scanner until a line matches upgradeRegexp.
// It returns (info, nil) on a matching line
// It returns (nil, err) if the input stream errored
// It returns (nil, nil) if the input closed without ever matching the regexp
func WaitForUpdate(scanner *bufio.Scanner) (*UpgradeInfo, error) {
	for scanner.Scan() {
		line := scanner.Text()

		// Don't use the regexp unless we are actually looking at an upgrade line.
		// Compiled regex matching is about 20x more expensive than strings.Contains(). (10 vs 200).
		if !(strings.Contains(line, "UPGRADE") && strings.Contains(line, "NEEDED at ")) {
			continue
		}

		if !strings.Contains(line, "panic: UPGRADE") && !strings.Contains(line, "CONSENSUS FAILURE!!!") {
			continue
		}

		// Match the line for upgrade text with plan for the panic messages or consensus messages.
		subs := consensusUpgradeRegex.FindStringSubmatch(line)
		info := UpgradeInfo{
			Name: subs[1],
			Info: subs[3],
		}
		return &info, nil
	}
	return nil, scanner.Err()
}
