package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alex-torok/metadata/metadata"
	"github.com/spf13/cobra"
	"go.starlark.net/starlark"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get some metadata ma bois",
}

var getMultiCmd = &cobra.Command{
	Use:   "multi ROOT KEY FILE...",
	Short: "Get metadata values for multiple files",
	Args:  cobra.MinimumNArgs(3),
	RunE:  runGetMulti,
}

func runGetMulti(cmd *cobra.Command, args []string) error {
	repoRoot, _ := filepath.Abs(args[0])
	key := args[1]
	files := args[2:]

	tree, err := metadata.NewEagerTree(repoRoot, "METADATA")
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		os.Exit(1)
	}

	allMetadata := make(map[string]starlark.Value)
	for _, file := range files {
		val, err := tree.GetMergedValue(file, key)
		if err != nil {
			switch e := err.(type) {
			case metadata.NoMetadataFoundError:
				allMetadata[file] = starlark.None
			default:
				return e
			}
		} else {
			allMetadata[file] = val
		}
	}

	j, err := metadata.FileMapToJson(allMetadata)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), j)
	return nil
}

var getOneCmd = &cobra.Command{
	Use:   "one ROOT KEY FILE",
	Short: "Get a metadata value for one file",
	Args:  cobra.ExactArgs(3),
	RunE:  runGetOne,
}

func runGetOne(cmd *cobra.Command, args []string) error {
	repoRoot, _ := filepath.Abs(args[0])
	key := args[1]
	file := args[2]

	tree, err := metadata.NewEagerTree(repoRoot, "METADATA")
	if err != nil {
		return err
	}

	fmt.Println(repoRoot)
	val, err := tree.GetMergedValue(file, key)
	if err != nil {
		return err
	}

	j, err := metadata.ValueToJson(val)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), j)
	return nil
}

func init() {
	getCmd.AddCommand(getOneCmd)
	getCmd.AddCommand(getMultiCmd)
	rootCmd.AddCommand(getCmd)
}
