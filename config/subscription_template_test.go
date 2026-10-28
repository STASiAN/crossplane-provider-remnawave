package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/google/go-cmp/cmp"
)

func TestSubscriptionTemplateConversion(t *testing.T) {
	c := subscriptionTemplateConversion{}
	pretty := "{\n  \"routing\": {\"rules\": [{\"outboundTag\": \"direct\", \"protocol\": [\"bittorrent\"]}]},\n  \"dns\": {\"servers\": [\"1.1.1.1\"]},\n  \"remarks\": \"a<b&c\"\n}\n"
	// What the Terraform provider plans: json.Unmarshal into any, json.Marshal.
	// json.Marshal escapes < and & as \u003c and \u0026; esc is a backslash.
	esc := string(rune(92))
	canonical := `{"dns":{"servers":["1.1.1.1"]},"remarks":"a` + esc + `u003cb` + esc + `u0026c","routing":{"rules":[{"outboundTag":"direct","protocol":["bittorrent"]}]}}`
	yaml := "proxies: []\nrules:\n  - MATCH,PROXY\n"
	encoded := "cHJveGllczogW10KcnVsZXM6CiAgLSBNQVRDSCxQUk9YWQo="

	got, err := c.Convert(map[string]any{attrTemplateJSON: pretty, attrEncodedTemplateYAML: yaml, "name": "Default"}, nil, ujconfig.ToTerraform)
	if err != nil {
		t.Fatalf("ToTerraform: %v", err)
	}
	if diff := cmp.Diff(map[string]any{attrTemplateJSON: canonical, attrEncodedTemplateYAML: encoded, "name": "Default"}, got); diff != "" {
		t.Errorf("ToTerraform (-want +got):\n%s", diff)
	}

	back, err := c.Convert(map[string]any{attrTemplateJSON: canonical, attrEncodedTemplateYAML: encoded}, nil, ujconfig.FromTerraform)
	if err != nil {
		t.Fatalf("FromTerraform: %v", err)
	}
	if diff := cmp.Diff(yaml, back[attrEncodedTemplateYAML]); diff != "" {
		t.Errorf("FromTerraform yaml (-want +got):\n%s", diff)
	}
	again, err := c.Convert(map[string]any{attrTemplateJSON: back[attrTemplateJSON]}, nil, ujconfig.ToTerraform)
	if err != nil {
		t.Fatalf("ToTerraform of the indented JSON: %v", err)
	}
	if diff := cmp.Diff(canonical, again[attrTemplateJSON]); diff != "" {
		t.Errorf("JSON round trip (-want +got):\n%s", diff)
	}

	if _, err := c.Convert(map[string]any{attrTemplateJSON: "{not json"}, nil, ujconfig.ToTerraform); err == nil {
		t.Error("ToTerraform accepted invalid JSON")
	}
}
