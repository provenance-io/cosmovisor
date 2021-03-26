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
var plainUpgradeRegex = regexp.MustCompile(`UPGRADE "(.*?)" NEEDED at height: (\d+):\s+(\S*)`)
var jsonUpgradeRegex = regexp.MustCompile(`UPGRADE (?:\\|)"(.*?)(?:\\|)" NEEDED at height: (\d+):\s+(.*)$`)

// UpgradeInfo is the details from the regexp
type UpgradeInfo struct {
	Name string
	Info string
}

type State int
const (
	INITIAL State = iota
	PENDING
)

const (
	UPGRADE_TEXT = "UPGRADE "
	NEEDED_TEXT = " NEEDED at "
	CONSENSUS_FAIL_TEXT = "CONSENSUS FAILURE!!!"
	PANIC_TEXT = "panic: UPGRADE"
)

// WaitForUpdate will listen to the scanner until a line matches upgradeRegexp.
// It returns (info, nil) on a matching line
// It returns (nil, err) if the input stream errored
// It returns (nil, nil) if the input closed without ever matching the regexp
func WaitForUpdate(scanner *bufio.Scanner) (*UpgradeInfo, error) {
	state := INITIAL

	var info *UpgradeInfo
	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case INITIAL:
			// Don't use the regexp unless we are actually looking at an upgrade line.
			// Compiled regex matching is about 20x more expensive than strings.Contains(). (10 vs 200).
			if !(strings.Contains(line, UPGRADE_TEXT) && strings.Contains(line, NEEDED_TEXT)) {
				continue
			}
			// Parse the info, and kick into holding for panic or consensus failure message.
			// Hacky: If starts with { and ends with }, parse into json object.
			if isJsonLog(line) {
				if !jsonUpgradeRegex.MatchString(line) {
					continue
				}

				jsonLine, err := parseJsonLog(line)
				if err != nil {
					return nil, err
				}

				subs := jsonUpgradeRegex.FindStringSubmatch(jsonLine.Message)
				info = &UpgradeInfo {
					Name: subs[1],
					Info: subs[3],
				}
				state = PENDING
				continue
			} else {
				subs := plainUpgradeRegex.FindStringSubmatch(line)
				info = &UpgradeInfo{
					Name: subs[1],
					Info: subs[3],
				}
				state = PENDING
				continue
			}
		case PENDING:
			// We have hit the panic or consensus failure after an upgrade log message, return out and update.
			if strings.Contains(line, PANIC_TEXT) || strings.Contains(line, CONSENSUS_FAIL_TEXT) {
				return info, nil
			}
			continue
		}
	}
	return nil, scanner.Err()
}
