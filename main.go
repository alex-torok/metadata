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
	r := metadata.Repo{
		Root:             fullPath,
		MetadataFilename: "METADATA",
	}

	files, _ := r.MetadataFiles()
	for _, f := range files {
		fmt.Printf("%+v\n", f)
	}

	parsed, err := metadata.ParseAll(files)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", parsed)
}
