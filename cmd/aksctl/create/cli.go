package create

import (
	"github.com/jaxxstorm/aksctl/cmd/aksctl/create/cluster"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "Create resources",
		Long:  "Commands that create resources in AKS",
	}

	command.AddCommand(cluster.Command())


	return command
}
