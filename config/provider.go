package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	tfprovider "github.com/batonogov/terraform-provider-remnawave/provider"
)

const (
	resourcePrefix = "remnawave"
	modulePath     = "github.com/STASiAN/crossplane-provider-remnawave"
)

// tfProviderVersion is only reported to the embedded Terraform provider.
const tfProviderVersion = "crossplane"

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("remnawave.crossplane.io"),
		// Every resource goes through the embedded Plugin Framework provider.
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
		ujconfig.WithTerraformPluginFrameworkProvider(tfprovider.New(tfProviderVersion)()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
			GroupKindOverrides(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		configureSubscriptionSettings,
		configureSubscriptionTemplate,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}

// GetProviderNamespaced returns the namespaced provider configuration
func GetProviderNamespaced() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("remnawave.m.crossplane.io"),
		// Every resource goes through the embedded Plugin Framework provider.
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
		ujconfig.WithTerraformPluginFrameworkProvider(tfprovider.New(tfProviderVersion)()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
			GroupKindOverrides(),
		),
		ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{
			ManagedResourceNamespace: "crossplane-system",
		}))

	for _, configure := range []func(provider *ujconfig.Provider){
		configureSubscriptionSettings,
		configureSubscriptionTemplate,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
