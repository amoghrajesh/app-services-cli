package kafkacmdutil

import (
	"context"

	"github.com/redhat-developer/app-services-cli/pkg/api/rbac"
	"github.com/redhat-developer/app-services-cli/pkg/api/rbac/rbacutil"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/factory"
	"github.com/redhat-developer/app-services-cli/pkg/connection"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	"github.com/spf13/cobra"
)

// RegisterNameFlagCompletionFunc adds dynamic completion for the --name flag
func RegisterNameFlagCompletionFunc(cmd *cobra.Command, f *factory.Factory) error {
	return cmd.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var validNames []string
		directive := cobra.ShellCompDirectiveNoSpace

		conn, err := f.Connection(connection.DefaultConfigSkipMasAuth)
		if err != nil {
			return validNames, directive
		}

		req := conn.API().Kafka().GetKafkas(f.Context)
		if toComplete != "" {
			searchQ := "name like " + toComplete + "%"
			req = req.Search(searchQ)
		}
		kafkas, _, err := req.Execute()
		if err != nil {
			return validNames, directive
		}

		items := kafkas.GetItems()
		for _, kafka := range items {
			validNames = append(validNames, kafka.GetName())
		}

		return validNames, directive
	})
}

// RegisterNameFlagCompletionFunc adds dynamic completion for the --name flag
func RegisterOwnerFlagCompletionFunc(cmd *cobra.Command, f *factory.Factory) error {
	return cmd.RegisterFlagCompletionFunc("owner", func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var usernames []string
		directive := cobra.ShellCompDirectiveNoSpace

		conn, err := f.Connection(connection.DefaultConfigSkipMasAuth)
		if err != nil {
			return usernames, directive
		}

		queryParams := []rbac.QueryParam{rbac.WithQueryParam("match_criteria", "partial"), rbac.WithQueryParam("usernames", toComplete)}
		principals, err := rbacutil.FetchAllUsers(context.Background(), conn.API().RBAC.PrincipalAPI, queryParams...)
		if err != nil || len(principals) == 0 {
			return usernames, directive
		}

		for _, p := range principals {
			usernames = append(usernames, p.Username)
		}

		return usernames, directive
	})
}

// GetCloudProviderCompletionValues returns the list of supported cloud providers for creating a Kafka instance
// This is used in the cmd.RegisterFlagCompletionFunc for dynamic completion of --provider
func GetCloudProviderCompletionValues(f *factory.Factory) (validProviders []string, directive cobra.ShellCompDirective) {
	validProviders = []string{}
	directive = cobra.ShellCompDirectiveNoSpace

	conn, err := f.Connection(connection.DefaultConfigSkipMasAuth)
	if err != nil {
		return validProviders, directive
	}

	cloudProviderResponse, _, err := conn.API().Kafka().GetCloudProviders(f.Context).Execute()
	if err != nil {
		return validProviders, directive
	}

	cloudProviders := cloudProviderResponse.GetItems()
	validProviders = GetEnabledCloudProviderNames(cloudProviders)

	return validProviders, directive
}

// GetEnabledCloudProviderNames returns a list of cloud provider names from the enabled cloud providers
func GetEnabledCloudProviderNames(cloudProviders []kafkamgmtclient.CloudProvider) []string {
	cloudProviderNames := []string{}
	for _, provider := range cloudProviders {
		if provider.GetEnabled() {
			cloudProviderNames = append(cloudProviderNames, provider.GetName())
		}
	}
	return cloudProviderNames
}

// FindCloudProviderByName finds and returns a cloud provider item from the list by its name
func FindCloudProviderByName(cloudProviders []kafkamgmtclient.CloudProvider, name string) *kafkamgmtclient.CloudProvider {
	for _, p := range cloudProviders {
		if p.GetName() == name {
			return &p
		}
	}
	return nil
}

// GetEnabledCloudRegionIDs extracts and returns a slice of the unique IDs of all enabled regions
func GetEnabledCloudRegionIDs(regions []kafkamgmtclient.CloudRegion) []string {
	regionIDs := []string{}
	for _, region := range regions {
		if region.GetEnabled() {
			regionIDs = append(regionIDs, region.GetId())
		}
	}
	return regionIDs
}
