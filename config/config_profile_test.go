package config

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInjectRealityPrivateKey(t *testing.T) {
	config := `{"inbounds":[{"tag":"Hysteria","streamSettings":{"security":"tls"}},` +
		`{"tag":"VLESS-xHTTP-SS","streamSettings":{"security":"reality","realitySettings":{"shortIds":["ab"]}}}]}`

	tf := map[string]any{attrConfig: config, attrRealityPrivateKey: "secret-key", "name": "p"}
	if err := injectRealityPrivateKey(nil, tf); err != nil {
		t.Fatal(err)
	}
	if _, ok := tf[attrRealityPrivateKey]; ok {
		t.Error("reality_private_key left in the Terraform parameters")
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(tf[attrConfig].(string)), &got); err != nil {
		t.Fatal(err)
	}
	ibs := got["inbounds"].([]any)
	rs := ibs[1].(map[string]any)["streamSettings"].(map[string]any)["realitySettings"].(map[string]any)
	if diff := cmp.Diff(map[string]any{"shortIds": []any{"ab"}, "privateKey": "secret-key"}, rs); diff != "" {
		t.Errorf("realitySettings (-want +got):\n%s", diff)
	}
	if _, ok := ibs[0].(map[string]any)["streamSettings"].(map[string]any)["realitySettings"]; ok {
		t.Error("realitySettings added to an inbound without Reality")
	}

	// No key: the config is left untouched.
	plain := map[string]any{attrConfig: config}
	if err := injectRealityPrivateKey(nil, plain); err != nil || plain[attrConfig] != config {
		t.Errorf("config changed without a key: %v, %v", err, plain[attrConfig])
	}

	// A key but no Reality inbound is a mistake in the manifest.
	noReality := map[string]any{attrConfig: `{"inbounds":[{"tag":"x","streamSettings":{}}]}`, attrRealityPrivateKey: "k"}
	if err := injectRealityPrivateKey(nil, noReality); err == nil {
		t.Error("no error for a key without a Reality inbound")
	}
}
