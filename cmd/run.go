package cmd

import (
	ddgchat "github.com/nerdneilsfield/go-template/ddg-chat"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "run",
		Short:        "run ddg-chat",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := ddgchat.LoadConfig(args[0])
			if err != nil {
				logger.Error("failed to load config", zap.Error(err))
				return err
			}

			return ddgchat.RunServer(config)
		},
	}
}
