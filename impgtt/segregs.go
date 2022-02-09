package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"

	"github.com/antchfx/xmlquery"
	_ "github.com/hhrutter/tiff"
	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(srCmd)
	srCmd.Flags().IntVarP(&srArgs.padding, "padding", "p", 0, "set padding for region image files")
	srCmd.Flags().IntVarP(&srArgs.workers, "workers", "w", runtime.NumCPU(), "set number of concurrent workers")
	srCmd.Flags().BoolVarP(&srArgs.lines, "lines", "l", false, "segment line regions")
}

var srCmd = &cobra.Command{
	Use:   "segregs XML IMG OUT",
	Short: "Segment regions from IMPACT-PageXML ground-truth files",
	Args:  cobra.ExactArgs(3),
	Run:   segregs,
}

var srArgs = struct {
	padding, workers int
	lines            bool
}{}

func segregs(_ *cobra.Command, args []string) {
	r := srRunner{
		padding: srArgs.padding,
		workers: srArgs.workers,
		lines:   srArgs.lines,
	}
	r.run(args[0], args[1], args[2])
}

type srRunner struct {
	padding, workers int
	lines            bool
}

func (ru srRunner) run(xmlName, imgName, outBase string) {
	// Read the iamge once.
	in, err := os.Open(imgName)
	chk(err)
	defer in.Close()
	img, _, err := image.Decode(in)
	chk(err)
	var wg sync.WaitGroup
	wg.Add(ru.workers + 1)
	out := make(chan region)
	go func() {
		defer wg.Done()
		defer close(out)
		for _, r := range ru.regions(xmlName) {
			r.PageImage = filepath.Base(imgName)
			out <- r
		}
	}()
	for i := 0; i < ru.workers; i++ {
		go func() {
			defer wg.Done()
			for r := range out {
				srInitRegion(r, img, outBase, ru.padding, ru.lines)
			}
		}()
	}
	wg.Wait()
}

func (ru srRunner) regions(name string) []region {
	in, err := os.Open(name)
	chk(err)
	defer in.Close()
	xml, err := xmlquery.Parse(in)
	chk(err)
	rs := ru.findRegions(xml)
	idxRe := regexp.MustCompile(`\d+$`)
	var ret []region
	for _, r := range rs {
		// Read the region's polygon and inner text.
		polygon, err := srMakePolygon(r)
		chk(err)
		textnodes := xmlquery.Find(r, "/*[local-name()='TextEquiv']/*[local-name()='Unicode']")
		if len(textnodes) == 0 { // Skip regions with missing Unicode node.
			continue
		}
		textnode := textnodes[len(textnodes)-1]
		reg := region{
			PageXML:     filepath.Base(name),
			Coordinates: polygon,
			Text:        textnode.InnerText(),
		}
		for _, attr := range r.Attr {
			reg.putAttr(attr.Name.Local, attr.Value)
		}
		idx := idxRe.FindString(reg.Attrs["id"])
		i, err := strconv.Atoi(idx)
		if err != nil {
			chk(fmt.Errorf("bad id: %s", reg.Attrs["id"]))
		}
		reg.Index = i
		ret = append(ret, reg)
	}
	return ret
}

func (ru srRunner) findRegions(root *xmlquery.Node) []*xmlquery.Node {
	if ru.lines {
		return xmlquery.Find(root, "//*[local-name()='TextLine']")
	}
	return xmlquery.Find(root, "//*[local-name()='TextRegion']")
}

func srMakePolygon(r *xmlquery.Node) (polygon, error) {
	fail := func(err error) (polygon, error) {
		return polygon{}, fmt.Errorf("make polygon: %v", err)
	}
	ps := xmlquery.Find(r, "./*[local-name()='Coords']/*[local-name()='Point']")
	if len(ps) > 0 {
		return srMakePolygonFromPoints(ps)
	}
	coords := xmlquery.FindOne(r, "./*[local-name()='Coords']")
	if coords == nil {
		return fail(fmt.Errorf("cannot find polygon for region"))
	}
	// No Point nodes; use points attribute.
	for _, attr := range coords.Attr {
		if attr.Name.Local == "points" {
			return makePolygon(attr.Value)
		}
	}
	// We cannot find the polygon for this region.
	return fail(fmt.Errorf("cannot find polygon for region"))
}

func srMakePolygonFromPoints(points []*xmlquery.Node) (polygon, error) {
	attrAsInt := func(node *xmlquery.Node, key string) (int, bool) {
		for _, attr := range node.Attr {
			if attr.Name.Local != key {
				continue
			}
			val, err := strconv.Atoi(attr.Value)
			if err != nil {
				return 0, false
			}
			return val, true
		}
		return 0, false
	}
	var ret polygon
	for _, point := range points {
		x, xok := attrAsInt(point, "x")
		y, yok := attrAsInt(point, "y")
		if xok && yok {
			ret = append(ret, image.Point{X: x, Y: y})
		}
	}
	return ret, nil
}

func srInitRegion(r region, img image.Image, outBase string, padding int, lines bool) {
	// Copy the subregion from the base image.
	coords := r.Coordinates
	rect := coords.boundingRectangle()
	newRect := srAddPadding(rect, img.Bounds().Max, padding)
	newImg := image.NewRGBA(newRect)
	draw.Draw(newImg, newImg.Bounds(), img, newRect.Min, draw.Src)

	// Mask off pixels outside of the polygon.  Since newImg
	// retains the bounds of the original sub image, we do not
	// need to adjust for the new x- and y-coordinates.
	for x := newImg.Bounds().Min.X; x < newImg.Bounds().Max.X; x++ {
		for y := newImg.Bounds().Min.Y; y < newImg.Bounds().Max.Y; y++ {
			if !coords.inside(image.Pt(x, y)) {
				newImg.Set(x, y, color.White)
			}
		}
	}
	if lines {
		srWriteLineRegion(r, newImg, outBase)
	} else {
		srWriteRegion(r, newImg, outBase)
	}
}

func srWriteLineRegion(r region, img image.Image, outBase string) {
	// Write line png and gt.txt to outBase directory.
	base := filepath.Join(outBase, fmt.Sprintf("%05d", r.Index))
	pout, err := os.Create(base + ".png")
	chk(err)
	defer func() { chk(pout.Close()) }()
	chk(png.Encode(pout, img))
	chk(ioutil.WriteFile(base+".gt.txt", []byte(r.Text+"\n"), 0666))
}

func srWriteRegion(r region, img image.Image, outBase string) {
	// Write region png, json and gt.txt files.
	dir := fmt.Sprintf("%s_%s", outBase, r.Attrs["id"])
	image := dir + ".png"

	// Write image file.
	pout, err := os.Create(image)
	chk(err)
	defer func() { chk(pout.Close()) }()
	chk(png.Encode(pout, img))

	// Write ground-truth text file.
	gtout := dir + ".gt.txt"
	chk(ioutil.WriteFile(gtout, []byte(r.Text+"\n"), 0666))

	// Write json metadata.
	jout, err := os.Create(dir + ".json")
	chk(err)
	defer func() { chk(jout.Close()) }()
	r.Dir = filepath.Base(dir)
	r.Image = filepath.Base(image)
	r.GT = filepath.Base(gtout)
	chk(json.NewEncoder(jout).Encode(r))
}

func srAddPadding(rect image.Rectangle, max image.Point, padding int) image.Rectangle {
	minCap := func(a int) int {
		if a < 0 {
			return 0
		}
		return a
	}
	maxCap := func(a, b int) int {
		if a > b {
			return b
		}
		return a
	}
	rect.Min.X = minCap(rect.Min.X - padding)
	rect.Min.Y = minCap(rect.Min.Y - padding)
	rect.Max.X = maxCap(rect.Max.X+padding, max.X)
	rect.Max.Y = maxCap(rect.Max.Y+padding, max.Y)
	return rect
}
