package main

import (
	"image"
	"math"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

const (
	heading = "heading"
	typ     = "type"
)

func init() {
	fixCmd.AddCommand(fhCmd)
	root.AddCommand(fixCmd)
	fhCmd.Flags().Float64VarP(&fhArgs.threshold, "threshold", "t", 0.75, "set relative levenshtein threshold")
}

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "fix the segmentation of regions",
}

var fhCmd = &cobra.Command{
	Use:   "headings [JSON...]",
	Short: "fix the segmentation of heading regions",
	Run:   fh,
}

var fhArgs = struct {
	threshold float64
}{}

func fh(_ *cobra.Command, jsons []string) {
	for _, json := range jsons {
		reg, err := readRegion(json)
		chk(err)
		if reg.Attrs[typ] != heading {
			continue
		}
		fixHeading(filepath.Dir(json), reg)
	}
}

func fixHeading(base string, reg region) {
	for _, snip := range reg.Snippets {
		cands := findMergeCandidates(reg, snip)
		if len(cands) == 0 {
			continue
		}
		if snip.relev() < fhArgs.threshold {
			continue
		}
		logf("merge candidates for %s (%g):", snip.Image, snip.relev())
		mergeSnippets(base, reg, snip, cands)
	}
}

type subimager interface {
	SubImage(image.Rectangle) image.Image
}

func mergeSnippets(base string, reg region, snip snippet, parts []snippet) {
	r := snip.Coordinates.boundingRectangle()
	logf(" - %s", r)
	for _, part := range parts {
		logf(" - %s", part.Coordinates.boundingRectangle())
		r = r.Union(part.Coordinates.boundingRectangle())
	}
	logf(" - %s: %s", reg.Image, r)
	regImg := filepath.Join(base, reg.Image)
	img := openImage(regImg)
	sub := img.(subimager).SubImage(r)
	writePNG(filepath.Join(base, snip.Image), sub)
}

func findMergeCandidates(reg region, snip snippet) []snippet {
	var ret []snippet
	neighbours := doFindMergeCandidates(reg, snip)
	ret = append(ret, neighbours...)
	for i := range ret {
		neighbours := doFindMergeCandidates(reg, ret[i])
		ret = append(ret, neighbours...)
	}
	// Do we need to sort this?
	sort.Slice(ret, func(i, j int) bool { return ret[i].num() < ret[j].num() })
	return ret
}

func doFindMergeCandidates(reg region, snip snippet) []snippet {
	neighbours := func(a, b snippet) bool {
		return int(math.Abs(float64(a.num()-b.num()))) == 1
	}
	var ret []snippet
	for _, x := range reg.Waste {
		if neighbours(snip, x) {
			ret = append(ret, x)
		}
	}
	return ret
}
