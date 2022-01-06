package cmd

import (
	"fmt"
	"os"

	"github.com/bentoml/yatai/common/command"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yatai-api-server",
	Short: "",
	Long:  "",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&command.GlobalCommandOption.Debug, "debug", "d", false, "debug mode, output verbose output")
	rootCmd.AddCommand(getServeCmd())
	rootCmd.AddCommand(getVersionCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}
