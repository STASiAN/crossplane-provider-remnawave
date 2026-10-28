package config

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// uuidAsIdentifier is IdentifierFromProvider for resources whose Terraform
// schema carries the server-assigned identifier in `uuid` and has no `id`
// attribute, so IDAsExternalName cannot find the identifier in the tfstate.
var uuidAsIdentifier = func() config.ExternalName {
	e := config.IdentifierFromProvider
	// The Terraform Plugin Framework client reads a resource through its prior
	// state, so the identifier has to be put into `uuid` there. It is
	// computed-only, hence kept out of the configuration.
	e.SetIdentifierArgumentFn = func(base map[string]any, name string) {
		if name != "" {
			base["uuid"] = name
		}
	}
	e.TFPluginFrameworkOptions.ComputedIdentifierAttributes = []string{"uuid"}
	e.GetExternalNameFn = func(tfstate map[string]any) (string, error) {
		if v, ok := tfstate["uuid"].(string); ok && v != "" {
			return v, nil
		}
		return config.IDAsExternalName(tfstate)
	}
	return e
}()

// uuidStateEmpty reports a state without `uuid` as empty. Before the first
// create, upjet rebuilds the prior state from the parameters, which is not
// null; without a `uuid` the provider cannot find the resource, so the state
// must not count as an existing resource.
func uuidStateEmpty(_ context.Context, state tftypes.Value, _ rschema.Schema) (bool, error) {
	if state.IsNull() || !state.IsKnown() {
		return true, nil
	}
	var attrs map[string]tftypes.Value
	if err := state.As(&attrs); err != nil {
		return false, err
	}
	uuid, ok := attrs["uuid"]
	if !ok || uuid.IsNull() || !uuid.IsKnown() {
		return true, nil
	}
	var s string
	if err := uuid.As(&s); err != nil {
		return false, err
	}
	return s == "", nil
}

// ExternalNameConfigs contains all external name configurations for this
// provider. Remnawave assigns identifiers (UUIDs) server-side, so every
// resource uses IdentifierFromProvider; those whose schema has `uuid` instead
// of `id` read the identifier from it. This map also acts as the include
// list: the imperative *_action and drop_connections resources are excluded
// by not being listed here.
var ExternalNameConfigs = map[string]config.ExternalName{
	"remnawave_user":                  config.IdentifierFromProvider,
	"remnawave_node":                  config.IdentifierFromProvider,
	"remnawave_host":                  uuidAsIdentifier,
	"remnawave_config_profile":        uuidAsIdentifier,
	"remnawave_external_squad":        uuidAsIdentifier,
	"remnawave_internal_squad":        uuidAsIdentifier,
	"remnawave_subscription_template": uuidAsIdentifier,
	"remnawave_subscription_settings": config.IdentifierFromProvider,
	"remnawave_subpage_config":        uuidAsIdentifier,
	"remnawave_panel_settings":        config.IdentifierFromProvider,
	"remnawave_snippet":               config.IdentifierFromProvider,
	"remnawave_node_plugin":           uuidAsIdentifier,
	"remnawave_infra_provider":        uuidAsIdentifier,
	"remnawave_billing_node":          uuidAsIdentifier,
	"remnawave_billing_history":       uuidAsIdentifier,
	"remnawave_api_token":             uuidAsIdentifier,
	"remnawave_passkey":               uuidAsIdentifier,
	"remnawave_user_metadata":         uuidAsIdentifier,
	"remnawave_node_metadata":         uuidAsIdentifier,
	"remnawave_hwid_device":           config.IdentifierFromProvider,
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
			for _, a := range e.TFPluginFrameworkOptions.ComputedIdentifierAttributes {
				if a == "uuid" {
					r.TerraformPluginFrameworkIsStateEmptyFn = uuidStateEmpty
				}
			}
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
