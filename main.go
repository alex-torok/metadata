// Copyright 2017 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"metadata/metadata"

	"github.com/spf13/cobra"
)

var filePath string
var metadataKey string

func do_it(root, filePath, key string) error {
	fullPath, _ := filepath.Abs(os.Args[1])
	tree, err := metadata.TreeFromDir(fullPath, "METADATA")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	val, err := tree.GetMergedValue(filePath, key)
	if err != nil {
		return err
	}

	j, err := metadata.ValueToJson(val)
	if err != nil {
		return err
	}
	fmt.Println(j)
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "meta ROOT FILE METADATA_KEY",
	Short: "Meta is a tool for tracking metadata associated with files in your repo",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return do_it(args[0], args[1], args[2])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
