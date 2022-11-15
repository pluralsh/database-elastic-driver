package main

import (
	"context"
	"flag"
	"strings"

	"github.com/pluralsh/database-eleastic-driver/pkg/driver"
	"github.com/pluralsh/database-eleastic-driver/pkg/elastic"
	"github.com/pluralsh/database-interface-controller/pkg/provisioner"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog"
)

const provisionerName = "elastic.database.plural.sh"

var (
	driverAddress = "unix:///var/lib/database/database.sock"
	dbUser        = ""
	dbPassword    = ""
	dbAddress     = ""
)

var cmd = &cobra.Command{
	Use:           "elastic-database-driver",
	Short:         "K8s database driver for Elastic database",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd.Context(), args)
	},
	DisableFlagsInUseLine: true,
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	flag.Set("alsologtostderr", "true")
	kflags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(kflags)

	persistentFlags := cmd.PersistentFlags()
	persistentFlags.AddGoFlagSet(kflags)

	stringFlag := persistentFlags.StringVarP

	stringFlag(&driverAddress,
		"driver-addr",
		"d",
		driverAddress,
		"path to unix domain socket where driver should listen")
	stringFlag(&dbUser,
		"db-user",
		"",
		dbUser,
		"elastic user")
	stringFlag(&dbPassword,
		"db-password",
		"",
		dbPassword,
		"elastic password")
	stringFlag(&dbAddress,
		"db-address",
		"",
		dbAddress,
		"elastic address")

	viper.BindPFlags(cmd.PersistentFlags())
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.PersistentFlags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}

func run(ctx context.Context, args []string) error {
	elasticDB := &elastic.Elastic{
		User:     dbUser,
		Password: dbPassword,
		Address:  dbAddress,
	}
	klog.Info("\nuser:", dbUser, "\naddress:", dbAddress, "\n")
	identityServer, databaseProvisioner := driver.NewDriver(provisionerName, elasticDB)
	server, err := provisioner.NewDefaultProvisionerServer(driverAddress,
		identityServer,
		databaseProvisioner)
	if err != nil {
		klog.Errorf("Failed to create provisioner server %v", err)
		return err
	}
	klog.Info("Starting Elastic provisioner")
	return server.Run(ctx)
}
