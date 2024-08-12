package cmd

import (
	"github.com/spf13/cobra"
)

type CMDI interface {
	Execute() error
}

var rootCmd = &cobra.Command{
	Use: "tensorfuse",
}

func New() CMDI {
	rootCmd.AddCommand(deployCMD)
	return rootCmd
}
