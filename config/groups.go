package config

import "github.com/crossplane/upjet/v2/pkg/config"

// kindMap places every resource in the root API group
// (remnawave.crossplane.io / remnawave.m.crossplane.io) with a fully
// descriptive kind, instead of upjet's default word-splitting which would
// produce groups like "config" with kind "Profile" for
// remnawave_config_profile.
var kindMap = map[string]string{
	"remnawave_user":                  "User",
	"remnawave_node":                  "Node",
	"remnawave_host":                  "Host",
	"remnawave_config_profile":        "ConfigProfile",
	"remnawave_external_squad":        "ExternalSquad",
	"remnawave_internal_squad":        "InternalSquad",
	"remnawave_subscription_template": "SubscriptionTemplate",
	"remnawave_subscription_settings": "SubscriptionSettings",
	"remnawave_subpage_config":        "SubpageConfig",
	"remnawave_panel_settings":        "PanelSettings",
	"remnawave_snippet":               "Snippet",
	"remnawave_node_plugin":           "NodePlugin",
	"remnawave_infra_provider":        "InfraProvider",
	"remnawave_billing_node":          "BillingNode",
	"remnawave_billing_history":       "BillingHistory",
	"remnawave_api_token":             "APIToken",
	"remnawave_passkey":               "Passkey",
	"remnawave_user_metadata":         "UserMetadata",
	"remnawave_node_metadata":         "NodeMetadata",
	"remnawave_hwid_device":           "HwidDevice",
}

// GroupKindOverrides moves all resources into the root group with explicit
// kinds.
func GroupKindOverrides() config.ResourceOption {
	return func(r *config.Resource) {
		if k, ok := kindMap[r.Name]; ok {
			r.ShortGroup = ""
			r.Kind = k
		}
	}
}
