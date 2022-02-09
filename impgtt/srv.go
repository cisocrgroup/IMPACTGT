package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(srvCmd)
	srvCmd.Flags().StringVarP(&srvArgs.host, "host", "H", ":4242", "set host")
}

var srvCmd = &cobra.Command{
	Use:   "srv DIR",
	Short: "Start a server to inspect the line snippets",
	Args:  cobra.ExactArgs(1),
	Run:   srv,
}

var srvArgs = struct {
	host string
}{}

//go:embed testdata/html/region.tmpl.html
var tmpl []byte

//go:embed testdata/img/app.jpeg
var favicon []byte

var funcs = template.FuncMap{
	"normalize": normalize,
	"split":     func(str string) []string { return strings.Split(str, "\n") },
}

func srv(_ *cobra.Command, args []string) {
	t, err := template.New("regions").Funcs(funcs).Parse(string(tmpl))
	chk(err)
	srv := server{dir: args[0], tmpl: t}
	chk(srv.loadRegions())
	logf("serving %d regions on %s", len(srv.regions), srvArgs.host)
	chk(http.ListenAndServe(srvArgs.host, &srv))
}

type server struct {
	regions []region
	tmpl    *template.Template
	dir     string
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logf("serving host %s: %s", r.RemoteAddr, r.URL)
	if strings.HasSuffix(r.URL.String(), ".png") {
		s.serveImg(w, r)
		return
	}
	if strings.HasSuffix(r.URL.String(), "favicon.ico") {
		s.serveFavicon(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "region") {
		s.serveRegion(w, r)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func (s *server) serveRegion(w http.ResponseWriter, r *http.Request) {
	// If region id is given use it to find the region otherwise
	// use the index directly; id parameter take precedence over
	// any index parameter.
	var i int
	id := r.URL.Query().Get("id")
	if id != "" {
		i = s.regionIndex(id)
	} else {
		i, _ = strconv.Atoi(r.URL.Query().Get("index"))
	}
	// Server region at index i.
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	n, p := i+1, i-1
	if n >= len(s.regions) {
		n = 0
	}
	if p < 0 {
		p = len(s.regions) - 1
	}
	err := s.tmpl.Execute(w, struct {
		Next, Prev, Index, NRegions int
		Dir                         string
		Region                      region
	}{n, p, i + 1, len(s.regions), filepath.Base(s.dir), s.regions[i]})
	if err != nil {
		logf("error serving template: %v", err)
	}
}

func (s *server) regionIndex(id string) int {
	for i := range s.regions {
		if id == s.regions[i].Dir {
			return i
		}
	}
	return 0
}

func (s *server) serveImg(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.dir, r.URL.String()))
}

func (s *server) serveFavicon(w http.ResponseWriter, r *http.Request) {
	buf := bytes.NewBuffer(favicon)
	w.Header().Add("Content-Type", "image/jpeg")
	w.Header().Add("Cache-Control", "public")
	_, err := io.Copy(w, buf)
	if err != nil {
		logf("error serving favicon: %v", err)
	}
}

func (s *server) loadRegions() error {
	err := filepath.Walk(s.dir, func(p string, info fs.FileInfo, err error) error {
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
		s.regions = append(s.regions, r)
		return nil
	})
	if err != nil {
		return fmt.Errorf("load regions %s: %v", s.dir, err)
	}
	s.sortRegions()
	return nil
}

func (s *server) sortRegions() {
	sort.Slice(s.regions, func(i, j int) bool {
		pi, pj := s.regions[i].page(), s.regions[j].page()
		ri, rj := s.regions[i].region(), s.regions[j].region()
		if pi == pj {
			return ri < rj
		}
		return pi < pj
	})
}
