package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(slCmd)
	slCmd.Flags().StringVar(&slArgs.imgext, "imgext", ".bin.png", "set file extension for the image files")
	slCmd.Flags().StringVar(&slArgs.gtext, "gtext", ".gt.txt", "set file extension for output gt files")
	slCmd.Flags().StringVar(&slArgs.rectext, "rectext", ".rect.txt", "set file extension for output gt files")
}

var slCmd = &cobra.Command{
	Use:   "seglines [JSON...]",
	Short: "Segment lines",
	Args:  cobra.MinimumNArgs(1),
	Run:   seglines,
}

var slArgs = struct {
	imgext, ocrext, gtext, rectext string
}{}

func seglines(_ *cobra.Command, args []string) {
	for _, json := range args {
		bdir := filepath.Dir(json)
		r, err := readRegion(json)
		chk(err)

		lines := slGatherGTLines(r)
		nimg := slCountImageFiles(filepath.Join(bdir, r.Dir))

		if len(lines) == 1 && nimg > 1 {
			slUseSingleRegionImage(json, bdir, lines[0], r, 3)
			continue
		}
		if nimg > 1 {
			continue
		}
		if nimg != 0 && (len(lines) == nimg || len(lines) <= 1) {
			continue
		}
		slSegImg(filepath.Join(bdir, r.Dir), filepath.Join(bdir, r.Image), lines)
	}
}

func slUseSingleRegionImage(json, bdir, line string, reg region, bounds int) {
	logf("using single image for %s", filepath.Join(bdir, reg.Image))
	var (
		basename  = filepath.Join(reg.Dir, fmt.Sprintf("%06x", indexStart))
		imagename = basename + slArgs.imgext
	)
	if len(reg.Snippets) > 0 {
		basename = reg.Snippets[0].BaseName
		imagename = reg.Snippets[0].Image
	}
	chk(slCleanDir(filepath.Join(bdir, reg.Dir)))
	dst := filepath.Join(bdir, imagename)
	src := filepath.Join(bdir, reg.Image)

	// Copy region image as snippet and remove .
	r, err := os.Open(src)
	chk(err)
	defer r.Close()
	img, err := png.Decode(r)
	chk(err)
	rect := slImageBounds(img, bounds)
	img = img.(subimager).SubImage(rect)
	w, err := os.Create(dst)
	chk(err)
	defer w.Close()
	chk(png.Encode(w, img))

	// Use single Snippet and write region metadata file.
	reg.Snippets = []snippet{{
		GT:          line,
		Image:       imagename, // Use the relative path.
		Coordinates: makePolygonFromRectangle(img.Bounds()),
		BaseName:    basename,
	}}
	chk(reg.write(json))
	chk(slWriteRect(filepath.Join(bdir, basename+slArgs.rectext), img.Bounds()))
}

func slSegImg(dir, name string, lines []string) {
	logf("segmenting %s into %d lines", name, len(lines))
	chk(os.MkdirAll(dir, 0777))
	n := len(lines)
	img := yClip(openImage(name))
	bnds := img.Bounds()
	var cs []int
	for y := 0; y < bnds.Max.Y; y++ {
		cs = append(cs, xPixCount(img, y))
	}
	from := bnds.Min
	for i := 0; i < n; i++ {
		for ; from.Y < bnds.Max.Y; from.Y++ {
			if cs[from.Y] != 0 {
				break
			}
		}
		to := image.Point{X: bnds.Max.X, Y: bnds.Max.Y}
		if i != n-1 {
			// Find minimum black pixel line.
			h := bnds.Max.Y - from.Y
			lh := h / (n - i)
			s, e := from.Y+(2*lh/3), from.Y+(2*lh)
			css := cs[s:e]
			min := argmin(css)
			to = image.Point{X: bnds.Max.X, Y: s + min}
		}
		rect := image.Rectangle{Min: from, Max: to}
		if rect.Empty() {
			logf("skipping snippet line %d: image is empty", i+1)
			continue
		}
		// Write image snippet, rectangle and gt line file.
		snippet := img.(interface {
			SubImage(image.Rectangle) image.Image
		}).SubImage(rect)
		snippet = xClip(snippet)
		outName := filepath.Join(dir, fmt.Sprintf("%06d%s", i+1, slArgs.imgext))
		writePNG(outName, snippet)
		gtName := filepath.Join(dir, fmt.Sprintf("%06d%s", i+1, slArgs.gtext))
		chk(ioutil.WriteFile(gtName, []byte(lines[i]+"\n"), 0666))
		rectName := filepath.Join(dir, fmt.Sprintf("%06d%s", i+1, slArgs.rectext))
		chk(slWriteRect(rectName, snippet.Bounds()))
		from.Y = to.Y
	}
}

func slWriteRect(name string, r image.Rectangle) error {
	return ioutil.WriteFile(
		name,
		[]byte(fmt.Sprintf("%d,%d %d,%d", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)),
		0666,
	)
}

func slCleanDir(name string) error {
	err := os.RemoveAll(name)
	if err != nil {
		return err
	}
	return os.Mkdir(name, 0755)
}

func openImage(name string) image.Image {
	in, err := os.Open(name)
	chk(err)
	defer in.Close()
	img, _, err := image.Decode(in)
	chk(err)
	return img
}

func writePNG(name string, img image.Image) {
	out, err := os.Create(name)
	chk(err)
	defer func() { chk(out.Close()) }()
	chk(png.Encode(out, img))
}

func yClip(img image.Image) image.Image {
	bnds := img.Bounds()
	var b int
	for b = 0; b < bnds.Max.Y; b++ {
		if xPixCount(img, b) != 0 {
			break
		}
	}
	var e int
	for e = bnds.Max.Y; e > b; e-- {
		if xPixCount(img, e-1) != 0 {
			break
		}
	}
	return img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(image.Rect(0, b, bnds.Max.X, e))
}

func xClip(img image.Image) image.Image {
	bnds := img.Bounds()
	var b int
	for b = bnds.Min.X; b < bnds.Max.X; b++ {
		if yPixCount(img, b) != 0 {
			break
		}
	}
	var e int
	for e = bnds.Max.X; e > b; e-- {
		if yPixCount(img, e-1) != 0 {
			break
		}
	}
	return img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(image.Rect(b, bnds.Min.Y, e, bnds.Max.Y))
}

func xPixCount(img image.Image, y int) int {
	black := img.ColorModel().Convert(color.Black)
	bnds := img.Bounds()
	var c int
	for x := bnds.Min.X; x < bnds.Max.X; x++ {
		if img.At(x, y) == black {
			c++
		}
	}
	return c
}

func yPixCount(img image.Image, x int) int {
	black := img.ColorModel().Convert(color.Black)
	bnds := img.Bounds()
	var c int
	for y := bnds.Min.Y; y < bnds.Max.Y; y++ {
		if img.At(x, y) == black {
			c++
		}
	}
	return c
}

// len(args)>0
func argmin(cs []int) int {
	min := cs[0]
	var argmin int
	for i := 1; i < len(cs); i++ {
		if cs[i] < min {
			min = cs[i]
			argmin = i
		}
		if cs[i] > min*10 {
			break
		}
	}
	return argmin
}

func slCountImageFiles(dir string) int {
	var n int
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), slArgs.imgext) {
			n++
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return n
}

func slGatherGTLines(r region) []string {
	var ret []string
	rd := strings.NewReader(r.Text)
	s := bufio.NewScanner(rd)
	for s.Scan() {
		ret = append(ret, s.Text())
	}
	chk(s.Err())
	return ret
}

func slImageBounds(img image.Image, min int) image.Rectangle {
	b := img.Bounds()
	top, bottom, left, right := 0, 0, 0, 0
	for x := b.Min.X; x < b.Max.X; x++ {
		if yPixCount(img, x) > 0 {
			left = x
			break
		}
	}
	for x := b.Max.X; x > b.Min.X; x-- {
		if yPixCount(img, x-1) > 0 {
			right = x
			break
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		if xPixCount(img, y) > 0 {
			top = y
			break
		}
	}
	for y := b.Max.Y; y > b.Min.Y; y-- {
		if xPixCount(img, y-1) > 0 {
			bottom = y
			break
		}
	}
	if top > min {
		top -= min
	}
	if left > min {
		left -= min
	}
	if bottom+min < b.Max.Y {
		bottom += min
	}
	if right+min < b.Max.Y {
		right += min
	}
	return image.Rect(left, top, right, bottom)
}
