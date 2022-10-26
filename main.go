package main

import (
	"github.com/mirefly/go-script/gitsyncer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-script",
	Short: "go-script is a collection of miscellaneous tools",
}

func init() {
	rootCmd.AddCommand(gitsyncer.Cmd)
}

func main() {
	rootCmd.Execute()
}
