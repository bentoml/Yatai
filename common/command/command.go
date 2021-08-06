package command

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GlobalCommandOption = struct {
	Debug bool
}{}

type ICommandOption interface {
	Complete(ctx context.Context, args []string, argsLenAtDash int) error
	Validate(ctx context.Context) error
	Run(ctx context.Context, args []string) error
}

func MakeRunE(opt ICommandOption) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if GlobalCommandOption.Debug {
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}
		argsLenAtDash := cmd.ArgsLenAtDash()
		ctx := cmd.Context()
		err := opt.Complete(ctx, args, argsLenAtDash)
		if err != nil {
			return err
		}
		err = opt.Validate(ctx)
		if err != nil {
			return err
		}
		return opt.Run(ctx, args)
	}
}
