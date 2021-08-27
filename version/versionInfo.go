package version

import (
	"fmt"
	"strings"
)

var (
	Name    string
	Module  string
	Version string
	Commit  string
	Built   string
)

func BuildInfo() string {
	items := []string{
		fmt.Sprintf("Name     : %s", Name),
		fmt.Sprintf("Module   : %s", Module),
		fmt.Sprintf("Version  : %s", Version),
		fmt.Sprintf("Commit   : %s", Commit),
		fmt.Sprintf("Built    : %s", Built),
	}
	return strings.Join(items, "\n")
}
