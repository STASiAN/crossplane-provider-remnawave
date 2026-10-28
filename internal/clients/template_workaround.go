package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// This file works around batonogov/terraform-provider-remnawave#283: the
// Terraform provider sends `name` on every subscription template update, and
// the panel rejects the reserved name "Default" (RESERVED_TEMPLATE_NAME) even
// when it does not change, so the built-in Default templates cannot be
// updated. The subscription template resource is wrapped and its Update sends
// the PATCH itself with name only on a rename; everything else is delegated.
// Remove once the upstream fix (#284) is released and adopted.

const subscriptionTemplateTypeName = "remnawave_subscription_template"

// panelConn is what the template workaround needs to call the panel API.
type panelConn struct {
	endpoint      string
	apiToken      string
	insecure      bool
	proxyHeaders  bool
	customHeaders map[string]string
	timeout       time.Duration
}

func panelConnFrom(cfg map[string]any) panelConn {
	c := panelConn{timeout: 30 * time.Second}
	c.endpoint, _ = cfg[keyEndpoint].(string)
	c.apiToken, _ = cfg[keyAPIToken].(string)
	c.insecure, _ = cfg[keyInsecureSkipVerify].(bool)
	c.proxyHeaders, _ = cfg[keyProxyHeaders].(bool)
	if h, ok := cfg[keyCustomHeaders].(map[string]any); ok {
		c.customHeaders = map[string]string{}
		for k, v := range h {
			if s, ok := v.(string); ok {
				c.customHeaders[k] = s
			}
		}
	}
	if s, ok := cfg[keyRequestTimeout].(string); ok {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			c.timeout = d
		}
	}
	return c
}

// templateWorkaroundProvider replaces the subscription template resource of the
// wrapped provider; all other calls are passed through.
type templateWorkaroundProvider struct {
	fwprovider.Provider
	conn panelConn
}

func withTemplateWorkaround(p fwprovider.Provider, conn panelConn) fwprovider.Provider {
	return &templateWorkaroundProvider{Provider: p, conn: conn}
}

func (p *templateWorkaroundProvider) Resources(ctx context.Context) []func() resource.Resource {
	fns := p.Provider.Resources(ctx)
	for i, fn := range fns {
		var md resource.MetadataResponse
		fn().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "remnawave"}, &md)
		if md.TypeName != subscriptionTemplateTypeName {
			continue
		}
		inner := fn
		fns[i] = func() resource.Resource {
			return &templateWorkaroundResource{Resource: inner(), conn: p.conn}
		}
	}
	return fns
}

type templateWorkaroundResource struct {
	resource.Resource
	conn panelConn
}

type templateModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Name         types.String `tfsdk:"name"`
	TemplateType types.String `tfsdk:"template_type"`
	TemplateJSON types.String `tfsdk:"template_json"`
	EncodedYaml  types.String `tfsdk:"encoded_template_yaml"`
}

func (r *templateWorkaroundResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c, ok := r.Resource.(resource.ResourceWithConfigure); ok {
		c.Configure(ctx, req, resp)
	}
}

func (r *templateWorkaroundResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if i, ok := r.Resource.(resource.ResourceWithImportState); ok {
		i.ImportState(ctx, req, resp)
	}
}

func (r *templateWorkaroundResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.conn.apiToken == "" {
		// Username/password sessions are not handled here.
		r.Resource.Update(ctx, req, resp)
		return
	}
	var plan, state templateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"uuid": state.UUID.ValueString()}
	if !plan.Name.Equal(state.Name) {
		body["name"] = plan.Name.ValueString()
	}
	if !plan.TemplateJSON.IsNull() && !plan.TemplateJSON.IsUnknown() {
		var v any
		if err := json.Unmarshal([]byte(plan.TemplateJSON.ValueString()), &v); err != nil {
			resp.Diagnostics.AddError("Invalid template_json", err.Error())
			return
		}
		body["templateJson"] = v
	}
	if !plan.EncodedYaml.IsNull() && !plan.EncodedYaml.IsUnknown() {
		body["encodedTemplateYaml"] = plan.EncodedYaml.ValueString()
	}
	if err := r.conn.patch(ctx, "/api/subscription-templates", body); err != nil {
		resp.Diagnostics.AddError("Failed to update subscription template", err.Error())
		return
	}
	// Read the result back through the provider so the state has its usual form.
	readResp := resource.ReadResponse{State: req.State}
	r.Read(ctx, resource.ReadRequest{State: req.State}, &readResp)
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

func (c panelConn) patch(ctx context.Context, path string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, strings.TrimRight(c.endpoint, "/")+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	for k, v := range c.customHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	if c.proxyHeaders {
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		req.Header.Set("X-Forwarded-Proto", "https")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if c.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in via the ProviderConfig
	}
	res, err := (&http.Client{Transport: transport, Timeout: c.timeout}).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck // read-only body
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("request failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}
