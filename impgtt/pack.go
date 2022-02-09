package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const indexStart = 65537

func init() {
	root.AddCommand(pkgCmd)
	pkgCmd.Flags().StringVar(&pkgArgs.imgext, "imgext", ".bin.png", "set file extension for the image files")
	pkgCmd.Flags().StringVar(&pkgArgs.gtext, "gtext", ".gt.txt", "set file extension for output gt files")
	pkgCmd.Flags().IntVar(&pkgArgs.start, "start", indexStart, "set start index for line snippets")
}

var pkgArgs = struct {
	imgext, gtext string
	start         int
}{}

var pkgCmd = &cobra.Command{
	Use:   "pack [DIR]... ODIR",
	Short: "Pack segmented directories",
	Args:  cobra.MinimumNArgs(1),
	Run:   pkg,
}

func pkg(_ *cobra.Command, args []string) {
	odir := args[len(args)-1]
	for _, dir := range args[0 : len(args)-1] {
		pkgDir(odir, dir)
	}
}

func pkgDir(odir, dir string) {
	regions, err := gatherRegionsMap(dir)
	chk(err)
	dst := filepath.Join(odir, filepath.Base(dir))

	for pagenum, rs := range regions {
		// Regions cannot not be empty; we skip empty regions anyway.
		if len(rs) == 0 {
			continue
		}
		page := fmt.Sprintf("%d", pagenum)
		pagedir := filepath.Join(dst, page)
		chk(os.MkdirAll(pagedir, 0755))
		chk(copy(filepath.Join(dst, page+pkgArgs.imgext), filepath.Join(dir, rs[0].PageImage)))

		sort.Slice(rs, func(i, j int) bool {
			return rs[i].region() < rs[j].region()
		})
		index := pkgArgs.start
		for _, r := range rs {
			for _, s := range r.Snippets {
				srcf := filepath.Join(dir, s.BaseName)
				dstf := filepath.Join(pagedir, fmt.Sprintf("%06x", index))
				chk(copy(dstf+pkgArgs.gtext, srcf+pkgArgs.gtext))
				chk(copy(dstf+pkgArgs.imgext, srcf+pkgArgs.imgext))
				index++
			}
		}
	}
}

func copy(dst, src string) error {
	fail := func(err error) error {
		return fmt.Errorf("copy %s to %s: %v", src, dst, err)
	}
	logf("copy %s -> %s", src, dst)

	r, err := os.Open(src)
	if err != nil {
		return fail(err)
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return fail(err)
	}
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return fail(err)
	}
	return nil
}

func gatherRegionsMap(dir string) (map[int][]region, error) {
	regions := make(map[int][]region)
	err := filepath.Walk(dir, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(p, ".json") {
			return nil
		}
		r, err := readRegion(p)
		if err != nil {
			return err
		}
		regions[r.page()] = append(regions[r.page()], r)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load regions %s: %v", dir, err)
	}
	return regions, nil
}
