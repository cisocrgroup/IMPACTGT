package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(ccCmd)
	ccCmd.Flags().BoolVarP(&ccArgs.unknown, "unknown", "u", false, "only display unknown codes")
}

var ccCmd = &cobra.Command{
	Use:   "codes [DIR...]",
	Short: "list not printable codes in the files",
	Run:   cc,
}

var ccArgs = struct {
	unknown bool
}{}

func cc(_ *cobra.Command, dirs []string) {
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		chk(err)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			region, err := readRegion(filepath.Join(dir, entry.Name()))
			chk(err)
			for _, snip := range region.Snippets {
				for _, c := range snip.GT {
					if unicode.IsPrint(c) {
						continue
					}
					repl, ok := glyphs[c]
					if ok && ccArgs.unknown {
						continue
					}
					if !ok {
						repl = "UNKNOWN"
					}
					_, err := fmt.Printf("[%4X] %s\n", c, repl)
					chk(err)
				}
			}
		}
	}
}
