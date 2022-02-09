package main // import "github.com/cisocrgroup/IMPACTGT/impgtt"

import (
	"log"

	"github.com/spf13/cobra"
)

var debug bool

var root = &cobra.Command{
	Use:   "impgtt",
	Short: "Tools to generate gt from IMPACT ground-truth files",
}

func init() {
	root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debugging messages")
}

func main() {
	root.Execute()
}

func logf(f string, args ...interface{}) {
	if debug {
		log.Printf(f, args...)
	}
}

func chk(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
