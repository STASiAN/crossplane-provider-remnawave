package clients

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakeResource stands in for a Terraform provider resource; its Read marks the
// state so the test can tell it ran.
type fakeResource struct{ typeName string }

func (f *fakeResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = f.typeName
}

func (f *fakeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"uuid":                  schema.StringAttribute{Computed: true},
		"name":                  schema.StringAttribute{Required: true},
		"template_type":         schema.StringAttribute{Required: true},
		"template_json":         schema.StringAttribute{Optional: true, Computed: true},
		"encoded_template_yaml": schema.StringAttribute{Optional: true, Computed: true},
	}}
}

func (f *fakeResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {}

func (f *fakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m templateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	m.EncodedYaml = types.StringValue("read-back")
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (f *fakeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("inner Update called", "the workaround must not delegate Update")
}

func (f *fakeResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {}

type fakeProvider struct{ fwprovider.Provider }

func (fakeProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &fakeResource{typeName: "remnawave_node"} },
		func() resource.Resource { return &fakeResource{typeName: subscriptionTemplateTypeName} },
	}
}

func TestTemplateWorkaroundWrapsOnlyTheTemplateResource(t *testing.T) {
	fns := withTemplateWorkaround(fakeProvider{}, panelConn{}).Resources(t.Context())
	if _, ok := fns[0]().(*templateWorkaroundResource); ok {
		t.Error("remnawave_node was wrapped")
	}
	if _, ok := fns[1]().(*templateWorkaroundResource); !ok {
		t.Error("remnawave_subscription_template was not wrapped")
	}
}

func TestTemplateWorkaroundUpdate(t *testing.T) {
	tests := []struct {
		name      string
		stateName string
		planName  string
		wantName  bool
	}{
		{name: "unchanged Default name is omitted", stateName: "Default", planName: "Default"},
		{name: "rename is sent", stateName: "old-name", planName: "new-name", wantName: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body map[string]any
			var header http.Header
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch || r.URL.Path != "/api/subscription-templates" {
					t.Errorf("request = %s %s, want PATCH /api/subscription-templates", r.Method, r.URL.Path)
				}
				header = r.Header
				b, _ := io.ReadAll(r.Body)
				if err := json.Unmarshal(b, &body); err != nil {
					t.Errorf("decode body: %v", err)
				}
				_, _ = io.WriteString(w, `{"response":{}}`)
			}))
			defer server.Close()

			r := &templateWorkaroundResource{
				Resource: &fakeResource{typeName: subscriptionTemplateTypeName},
				conn:     panelConn{endpoint: server.URL + "/", apiToken: "token", proxyHeaders: true, timeout: 5e9},
			}
			var sr resource.SchemaResponse
			r.Schema(t.Context(), resource.SchemaRequest{}, &sr)
			m := templateModel{
				UUID: types.StringValue("tpl-uuid"), Name: types.StringValue(tt.stateName), TemplateType: types.StringValue("XRAY_JSON"),
				TemplateJSON: types.StringValue(`{"dns":{}}`), EncodedYaml: types.StringNull(),
			}
			state := tfsdk.State{Schema: sr.Schema}
			state.Set(t.Context(), &m)
			m.Name = types.StringValue(tt.planName)
			m.TemplateJSON = types.StringValue(`{"routing":{"rules":[]}}`)
			m.EncodedYaml = types.StringUnknown()
			plan := tfsdk.Plan{Schema: sr.Schema}
			plan.Set(t.Context(), &m)

			resp := resource.UpdateResponse{State: state}
			r.Update(t.Context(), resource.UpdateRequest{State: state, Plan: plan}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Update diagnostics: %v", resp.Diagnostics)
			}
			if _, ok := body["name"]; ok != tt.wantName {
				t.Errorf("name in body = %v, want %v (body: %v)", ok, tt.wantName, body)
			}
			if body["uuid"] != "tpl-uuid" {
				t.Errorf("uuid = %v, want tpl-uuid", body["uuid"])
			}
			if _, ok := body["templateJson"].(map[string]any)["routing"]; !ok {
				t.Errorf("templateJson not sent from the plan: %v", body)
			}
			if _, ok := body["encodedTemplateYaml"]; ok {
				t.Errorf("unknown encodedTemplateYaml was sent: %v", body)
			}
			if header.Get("Authorization") != "Bearer token" || header.Get("X-Forwarded-Proto") != "https" {
				t.Errorf("headers = %v", header)
			}
			var got templateModel
			resp.State.Get(t.Context(), &got)
			if got.EncodedYaml.ValueString() != "read-back" {
				t.Error("state was not refreshed through the inner Read")
			}
		})
	}
}

func TestTemplateWorkaroundPatchReportsPanelError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"message":"Reserved template name"}`)
	}))
	defer server.Close()
	err := panelConn{endpoint: server.URL, apiToken: "token", timeout: 5e9}.patch(t.Context(), "/api/subscription-templates", map[string]any{})
	if err == nil || err.Error() != `request failed: status 400: {"message":"Reserved template name"}` {
		t.Errorf("err = %v", err)
	}
}
