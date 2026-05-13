package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProcessAPIWorkflow(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)
	create := postJSON(h, "/api/v1/products/PB-ACC-MBX-001/processes", `{"id":"PRC-API-001","name":"API Process"}`)
	if create.Code != http.StatusCreated || !strings.Contains(create.Body.String(), "PRC-API-001") || !strings.Contains(create.Body.String(), "Task_ValidateRequest") {
		t.Fatalf("create process status=%d body=%s", create.Code, create.Body.String())
	}
	if rr := get(h, "/api/v1/processes/PRC-API-001/bpmn"); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "bpmn:definitions") {
		t.Fatalf("bpmn status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := get(h, "/api/v1/processes/PRC-API-001/tasks"); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "Task_PerformFulfillment") {
		t.Fatalf("tasks status=%d body=%s", rr.Code, rr.Body.String())
	}
	mapReq := postJSONMethod(h, http.MethodPut, "/api/v1/processes/PRC-API-001/task-mappings", `{"task_mappings":[{"bpmn_element_id":"Task_PerformFulfillment","service_ref":"identity.blumer.cloud/user-account","role":"primary","required":true}]}`)
	if mapReq.Code != http.StatusOK || !strings.Contains(mapReq.Body.String(), "identity.blumer.cloud/user-account") {
		t.Fatalf("mapping status=%d body=%s", mapReq.Code, mapReq.Body.String())
	}
	bad := postJSONMethod(h, http.MethodPut, "/api/v1/processes/PRC-API-001/bpmn", `<bad`)
	if bad.Code != http.StatusBadRequest || !strings.Contains(bad.Body.String(), "Invalid BPMN XML") {
		t.Fatalf("bad bpmn status=%d body=%s", bad.Code, bad.Body.String())
	}
}

func postJSONMethod(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}
