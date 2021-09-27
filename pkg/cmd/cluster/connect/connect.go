package connect

import (
	"context"

	"github.com/redhat-developer/app-services-cli/internal/build"
	"github.com/redhat-developer/app-services-cli/internal/config"
	"github.com/redhat-developer/app-services-cli/pkg/api"
	"github.com/redhat-developer/app-services-cli/pkg/cluster"
	"github.com/redhat-developer/app-services-cli/pkg/cluster/kafkaservice"
	"github.com/redhat-developer/app-services-cli/pkg/cluster/registryservice"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/factory"
	"github.com/redhat-developer/app-services-cli/pkg/connection"
	"github.com/redhat-developer/app-services-cli/pkg/iostreams"
	"github.com/redhat-developer/app-services-cli/pkg/kafka"
	"github.com/redhat-developer/app-services-cli/pkg/localize"
	"github.com/redhat-developer/app-services-cli/pkg/logging"
	"github.com/redhat-developer/app-services-cli/pkg/serviceregistry"
	"github.com/spf13/cobra"
)

type options struct {
	Config     config.IConfig
	Connection func(connectionCfg *connection.Config) (connection.Connection, error)
	Logger     logging.Logger
	IO         *iostreams.IOStreams
	localizer  localize.Localizer
	Context    context.Context

	kubeconfigLocation string
	namespace          string

	offlineAccessToken      string
	forceCreationWithoutAsk bool
	ignoreContext           bool
	serviceID               string
	serviceType             string
}

func NewConnectCommand(f *factory.Factory) *cobra.Command {
	opts := &options{
		Config:     f.Config,
		Connection: f.Connection,
		Logger:     f.Logger,
		IO:         f.IOStreams,
		localizer:  f.Localizer,
		Context:    f.Context,
	}

	cmd := &cobra.Command{
		Use:     "connect",
		Short:   opts.localizer.MustLocalize("cluster.connect.cmd.shortDescription"),
		Long:    opts.localizer.MustLocalize("cluster.connect.cmd.longDescription"),
		Example: opts.localizer.MustLocalize("cluster.connect.cmd.example"),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.ignoreContext == true && !opts.IO.CanPrompt() {
				return opts.localizer.MustLocalizeError("flag.error.requiredWhenNonInteractive", localize.NewEntry("Flag", "ignore-context"))
			}
			return runConnect(opts)
		},
	}

	cmd.Flags().StringVar(&opts.kubeconfigLocation, "kubeconfig", "", opts.localizer.MustLocalize("cluster.common.flag.kubeconfig.description"))
	cmd.Flags().StringVar(&opts.offlineAccessToken, "token", "", opts.localizer.MustLocalize("cluster.common.flag.offline.token.description", localize.NewEntry("OfflineTokenURL", build.OfflineTokenURL)))
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "", opts.localizer.MustLocalize("cluster.common.flag.namespace.description"))
	cmd.Flags().BoolVarP(&opts.forceCreationWithoutAsk, "yes", "y", false, opts.localizer.MustLocalize("cluster.common.flag.yes.description"))
	cmd.Flags().BoolVar(&opts.ignoreContext, "ignore-context", false, opts.localizer.MustLocalize("cluster.common.flag.ignoreContext.description"))
	cmd.Flags().StringVar(&opts.serviceID, "service-id", "", opts.localizer.MustLocalize("cluster.common.flag.serviceId.description"))
	cmd.Flags().StringVar(&opts.serviceType, "service-type", "", opts.localizer.MustLocalize("cluster.common.flag.serviceType.description"))

	return cmd
}

func runConnect(opts *options) error {
	conn, err := opts.Connection(connection.DefaultConfigSkipMasAuth)
	if err != nil {
		return err
	}

	conn.API()

	err = validateServiceID(opts.Context, conn.API(), opts.serviceType, opts.serviceID)
	if err != nil {
		return err
	}

	clusterConn, err := cluster.NewKubernetesClusterConnection(conn, opts.Config, opts.Logger, opts.kubeconfigLocation, opts.IO, opts.localizer)
	if err != nil {
		return err
	}

	arguments := &cluster.ConnectArguments{
		OfflineAccessToken:      opts.offlineAccessToken,
		ForceCreationWithoutAsk: opts.forceCreationWithoutAsk,
		IgnoreContext:           opts.ignoreContext,
		Namespace:               opts.namespace,
		SelectedService:         opts.serviceType,
		SelectedServiceID:       opts.serviceID,
	}

	connectOpts := cluster.Options{
		IO:         opts.IO,
		Logger:     opts.Logger,
		Localizer:  opts.localizer,
		Config:     opts.Config,
		Connection: conn,
	}

	var service cluster.CustomConnection

	switch opts.serviceType {
	case "kafka":
		service = &kafkaservice.KafkaConnection{}
	case "service-registry":
		service = &registryservice.ServiceRegistryConnection{}
	}

	err = clusterConn.Connect(context.Background(), arguments, service, connectOpts)
	if err != nil {
		return err
	}

	return nil
}

func validateServiceID(ctx context.Context, api *api.API, serviceType string, serviceID string) error {

	switch serviceType {
	case "kafka":
		_, _, err := kafka.GetKafkaByID(ctx, api.Kafka(), serviceID)
		if err != nil {
			return err
		}
	case "service-registry":
		_, _, err := serviceregistry.GetServiceRegistryByID(ctx, api.ServiceRegistryMgmt(), serviceID)
		if err != nil {
			return err
		}
	}
	return nil
}
