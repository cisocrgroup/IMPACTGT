package main

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	dropcapital = "drop-capital"
)

func init() {
	fixCmd.AddCommand(fdcCmd)
	fdcCmd.Flags().IntVarP(&fdcArgs.border, "border", "b", 3, "set the minimal white border around the image")
}

var fdcCmd = &cobra.Command{
	Use:   "dropcapitals [JSON...]",
	Short: "fix the segmentation of drop-capitals",
	Run:   fdc,
}

var fdcArgs = struct {
	border int
}{}

func fdc(_ *cobra.Command, jsons []string) {
	for _, json := range jsons {
		reg, err := readRegion(json)
		chk(err)
		if reg.Attrs[typ] != dropcapital {
			continue
		}
		fixDropCapital(json, reg)
	}
}

func fixDropCapital(json string, reg region) {
	slUseSingleRegionImage(json, filepath.Dir(json), reg.Text, reg, fdcArgs.border)
}
