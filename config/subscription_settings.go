package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

// configureSubscriptionSettings exposes custom_response_headers as a map of
// header name to value instead of the JSON string the Terraform schema uses,
// with the Happ routing header as readable JSON (see happRoutingHeader).
func configureSubscriptionSettings(p *ujconfig.Provider) {
	p.AddResourceConfigurator("remnawave_subscription_settings", func(r *ujconfig.Resource) {
		r.TerraformResource.Schema["custom_response_headers"] = &schema.Schema{
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Headers the panel returns on every subscription fetch, by name.",
		}
		r.TerraformConversions = append(r.TerraformConversions, jsonStringMapConversion{"custom_response_headers"})
	})
}

// jsonStringMapConversion converts the named Terraform attributes between the
// JSON-object string Terraform expects and the map exposed in the CRD.
type jsonStringMapConversion []string

func (c jsonStringMapConversion) Convert(params map[string]any, _ *ujconfig.Resource, mode ujconfig.Mode) (map[string]any, error) {
	for _, k := range c {
		v, ok := params[k]
		if !ok || v == nil {
			continue
		}
		switch mode {
		case ujconfig.ToTerraform:
			m, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if err := encodeHappRouting(m); err != nil {
				return nil, errors.Wrapf(err, "cannot encode %s", k)
			}
			s, err := canonicalJSON(m)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot encode %s", k)
			}
			params[k] = s
		case ujconfig.FromTerraform:
			s, ok := v.(string)
			if !ok {
				continue
			}
			m := map[string]any{}
			if s != "" {
				if err := json.Unmarshal([]byte(s), &m); err != nil {
					return nil, errors.Wrapf(err, "cannot decode %s", k)
				}
			}
			decodeHappRouting(m)
			params[k] = m
		}
	}
	return params, nil
}

// canonicalJSON encodes m with sorted keys, no extra whitespace and no HTML
// escaping: the Terraform provider rejects any other formatting of the value
// it plans.
func canonicalJSON(m map[string]any) (string, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(m); err != nil {
		return "", err
	}
	return string(bytes.TrimRight(b.Bytes(), "\n")), nil
}

const (
	// happRoutingHeader carries a Happ routing profile as
	// happ://routing/onadd/<base64url of the profile JSON>; the CRD holds the
	// JSON itself.
	happRoutingHeader = "routing"
	happRoutingPrefix = "happ://routing/onadd/"
)

// encodeHappRouting turns a JSON routing profile into the Happ link. The JSON
// is compacted rather than re-marshalled, so the key order and escapes are
// kept and the link matches the one the panel holds byte for byte.
func encodeHappRouting(headers map[string]any) error {
	s, ok := headers[happRoutingHeader].(string)
	if !ok || !strings.HasPrefix(strings.TrimSpace(s), "{") {
		return nil
	}
	var b bytes.Buffer
	if err := json.Compact(&b, []byte(s)); err != nil {
		return errors.Wrapf(err, "invalid JSON in the %q header", happRoutingHeader)
	}
	headers[happRoutingHeader] = happRoutingPrefix + base64.RawURLEncoding.EncodeToString(b.Bytes())
	return nil
}

// decodeHappRouting turns the Happ link back into indented JSON. A value it
// cannot decode is left as is.
func decodeHappRouting(headers map[string]any) {
	s, ok := headers[happRoutingHeader].(string)
	if !ok || !strings.HasPrefix(s, happRoutingPrefix) {
		return
	}
	enc := strings.TrimRight(strings.TrimPrefix(s, happRoutingPrefix), "=")
	raw, err := base64.RawURLEncoding.DecodeString(enc)
	if err != nil {
		if raw, err = base64.RawStdEncoding.DecodeString(enc); err != nil {
			return
		}
	}
	var b bytes.Buffer
	if err := json.Indent(&b, raw, "", "  "); err != nil {
		return
	}
	headers[happRoutingHeader] = b.String()
}
