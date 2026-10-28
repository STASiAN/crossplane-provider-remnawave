package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/pkg/errors"
)

const (
	attrTemplateJSON        = "template_json"
	attrEncodedTemplateYAML = "encoded_template_yaml"
)

// configureSubscriptionTemplate keeps subscription templates readable in the
// CRD: templateJson as indented JSON and encodedTemplateYaml as the plain YAML
// text, which the provider base64-encodes for the panel.
func configureSubscriptionTemplate(p *ujconfig.Provider) {
	p.AddResourceConfigurator("remnawave_subscription_template", func(r *ujconfig.Resource) {
		if s, ok := r.TerraformResource.Schema[attrEncodedTemplateYAML]; ok {
			s.Description = "Template YAML as plain text; the provider base64-encodes it for the panel."
		}
		r.TerraformConversions = append(r.TerraformConversions, subscriptionTemplateConversion{})
	})
}

type subscriptionTemplateConversion struct{}

func (subscriptionTemplateConversion) Convert(params map[string]any, _ *ujconfig.Resource, mode ujconfig.Mode) (map[string]any, error) {
	if s, ok := params[attrTemplateJSON].(string); ok && s != "" {
		v, err := convertTemplateJSON(s, mode)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot convert %s", attrTemplateJSON)
		}
		params[attrTemplateJSON] = v
	}
	if s, ok := params[attrEncodedTemplateYAML].(string); ok && s != "" {
		params[attrEncodedTemplateYAML] = convertTemplateYAML(s, mode)
	}
	return params, nil
}

// convertTemplateJSON sends the JSON in the form the Terraform provider plans
// (json.Unmarshal into any, then json.Marshal: sorted keys, compact, HTML
// escaped) and shows it indented.
func convertTemplateJSON(s string, mode ujconfig.Mode) (string, error) {
	if mode == ujconfig.FromTerraform {
		var b bytes.Buffer
		if err := json.Indent(&b, []byte(s), "", "  "); err != nil {
			return s, nil //nolint:nilerr // keep a value that is not JSON as is
		}
		return b.String(), nil
	}
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// convertTemplateYAML base64-encodes the YAML for Terraform and decodes it
// back; a state value that is not base64 is kept as is.
func convertTemplateYAML(s string, mode ujconfig.Mode) string {
	if mode == ujconfig.ToTerraform {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}
	return string(b)
}
