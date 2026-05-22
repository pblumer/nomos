package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OperationInvokeResult is the outcome of proxying an operation call.
type OperationInvokeResult struct {
	Protocol    string `json:"protocol"`
	Request     string `json:"request"`          // human-readable request line, e.g. "GET https://…"
	Status      int    `json:"status,omitempty"` // HTTP status (REST)
	ContentType string `json:"content_type,omitempty"`
	Body        string `json:"body,omitempty"`
	DurationMs  int64  `json:"duration_ms,omitempty"`
}

// invokeMaxBody caps how much of a proxied response we read into memory.
const invokeMaxBody = 1 << 20 // 1 MiB

// InvokeOperation proxies a call to the external API a service operation binds
// to. REST is implemented; MCP and gRPC return a not-yet-supported error.
//
// inputs supplies values keyed by parameter/payload-field name. Reserved keys:
// "_body" overrides the request body, "_auth" supplies the bearer token or API
// key value for the operation's declared security scheme.
//
// Note: this performs a server-side request to an operator-configured URL. It is
// a deliberate proxy; callers control the target via the operation's base_url.
func InvokeOperation(loc, serviceName, operationName string, inputs map[string]any) (OperationInvokeResult, error) {
	svc, err := GetService(loc, serviceName)
	if err != nil {
		return OperationInvokeResult{}, err
	}
	var op *OperationDTO
	for i := range svc.Operations {
		if svc.Operations[i].Name == operationName {
			op = &svc.Operations[i]
			break
		}
	}
	if op == nil {
		return OperationInvokeResult{}, Error(CodeServiceNotFound, "operation not found: "+operationName, http.StatusNotFound, nil)
	}
	if inputs == nil {
		inputs = map[string]any{}
	}
	protocol := op.Protocol
	if protocol == "" {
		protocol = "rest"
	}
	if protocol != "rest" {
		return OperationInvokeResult{Protocol: protocol}, Error(CodeInvalidInput, protocol+" operation invocation is not yet supported (REST only)", http.StatusNotImplemented, nil)
	}
	return invokeREST(*op, inputs)
}

func invokeREST(op OperationDTO, inputs map[string]any) (OperationInvokeResult, error) {
	res := OperationInvokeResult{Protocol: "rest"}
	method := strings.ToUpper(strings.TrimSpace(op.HTTPMethod))
	if method == "" {
		method = http.MethodGet
	}
	if strings.TrimSpace(op.BaseURL) == "" {
		return res, Error(CodeInvalidInput, "operation has no base_url to call", http.StatusBadRequest, nil)
	}

	// Substitute {name} path placeholders from inputs.
	rawPath := op.Path
	for _, p := range op.Parameters {
		if p.In == "path" {
			rawPath = strings.ReplaceAll(rawPath, "{"+p.Name+"}", url.PathEscape(asString(inputs[p.Name])))
		}
	}
	u, err := url.Parse(strings.TrimRight(op.BaseURL, "/") + "/" + strings.TrimLeft(rawPath, "/"))
	if err != nil {
		return res, Error(CodeInvalidInput, "invalid operation URL: "+err.Error(), http.StatusBadRequest, err)
	}
	// Query parameters.
	q := u.Query()
	for _, p := range op.Parameters {
		if p.In == "query" {
			if v, ok := inputs[p.Name]; ok {
				q.Set(p.Name, asString(v))
			}
		}
	}
	u.RawQuery = q.Encode()

	// Body for write methods.
	var body io.Reader
	contentType := ""
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if b, ok := inputs["_body"]; ok {
			data, _ := json.Marshal(b)
			body = bytes.NewReader(data)
			contentType = "application/json"
		} else if op.Payload != nil {
			obj := map[string]any{}
			for _, f := range op.Payload.Fields {
				if v, ok := inputs[f.Name]; ok {
					obj[f.Name] = v
				}
			}
			data, _ := json.Marshal(obj)
			body = bytes.NewReader(data)
			contentType = firstNonEmpty(op.Payload.ContentType, "application/json")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return res, Error(CodeInvalidInput, "failed to build request: "+err.Error(), http.StatusBadRequest, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	// Static + templated headers.
	for _, h := range op.Headers {
		req.Header.Set(h.Name, substituteVars(h.Value, inputs))
	}
	// Header parameters.
	for _, p := range op.Parameters {
		if p.In == "header" {
			if v, ok := inputs[p.Name]; ok {
				req.Header.Set(p.Name, asString(v))
			}
		}
	}
	// Declared security scheme, filled from the reserved "_auth" input.
	if op.Security != nil {
		if token := asString(inputs["_auth"]); token != "" {
			switch strings.ToLower(op.Security.Scheme) {
			case "bearer", "oauth2":
				req.Header.Set("Authorization", "Bearer "+token)
			case "api-key", "apikey":
				name := firstNonEmpty(op.Security.Name, "X-API-Key")
				if op.Security.In == "query" {
					qq := req.URL.Query()
					qq.Set(name, token)
					req.URL.RawQuery = qq.Encode()
				} else {
					req.Header.Set(name, token)
				}
			case "basic":
				req.Header.Set("Authorization", "Basic "+token)
			}
		}
	}

	res.Request = method + " " + req.URL.String()
	started := time.Now()
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	res.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		return res, Error(CodeInternalError, "operation call failed: "+err.Error(), http.StatusBadGateway, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, invokeMaxBody))
	res.Status = resp.StatusCode
	res.ContentType = resp.Header.Get("Content-Type")
	res.Body = string(data)
	return res, nil
}

// asString renders an input value as a string for URL/header use.
func asString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

// substituteVars replaces {{name}} tokens in s with inputs[name].
func substituteVars(s string, inputs map[string]any) string {
	for k, v := range inputs {
		s = strings.ReplaceAll(s, "{{"+k+"}}", asString(v))
	}
	return s
}
