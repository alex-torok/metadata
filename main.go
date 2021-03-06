// Copyright 2017 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"metadata/metadata"
)

func main() {
	fullPath, _ := filepath.Abs(os.Args[1])
	tree, err := metadata.TreeFromDir(fullPath, "METADATA")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	p := func(path string, key string) {
		val, _ := tree.GetClosestValue(path, key)
		fmt.Printf("%s (%s): %d\n", key, path, val)
	}

	p("someFile.txt", "cool factor")
	p("one/other/someFile.txt", "cool factor")

	p("someFile.txt", "minimum_coverage")
	p("one/other/someFile.txt", "minimum_coverage")
}
