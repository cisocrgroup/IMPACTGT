package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type region struct {
	Snippets, Waste                          []snippet
	Coordinates                              polygon
	Attrs                                    map[string]string
	Dir, Image, Text, GT, PageImage, PageXML string
	SkewAngle                                float32
	Index                                    int
}

func readRegion(name string) (region, error) {
	fail := func(err error) (region, error) {
		return region{}, fmt.Errorf("read region %s: %v", name, err)
	}
	in, err := os.Open(name)
	if err != nil {
		return fail(err)
	}
	defer in.Close()
	var r region
	if err := json.NewDecoder(in).Decode(&r); err != nil {
		return fail(err)
	}
	return r, nil
}

func (r *region) write(name string) (err error) {
	out, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("write json %s: %v", name, err)
	}
	defer func() {
		if err != nil {
			err = out.Close()
		}
	}()
	if err := json.NewEncoder(out).Encode(r); err != nil {
		return fmt.Errorf("write %s: encode: %v", name, err)
	}
	return nil
}

func (r *region) putAttr(k, v string) {
	if r.Attrs == nil {
		r.Attrs = make(map[string]string)
	}
	r.Attrs[k] = v
}

func (r *region) page() int {
	page := strings.Split(filepath.Base(r.Dir), "_")[0]
	no, _ := strconv.Atoi(page)
	return no
}

func (r *region) region() int {
	region := strings.Split(filepath.Base(r.Dir), "_")[1]
	pos := strings.IndexFunc(region, unicode.IsDigit)
	if pos < 0 {
		return 0
	}
	no, err := strconv.Atoi(region[pos:])
	if err != nil {
		return 0
	}
	return no
}

type snippet struct {
	Coordinates              polygon
	BaseName, Image, GT, OCR string
	Distance                 int
}

func (s snippet) num() int {
	name := filepath.Base(s.BaseName)
	var num int
	fmt.Sscanf(name, "%x", &num)
	return num
}

func (s snippet) relev() float64 {
	return float64(s.Distance) / float64(len(s.OCR))
}
