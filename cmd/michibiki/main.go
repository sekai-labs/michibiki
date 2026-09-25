package main

import (
	"fmt"
	"os"

	"github.com/sekai-labs/michibiki/internal/cli"
	_ "github.com/sekai-labs/michibiki/providers/frr"
	_ "github.com/sekai-labs/michibiki/providers/openwrt"
	_ "github.com/sekai-labs/michibiki/providers/opnsense"
	_ "github.com/sekai-labs/michibiki/providers/pfsense"
	_ "github.com/sekai-labs/michibiki/providers/routeros"
	_ "github.com/sekai-labs/michibiki/providers/vyos"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
