// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/cingo/cc"
)

var dst = flag.String("dst", "/tmp/c2go", "GOPATH root of destination")

// writeGoFiles writes prog to Go source files in a tree of packages.
func writeGoFiles(cfg *Config, prog *cc.Prog) {
	printers := map[string]*Printer{}
	cfiles := map[string]string{}
	for _, decl := range prog.Decls {
		if decl.GoPackage == "" {
			decl.GoPackage = "other"
		}
		cfile := decl.Span.Start.File
		gofile := strings.TrimSuffix(strings.TrimSuffix(cfile, ".c"), ".h") + ".go"
		gofile = decl.GoPackage + "/" + filepath.Base(gofile)
		cfiles[gofile] = cfile
		p := printers[gofile]
		if p == nil {
			p = new(Printer)
			p.Package = decl.GoPackage
			pkg := path.Base(p.Package)
			if strings.Count(p.Package, "/") == 1 && strings.HasPrefix(p.Package, "cmd/") {
				pkg = "main"
			}
			p.Print("package ", pkg, "\n\n")
			switch p.Package {
			case "cmd/new5g", "cmd/new6g", "cmd/new8g", "cmd/new9g":
				p.Print(`import "cmd/internal/obj"`, "\n")
				p.Print(`import "cmd/internal/gc"`, "\n")

			case "cmd/new5l", "cmd/new6l", "cmd/new8l", "cmd/new9l":
				p.Print(`import "cmd/internal/obj"`, "\n")
				p.Print(`import "cmd/internal/ld"`, "\n")

			case "cmd/internal/gc", "cmd/internal/ld", "cmd/internal/obj/arm", "cmd/internal/obj/ppc64", "cmd/internal/obj/x86", "cmd/internal/obj/amd64":
				p.Print(`import "cmd/internal/obj"`, "\n")
			}

			printers[gofile] = p
		}

		off := len(p.Bytes())
		repl, ok := cfg.replace[decl.Name]
		if !ok {
			repl, ok = cfg.replace[strings.ToLower(decl.Name)]
		}
		if cfg.delete[decl.Name] || cfg.delete[strings.ToLower(decl.Name)] {
			repl, ok = "", true
		}
		if ok {
			// Use replacement text from config but keep surrounding comments.
			p.Print(decl.Comments.Before)
			p.Print(repl)
			p.Print(decl.Comments.Suffix, decl.Comments.After)
		} else {
			p.Print(decl)
		}
		if len(p.Bytes()) > off {
			p.Print(Newline)
			p.Print(Newline)
		}
	}

	for gofile, p := range printers {
		dstfile := filepath.Join(*dst+"/src", gofile)
		os.MkdirAll(filepath.Dir(dstfile), 0777)
		buf := p.Bytes()

		// Not entirely sure why these lines get broken.
		buf = bytes.Replace(buf, []byte("\n,"), []byte(","), -1)
		buf = bytes.Replace(buf, []byte("\n {"), []byte(" {"), -1)

		buf1, err := format.Source(buf)
		if err != nil {
			// Scream because it invalidates diffs.
			log.Printf("ERROR formatting %s: %v", gofile, err)
		}
		if err == nil {
			buf = buf1
		}

		// Not sure where these blank lines come from.
		buf = bytes.Replace(buf, []byte("{\n\n"), []byte("{\n"), -1)

		for i, d := range cfg.diffs {
			if bytes.Contains(buf, d.before) {
				buf = bytes.Replace(buf, d.before, d.after, -1)
				cfg.diffs[i].used++
			}
		}

		if err := ioutil.WriteFile(dstfile, buf, 0666); err != nil {
			log.Print(err)
		}
	}
}
