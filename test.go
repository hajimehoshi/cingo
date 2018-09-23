// Copyright 2018 Hajime Hoshi.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hajimehoshi/cingo/cc"
)

var (
	testsuite = flag.String("testsuite", "", "path to the c-testsuite")
)

type Case struct {
	In       []byte
	Expected []byte
}

func run() error {
	if *testsuite == "" {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			return err
		}
		if err := os.Chdir(dir); err != nil {
			return err
		}
		fmt.Println(dir)

		cmd := exec.Command("git", "clone", "https://github.com/c-testsuite/c-testsuite", "--depth", "1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

		if err := os.Chdir("c-testsuite"); err != nil {
			return err
		}
	} else {
		if err := os.Chdir(*testsuite); err != nil {
			return err
		}
		
	}

	if err := os.Chdir(filepath.Join("tests", "single-exec")); err != nil {
		return err
	}

	cases := map[string]*Case{}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		toks := strings.SplitN(info.Name(), ".", 2)
		name := toks[0]
		ext := toks[1]
		switch ext {
		case "c":
			if _, ok := cases[name]; !ok {
				cases[name] = &Case{}
			}
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			cases[name].In = content
		case "c.expected":
			if _, ok := cases[name]; !ok {
				cases[name] = &Case{}
			}
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			cases[name].Expected = content
		}
		return nil
	})

	names := []string{}
	for name, _ := range cases {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		c := cases[name]
		prog, err := cc.Read(name + ".c", bytes.NewReader(c.In))
		if err != nil {
			fmt.Println(err)
			continue
		}
		_ = prog
	}

	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}