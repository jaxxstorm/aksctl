package main

import (
	"fmt"
	"github.com/jaxxstorm/aksctl/cmd/aksctl/create"
	"github.com/jaxxstorm/aksctl/cmd/aksctl/delete"
	"github.com/jaxxstorm/aksctl/cmd/aksctl/version"
	"github.com/jaxxstorm/aksctl/pkg/contract"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	org   string
	debug bool
)

func configureCLI() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:  "aksctl",
		Long: "Create AKS clusters with ease",
	}

	rootCommand.AddCommand(create.Command())
	rootCommand.AddCommand(delete.Command())
	rootCommand.AddCommand(version.Command())

	rootCommand.PersistentFlags().StringVarP(&org, "org", "o", "", "Pulumi org to use for your stack")
	rootCommand.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	viper.BindPFlag("org", rootCommand.PersistentFlags().Lookup("org"))

	return rootCommand
}

func init() {
	log.SetLevel(log.InfoLevel)
	cobra.OnInitialize(initConfig)
}

func initConfig() {

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.aksctl") // adding home directory as first search path
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file: ", viper.ConfigFileUsed())
	}
}

func main() {
	rootCommand := configureCLI()

	if err := rootCommand.Execute(); err != nil {
		contract.IgnoreIoError(fmt.Fprintf(os.Stderr, "%s", err))
		os.Exit(1)
	}
}
