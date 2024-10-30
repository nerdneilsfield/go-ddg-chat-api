package cmd

import (
	"os"

	ddgchat "github.com/nerdneilsfield/go-template/ddg-chat"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newRunCmd(version string, buildTime string, gitCommit string) *cobra.Command {
	return &cobra.Command{
		Use:          "run",
		Short:        "run ddg-chat",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ddgchat.LoadConfig(args[0])
			if err != nil {
				logger.Error("failed to load config", zap.Error(err))
				os.Exit(1)
			}

			ddgchat.RunServer(config)
		},
	}
}
