package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// ExternalNameConfigs contains all external name configurations for this
// provider. Remnawave assigns identifiers (UUIDs) server-side, so every
// resource uses IdentifierFromProvider. This map also acts as the include
// list: the imperative *_action and drop_connections resources are excluded
// by not being listed here.
var ExternalNameConfigs = map[string]config.ExternalName{
	"remnawave_user":                  config.IdentifierFromProvider,
	"remnawave_node":                  config.IdentifierFromProvider,
	"remnawave_host":                  config.IdentifierFromProvider,
	"remnawave_config_profile":        config.IdentifierFromProvider,
	"remnawave_external_squad":        config.IdentifierFromProvider,
	"remnawave_internal_squad":        config.IdentifierFromProvider,
	"remnawave_subscription_template": config.IdentifierFromProvider,
	"remnawave_subscription_settings": config.IdentifierFromProvider,
	"remnawave_subpage_config":        config.IdentifierFromProvider,
	"remnawave_panel_settings":        config.IdentifierFromProvider,
	"remnawave_snippet":               config.IdentifierFromProvider,
	"remnawave_node_plugin":           config.IdentifierFromProvider,
	"remnawave_infra_provider":        config.IdentifierFromProvider,
	"remnawave_billing_node":          config.IdentifierFromProvider,
	"remnawave_billing_history":       config.IdentifierFromProvider,
	"remnawave_api_token":             config.IdentifierFromProvider,
	"remnawave_passkey":               config.IdentifierFromProvider,
	"remnawave_user_metadata":         config.IdentifierFromProvider,
	"remnawave_node_metadata":         config.IdentifierFromProvider,
	"remnawave_hwid_device":           config.IdentifierFromProvider,
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
