package config

import (
	"encoding/json"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

const (
	attrConfig            = "config"
	attrRealityPrivateKey = "reality_private_key"
)

// configureConfigProfile adds realityPrivateKeySecretRef to ConfigProfile, so
// the Reality private key can live in its own Secret while the rest of the
// Xray config stays reviewable. The key is not a Terraform attribute: it is
// injected into the config on the way to Terraform.
func configureConfigProfile(p *ujconfig.Provider) {
	p.AddResourceConfigurator("remnawave_config_profile", func(r *ujconfig.Resource) {
		r.TerraformResource.Schema[attrRealityPrivateKey] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "Reality private key, set as realitySettings.privateKey of every inbound that has realitySettings.",
		}
		r.TerraformConfigurationInjector = injectRealityPrivateKey
	})
}

// injectRealityPrivateKey moves the resolved reality_private_key into the
// config. It runs after the secret references are resolved.
func injectRealityPrivateKey(_ map[string]any, tfMap map[string]any) error {
	key, _ := tfMap[attrRealityPrivateKey].(string)
	delete(tfMap, attrRealityPrivateKey)
	if key == "" {
		return nil
	}
	s, _ := tfMap[attrConfig].(string)
	if s == "" {
		return errors.New("realityPrivateKeySecretRef is set but the config is empty")
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(s), &cfg); err != nil {
		return errors.Wrap(err, "cannot decode config")
	}
	inbounds, _ := cfg["inbounds"].([]any)
	n := 0
	for _, ib := range inbounds {
		ibm, _ := ib.(map[string]any)
		ss, _ := ibm["streamSettings"].(map[string]any)
		rs, ok := ss["realitySettings"].(map[string]any)
		if !ok {
			continue
		}
		rs["privateKey"] = key
		n++
	}
	if n == 0 {
		return errors.New("realityPrivateKeySecretRef is set but no inbound has realitySettings")
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "cannot encode config")
	}
	tfMap[attrConfig] = string(b)
	return nil
}
