package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alex-torok/metadata/metadata"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get some metadata ma bois",
	Args:  cobra.ExactArgs(3),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	repoRoot, _ := filepath.Abs(os.Args[0])
	file := args[1]
	key := args[2]

	tree, err := metadata.NewEagerTree(repoRoot, "METADATA")
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		os.Exit(1)
	}

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
	rootCmd.AddCommand(getCmd)
}
