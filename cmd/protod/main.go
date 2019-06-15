package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/dennwc/protod"
)

var (
	fOut = flag.String("out", ".", "output directory")
	fRaw = flag.Bool("raw", false, "dump raw proto file descriptors as well")
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := filepath.Base(os.Args[0])
	if len(commit) > 8 {
		commit = commit[:8]
	}
	fmt.Printf("%s %v, commit %v, built at %v\n", app, version, commit, date)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [--out dir] <files...>\n", app)
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "expected at least one argument\n")
		flag.Usage()
		os.Exit(2)
	}
	for _, fname := range args {
		if err := run(*fOut, fname, *fRaw); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", fname, err)
			os.Exit(1)
		}
	}
}

func run(out, file string, raw bool) error {
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	var r io.Reader
	if file == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	return protod.Extract(r, func(data []byte, fd *protod.FileDescriptorProto) error {
		name := fd.GetName()
		fmt.Println(name)
		name = path.Base(name)
		if name == "" {
			name = "unknown.proto"
		}
		if raw {
			if err := ioutil.WriteFile(filepath.Join(out, name+".protoc"), data, 0644); err != nil {
				return err
			}
		}
		f, err := os.Create(filepath.Join(out, name))
		if err != nil {
			return err
		}
		defer f.Close()
		if err = protod.Render(f, fd); err != nil {
			return err
		}
		return f.Close()
	})
}
