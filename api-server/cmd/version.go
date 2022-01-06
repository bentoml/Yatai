package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bentoml/yatai/api-server/version"
	"github.com/bentoml/yatai/common/command"
)

type VersionOption struct {
}

func (opt *VersionOption) Complete(ctx context.Context, args []string, argsLenAtDash int) error {
	return nil
}

func (opt *VersionOption) Validate(ctx context.Context) error {
	return nil
}

func (opt *VersionOption) Run(ctx context.Context, args []string) error {
	version.Print()
	return nil
}

func getVersionCmd() *cobra.Command {
	var opt VersionOption
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the api-server version information",
		Long:  "",
		RunE:  command.MakeRunE(&opt),
	}
	return cmd
}
