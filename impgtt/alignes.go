package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"git.sr.ht/~flobar/lev"
	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(alCmd)
	alCmd.Flags().StringVar(&alArgs.imgext, "imgext", ".bin.png", "set file extension for the image files")
	alCmd.Flags().StringVar(&alArgs.ocrext, "ocrext", ".pred.txt", "set file extension for input ocr files")
	alCmd.Flags().StringVar(&alArgs.gtext, "gtext", ".gt.txt", "set file extension for output gt files")
	alCmd.Flags().StringVar(&alArgs.rectext, "rectext", ".rect.txt", "set file extension for rectangle files")
	alCmd.Flags().StringVar(&alArgs.anglext, "anglext", ".angle.txt", "set file extension for angle files")
}

var alCmd = &cobra.Command{
	Use:   "alignes [JSON...]",
	Short: "Alignes GT lines from metadata files with their image files",
	Args:  cobra.MinimumNArgs(1),
	Run:   alignes,
}

var alArgs = struct {
	imgext, ocrext, gtext, rectext, anglext string
}{}

func alignes(_ *cobra.Command, args []string) {
	for _, name := range args {
		chk(alAlign(name))
	}
}

func alAlign(name string) error {
	logf("aligning %s", name)
	r, err := readRegion(name)
	if err != nil {
		return fmt.Errorf("align: %v", err)
	}
	dir := filepath.Join(filepath.Dir(name), r.Dir)
	if !exists(dir) {
		log.Printf("warning: directory %s does not exit; skipping", dir)
		return nil
	}
	files, ocr, rects, err := alGatherOCRFiles(dir)
	if err != nil {
		return fmt.Errorf("align: %v", err)
	}
	gt := strings.Split(r.Text, "\n")
	m, trace := alAlignLines(gt, ocr)
	m.print(os.Stdout, gt, ocr)
	log.Printf("trace: %s", trace)
	var snippets, waste []snippet
	i, j := 0, 0
	for _, t := range trace {
		switch t {
		case '#':
			base := filepath.Join(filepath.Base(dir), filepath.Base(files[i]))
			snippets = append(snippets, snippet{
				BaseName:    base,
				Image:       base + alArgs.imgext,
				OCR:         ocr[i],
				GT:          gt[j],
				Distance:    lev.Distance(ocr[i], gt[j]),
				Coordinates: rects[i],
			})
			i++
			j++
		case 'd':
			base := filepath.Join(filepath.Base(dir), filepath.Base(files[i]))
			waste = append(waste, snippet{
				BaseName:    base,
				Image:       base + alArgs.imgext,
				OCR:         ocr[i],
				Distance:    len(ocr[i]),
				Coordinates: rects[i],
			})
			i++
		case 'i':
			j++
		default:
			panic("bad trace")
		}
	}
	r.Snippets = snippets
	r.Waste = waste
	a, err := alReadAngle(name[:len(name)-5] + alArgs.anglext)
	if err != nil {
		return err
	}
	r.SkewAngle = a
	for i := range snippets {
		log.Printf("%s GT:  %s", snippets[i].BaseName, snippets[i].GT)
		log.Printf("%s OCR: %s", snippets[i].BaseName, snippets[i].OCR)
		ofile := filepath.Join(filepath.Dir(dir), snippets[i].BaseName+alArgs.gtext)
		if err := ioutil.WriteFile(ofile, []byte(snippets[i].GT+"\n"), 0666); err != nil {
			return err
		}
	}
	return r.write(name)
}

func alGatherOCRFiles(dir string) ([]string, []string, []polygon, error) {
	fail := func(err error) ([]string, []string, []polygon, error) {
		return nil, nil, nil, fmt.Errorf("gather ocr files %s: %v", dir, err)
	}
	var files []string
	err := filepath.Walk(dir, func(name string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(name, alArgs.ocrext) {
			return nil
		}
		files = append(files, name[0:len(name)-len(alArgs.ocrext)])
		return nil
	})
	if err != nil {
		return fail(err)
	}
	ocr := make([]string, len(files))
	rects := make([]polygon, len(files))
	for i := range files {
		line, err := alReadOCRFile(files[i] + alArgs.ocrext)
		if err != nil {
			return fail(err)
		}
		p, err := alReadRectangle(files[i] + alArgs.rectext)
		if err != nil {
			return fail(err)
		}
		rects[i] = p
		ocr[i] = line
	}
	return files, ocr, rects, nil
}

func exists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

type mat struct {
	r, c int
	tab  []int
}

func newmat(r, c int) *mat {
	return &mat{r: r, c: c, tab: make([]int, r*c)}
}

func (m *mat) at(i, j int) int {
	idx := i*m.c + j
	if idx >= len(m.tab) {
		return math.MaxInt32
	}
	return m.tab[i*m.c+j]
}

func (m *mat) set(i, j, val int) int {
	m.tab[i*m.c+j] = val
	return val
}

func (m *mat) trace() string {
	var x []byte
	for i, j := m.r-1, m.c-1; i > 0 || j > 0; {
		switch m.at(i, j) {
		case 0:
			i--
			j--
			x = append(x, '#')
		case 1:
			i--
			x = append(x, 'd')
		case 2:
			j--
			x = append(x, 'i')
		default:
			panic("bad entry")
		}
	}
	for i, j := 0, len(x); i < j; i, j = i+1, j-1 {
		x[i], x[j-1] = x[j-1], x[i]
	}
	return string(x)
}

func (m *mat) print(out io.Writer, gt, ocr []string) {
	max := 0
	for i := range ocr {
		if len(tostr(ocr[i], 10)) > max {
			max = len(tostr(ocr[i], 10))
		}
	}
	for i := range gt {
		if len(tostr(gt[i], 10)) > max {
			max = len(tostr(gt[i], 10))
		}
	}
	var w tabwriter.Writer
	w.Init(out, 1, 8, 1, ' ', 0)
	defer w.Flush()
	fmt.Fprint(&w, " \t ")
	for i := range gt {
		fmt.Fprintf(&w, "\t%s", tostr(gt[i], 10))
	}
	fmt.Fprintln(&w)
	for i := 0; i < m.r; i++ {
		if i == 0 {
			fmt.Fprint(&w, " ")
		} else {
			fmt.Fprintf(&w, "%s", tostr(ocr[i-1], 10))
		}
		for j := 0; j < m.c; j++ {
			fmt.Fprintf(&w, "\t%d", m.at(i, j))
		}
		fmt.Fprintln(&w)
	}
}

func tostr(str string, n int) string {
	if len(str) > n {
		return str[:n-3] + "..."
	}
	return str
}

func alAlignLines(gt, ocr []string) (*mat, string) {
	m := newmat(len(ocr)+1, len(gt)+1)
	t := newmat(len(ocr)+1, len(gt)+1)
	for i := range ocr {
		m.set(i+1, 0, len(ocr[i])+m.at(i, 0))
		t.set(i+1, 0, 1)
	}
	for i := range gt {
		m.set(0, i+1, len(gt[i])+m.at(0, i))
		t.set(0, i+1, 2)
	}
	for i := 1; i < m.r; i++ {
		for j := 1; j < m.c; j++ {
			a := m.at(i-1, j-1) + lev.Distance(gt[j-1], ocr[i-1])
			b := m.at(i-1, j) + len(ocr[i-1])
			c := m.at(i, j-1) + len(gt[j-1])
			min, pos := min(a, b, c)
			m.set(i, j, min)
			t.set(i, j, pos)
		}
	}
	return m, t.trace()
}

func alReadLine(name string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("read line %s: %v", name, err)
	}
	in, err := os.Open(name)
	if err != nil {
		return fail(err)
	}
	defer in.Close()
	line, err := ioutil.ReadAll(in)
	if err != nil {
		return fail(err)
	}
	return string(line), nil
}

func alReadOCRFile(name string) (string, error) {
	ocr, err := alReadLine(name)
	if err != nil {
		return "", fmt.Errorf("read ocr file %s: %v", name, err)
	}
	return ocr, nil
}

func alReadRectangle(name string) (polygon, error) {
	fail := func(err error) (polygon, error) {
		return nil, fmt.Errorf("read rectangle %s: %v", name, err)
	}
	line, err := alReadLine(name)
	if err != nil {
		return fail(err)
	}
	p, err := makePolygon(string(line))
	if err != nil {
		return fail(err)
	}
	return p, nil
}

func alReadAngle(name string) (float32, error) {
	fail := func(err error) (float32, error) {
		return 0.0, fmt.Errorf("read angle %s: %v", name, err)
	}
	line, err := alReadLine(name)
	if err != nil {
		return fail(err)
	}
	var a float32
	if _, err := fmt.Sscanf(line, "%f", &a); err != nil {
		return fail(err)
	}
	return a, nil
}

func min(xs ...int) (int, int) {
	min := xs[0]
	arg := 0
	for i, x := range xs[1:] {
		if x < min {
			min = x
			arg = i + 1
		}
	}
	return min, arg
}
