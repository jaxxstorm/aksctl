package cluster

import (
	"context"
	"fmt"
	containerservice "github.com/pulumi/pulumi-azure-nextgen/sdk/go/azure/containerservice/latest"
	resources "github.com/pulumi/pulumi-azure-nextgen/sdk/go/azure/resources/latest"
	"github.com/pulumi/pulumi-azuread/sdk/v2/go/azuread"
	"github.com/pulumi/pulumi-random/sdk/v2/go/random"
	"github.com/pulumi/pulumi-tls/sdk/v2/go/tls"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v2/go/x/auto"
	"github.com/pulumi/pulumi/sdk/v2/go/x/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v2/go/x/auto/optup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	name           string
	region         string
	stack          string
	project        string
	dryrun         bool
	clusterVersion string
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "cluster",
		Short: "Create a cluster",
		Long:  "Create an AKS Cluster",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := context.Background()
			org := viper.GetString("org")

			if org == "" {
				return fmt.Errorf("must specify pulumi org via flag or config file")
			}

			projName := name
			stackName := auto.FullyQualifiedStackName(org, project, stack)

			stack, err := auto.UpsertStackInlineSource(ctx, stackName, projName, createClusterFunc)

			if err != nil {
				return fmt.Errorf("failed to create or select stack: %v\n", err)
			}

			workspace := stack.Workspace()
			err = workspace.InstallPlugin(ctx, "azure-nextgen", "v0.2.2")
			if err != nil {
				return fmt.Errorf("error installing azure nextgen plugin: %v\n", err)
			}
			err = workspace.InstallPlugin(ctx, "azuread", "v2.5.1")
			if err != nil {
				return fmt.Errorf("error installing azuread plugin: %v\n", err)
			}
			err = workspace.InstallPlugin(ctx, "random", "v2.3.1")
			if err != nil {
				return fmt.Errorf("error installing random plugin: %v\n", err)
			}
			err = workspace.InstallPlugin(ctx, "tls", "v2.3.0")
			if err != nil {
				return fmt.Errorf("error installing tls plugin: %v\n", err)
			}

			if dryrun {
				_, err = stack.Preview(ctx, optpreview.Message("Running aksctl dryrun"))
				if err != nil {
					return fmt.Errorf("error creating stack: %v\n", err)
				}
			} else {
				// wire up our update to stream progress to stdout
				stdoutStreamer := optup.ProgressStreams(os.Stdout)
				_, err = stack.Up(ctx, stdoutStreamer)
			}
			return nil
		},
	}

	f := command.Flags()
	f.BoolVarP(&dryrun, "dry-run", "d", false, "Preview changes, dry-run mode")
	f.StringVarP(&name, "name", "n", "", "Name to give to your cluster")
	f.StringVarP(&project, "project", "p", "", "Pulumi project to use")
	f.StringVarP(&region, "region", "r", "westus", "Azure region to deploy to")
	f.StringVarP(&stack, "stack", "s", "", "Pulumi stack to create or use")
	f.StringVar(&clusterVersion, "cluster-version", "1.19.0", "AKS version to deploy")

	cobra.MarkFlagRequired(f, "name")
	cobra.MarkFlagRequired(f, "project")
	cobra.MarkFlagRequired(f, "stack")

	return command
}

// createClusterFunc defines the Pulumi program that creates the cluster
var createClusterFunc = func(ctx *pulumi.Context) error {

	// Create a resource group to store everything in
	resourceGroup, err := resources.NewResourceGroup(ctx, fmt.Sprintf("%s-aksctl-resourceGroup", name), &resources.ResourceGroupArgs{
		ResourceGroupName: pulumi.String(name),
		Location:          pulumi.String(region),
	})
	if err != nil {
		return fmt.Errorf("error creating resource group: %v", err)
	}

	// Generate a random password.
	password, err := random.NewRandomPassword(ctx, fmt.Sprintf("%s-aksctl-password", name), &random.RandomPasswordArgs{
		Length:  pulumi.Int(20),
		Special: pulumi.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("error creating resource random password: %v", err)
	}

	// Create an AD service principal.
	adApp, err := azuread.NewApplication(ctx, fmt.Sprintf("%s-aksctl-application", name), &azuread.ApplicationArgs{
		Name: pulumi.String(name),
	})
	if err != nil {
		return fmt.Errorf("error creating resource azuread application: %v", err)
	}

	adSp, err := azuread.NewServicePrincipal(ctx, fmt.Sprintf("%s-aksctl-servicePrincipal", name), &azuread.ServicePrincipalArgs{
		ApplicationId: adApp.ApplicationId,
	})
	if err != nil {
		return fmt.Errorf("error creating resource azuread service principal: %v", err)
	}

	// Create the Service Principal Password.
	adSpPassword, err := azuread.NewServicePrincipalPassword(ctx, fmt.Sprintf("%s-aksctl-servicePrincipal", name), &azuread.ServicePrincipalPasswordArgs{
		ServicePrincipalId: adSp.ID(),
		Value:              password.Result,
		EndDate:            pulumi.String("2099-01-01T00:00:00Z"),
	})
	if err != nil {
		return err
	}

	// Generate an SSH key.
	sshArgs := tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
		RsaBits:   pulumi.Int(4096),
	}
	sshKey, err := tls.NewPrivateKey(ctx, fmt.Sprintf("%s-aksctl-sshKey", name), &sshArgs)
	if err != nil {
		return fmt.Errorf("error creating resource SSH key: %v", err)
	}

	// Create the Azure Kubernetes Service cluster.
	_, err = containerservice.NewManagedCluster(ctx, fmt.Sprintf("%s-aksctl-cluster", name), &containerservice.ManagedClusterArgs{
		ResourceName:      pulumi.String(name),
		Location:          pulumi.String("WestUS"),
		DnsPrefix:         pulumi.String(fmt.Sprintf("%s-%s-kube", name, stack)),
		ResourceGroupName: resourceGroup.Name,
		AgentPoolProfiles: containerservice.ManagedClusterAgentPoolProfileArray{
			&containerservice.ManagedClusterAgentPoolProfileArgs{
				Name:         pulumi.String(name),
				Mode:         pulumi.String("System"),
				OsDiskSizeGB: pulumi.Int(30),
				Count:        pulumi.Int(3),
				VmSize:       pulumi.String("Standard_DS2_v2"),
				OsType:       pulumi.String("Linux"),
			},
		},
		LinuxProfile: &containerservice.ContainerServiceLinuxProfileArgs{
			AdminUsername: pulumi.String(name),
			Ssh: containerservice.ContainerServiceSshConfigurationArgs{
				PublicKeys: containerservice.ContainerServiceSshPublicKeyArray{
					containerservice.ContainerServiceSshPublicKeyArgs{
						KeyData: sshKey.PublicKeyOpenssh,
					},
				},
			},
		},
		ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfileArgs{
			ClientId: adApp.ApplicationId,
			Secret:   adSpPassword.Value,
		},
		KubernetesVersion: pulumi.String(clusterVersion),
	})

	return nil
}
