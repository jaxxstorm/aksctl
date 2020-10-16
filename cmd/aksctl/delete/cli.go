package delete

import (
	"github.com/jaxxstorm/aksctl/cmd/aksctl/delete/cluster"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "delete",
		Short: "Create resources",
		Long:  "Commands that create resources in AKS",
	}

	command.AddCommand(cluster.Command())


	return command
}
