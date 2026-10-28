package config

import (
	"strings"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/google/go-cmp/cmp"
)

func TestJSONStringMapConversion(t *testing.T) {
	c := jsonStringMapConversion{"custom_response_headers"}
	headers := map[string]any{
		"routing":            "happ://routing/onadd/eyJ&<>",
		"exclude-routes-set": "172.16.0.0/12, 192.168.0.0/16",
		"providerid":         "JAyGKgQq",
	}
	// Sorted keys, compact, no HTML escaping: the form the provider plans.
	want := `{"exclude-routes-set":"172.16.0.0/12, 192.168.0.0/16","providerid":"JAyGKgQq","routing":"happ://routing/onadd/eyJ&<>"}`

	got, err := c.Convert(map[string]any{"custom_response_headers": headers, "randomize_hosts": true}, nil, ujconfig.ToTerraform)
	if err != nil {
		t.Fatalf("ToTerraform: %v", err)
	}
	if diff := cmp.Diff(want, got["custom_response_headers"]); diff != "" {
		t.Errorf("ToTerraform (-want +got):\n%s", diff)
	}
	if got["randomize_hosts"] != true {
		t.Errorf("ToTerraform changed an unrelated attribute: %v", got["randomize_hosts"])
	}

	back, err := c.Convert(map[string]any{"custom_response_headers": want}, nil, ujconfig.FromTerraform)
	if err != nil {
		t.Fatalf("FromTerraform: %v", err)
	}
	if diff := cmp.Diff(headers, back["custom_response_headers"]); diff != "" {
		t.Errorf("FromTerraform (-want +got):\n%s", diff)
	}

	empty, err := c.Convert(map[string]any{"custom_response_headers": ""}, nil, ujconfig.FromTerraform)
	if err != nil {
		t.Fatalf("FromTerraform empty: %v", err)
	}
	if diff := cmp.Diff(map[string]any{}, empty["custom_response_headers"]); diff != "" {
		t.Errorf("FromTerraform empty (-want +got):\n%s", diff)
	}
}

func TestHappRoutingRoundTrip(t *testing.T) {
	// The routing header of the infra panel.
	link := "happ://routing/onadd/eyJOYW1lIjoiIiwiR2xvYmFsUHJveHkiOiJ0cnVlIiwiUm91dGVPcmRlciI6ImJsb2NrLXByb3h5LWRpcmVjdCIsIlJlbW90ZUROU1R5cGUiOiJEb0giLCJSZW1vdGVETlNEb21haW4iOiJodHRwczovL2Nsb3VkZmxhcmUtZG5zLmNvbS9kbnMtcXVlcnkiLCJSZW1vdGVETlNJUCI6IjEuMS4xLjEiLCJEb21lc3RpY0ROU1R5cGUiOiJEb0giLCJEb21lc3RpY0ROU0RvbWFpbiI6Imh0dHBzOi8vZG5zLmdvb2dsZS9kbnMtcXVlcnkiLCJEb21lc3RpY0ROU0lQIjoiOC44LjguOCIsIkdlb2lwdXJsIjoiaHR0cHM6Ly9naXRodWIuY29tL0xveWFsc29sZGllci92MnJheS1ydWxlcy1kYXQvcmVsZWFzZXMvbGF0ZXN0L2Rvd25sb2FkL2dlb2lwLmRhdCIsIkdlb3NpdGV1cmwiOiJodHRwczovL2dpdGh1Yi5jb20vTG95YWxzb2xkaWVyL3YycmF5LXJ1bGVzLWRhdC9yZWxlYXNlcy9sYXRlc3QvZG93bmxvYWQvZ2Vvc2l0ZS5kYXQiLCJMYXN0VXBkYXRlZCI6IjE3ODY5NjQ0NTUiLCJEbnNIb3N0cyI6eyJjbG91ZGZsYXJlLWRucy5jb20iOiIxLjEuMS4xIiwiZG5zLmdvb2dsZSI6IjguOC44LjgifSwiRGlyZWN0U2l0ZXMiOltdLCJEaXJlY3RJcCI6WyIxNzIuMTYuMC4wLzEyIiwiMTkyLjE2OC4wLjAvMTYiLCIxNjkuMjU0LjAuMC8xNiIsIjIyNC4wLjAuMC80IiwiMjU1LjI1NS4yNTUuMjU1Il0sIlByb3h5U2l0ZXMiOltdLCJQcm94eUlwIjpbXSwiQmxvY2tTaXRlcyI6W10sIkJsb2NrSXAiOltdLCJEb21haW5TdHJhdGVneSI6IklQSWZOb25NYXRjaCIsIkZha2VETlMiOiJmYWxzZSIsIlVzZUNodW5rRmlsZXMiOiJ0cnVlIn0"
	c := jsonStringMapConversion{"custom_response_headers"}
	state, err := canonicalJSON(map[string]any{"routing": link, "providerid": "JAyGKgQq"})
	if err != nil {
		t.Fatal(err)
	}

	obs, err := c.Convert(map[string]any{"custom_response_headers": state}, nil, ujconfig.FromTerraform)
	if err != nil {
		t.Fatalf("FromTerraform: %v", err)
	}
	pretty := obs["custom_response_headers"].(map[string]any)["routing"].(string)
	if !strings.HasPrefix(pretty, "{\n  \"Name\": \"\",\n  \"GlobalProxy\": \"true\",") {
		t.Errorf("routing is not indented JSON in the original key order:\n%s", pretty)
	}

	// A YAML block scalar adds a trailing newline; the link must not change.
	params := map[string]any{"custom_response_headers": map[string]any{"routing": pretty + "\n", "providerid": "JAyGKgQq"}}
	got, err := c.Convert(params, nil, ujconfig.ToTerraform)
	if err != nil {
		t.Fatalf("ToTerraform: %v", err)
	}
	if diff := cmp.Diff(state, got["custom_response_headers"]); diff != "" {
		t.Errorf("round trip changed the header (-want +got):\n%s", diff)
	}

	if _, err := c.Convert(map[string]any{"custom_response_headers": map[string]any{"routing": "{not json"}}, nil, ujconfig.ToTerraform); err == nil {
		t.Error("ToTerraform accepted invalid routing JSON")
	}
}
