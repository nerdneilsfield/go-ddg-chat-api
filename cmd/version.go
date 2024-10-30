package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string, buildTime string, gitCommit string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "go-template version",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("go-template")
			fmt.Println("A template for go projects.")
			fmt.Println("Author: dengqi935@gmail.com")
			fmt.Println("Github: https://github.com/nerdneilsfield/go-template")
			fmt.Fprintf(cmd.OutOrStdout(), "go-template: %s\n", version)
			fmt.Fprintf(cmd.OutOrStdout(), "buildTime: %s\n", buildTime)
			fmt.Fprintf(cmd.OutOrStdout(), "gitCommit: %s\n", gitCommit)
			fmt.Fprintf(cmd.OutOrStdout(), "goVersion: %s\n", runtime.Version())
		},
	}
}
