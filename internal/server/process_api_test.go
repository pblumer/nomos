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
	if create.Code != http.StatusCreated || !strings.Contains(create.Body.String(), "PRC-API-001") {
		t.Fatalf("create process status=%d body=%s", create.Code, create.Body.String())
	}
	if rr := get(h, "/api/v1/processes/PRC-API-001/bpmn"); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "bpmn:definitions") {
		t.Fatalf("bpmn status=%d body=%s", rr.Code, rr.Body.String())
	}
	// New processes have a blank BPMN; tasks list starts empty.
	if rr := get(h, "/api/v1/processes/PRC-API-001/tasks"); rr.Code != http.StatusOK {
		t.Fatalf("tasks status=%d body=%s", rr.Code, rr.Body.String())
	}
	// Add a custom BPMN task so we can test mappings.
	customBPMN := `<?xml version="1.0" encoding="UTF-8"?><bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Defs" targetNamespace="https://nomos.local/bpmn"><bpmn:process id="Process_prc_api_001" isExecutable="false"><bpmn:startEvent id="Start"><bpmn:outgoing>F1</bpmn:outgoing></bpmn:startEvent><bpmn:task id="Task_Custom" name="Custom"><bpmn:incoming>F1</bpmn:incoming><bpmn:outgoing>F2</bpmn:outgoing></bpmn:task><bpmn:endEvent id="End"><bpmn:incoming>F2</bpmn:incoming></bpmn:endEvent><bpmn:sequenceFlow id="F1" sourceRef="Start" targetRef="Task_Custom"/><bpmn:sequenceFlow id="F2" sourceRef="Task_Custom" targetRef="End"/></bpmn:process><bpmndi:BPMNDiagram id="D1"><bpmndi:BPMNPlane id="P1" bpmnElement="Process_prc_api_001"><bpmndi:BPMNShape id="S1" bpmnElement="Start"><dc:Bounds x="100" y="100" width="36" height="36"/></bpmndi:BPMNShape><bpmndi:BPMNShape id="S2" bpmnElement="Task_Custom"><dc:Bounds x="200" y="80" width="100" height="80"/></bpmndi:BPMNShape><bpmndi:BPMNShape id="S3" bpmnElement="End"><dc:Bounds x="360" y="100" width="36" height="36"/></bpmndi:BPMNShape><bpmndi:BPMNEdge id="E1" bpmnElement="F1"><di:waypoint x="136" y="118"/><di:waypoint x="200" y="120"/></bpmndi:BPMNEdge><bpmndi:BPMNEdge id="E2" bpmnElement="F2"><di:waypoint x="300" y="120"/><di:waypoint x="360" y="118"/></bpmndi:BPMNEdge></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
	if rr := postJSONMethod(h, http.MethodPut, "/api/v1/processes/PRC-API-001/bpmn", customBPMN); rr.Code != http.StatusOK {
		t.Fatalf("bpmn update status=%d body=%s", rr.Code, rr.Body.String())
	}
	mapReq := postJSONMethod(h, http.MethodPut, "/api/v1/processes/PRC-API-001/task-mappings", `{"task_mappings":[{"bpmn_element_id":"Task_Custom","service_ref":"identity.blumer.cloud/user-account","method":"deactivateUser","role":"primary","required":true}]}`)
	if mapReq.Code != http.StatusOK || !strings.Contains(mapReq.Body.String(), "identity.blumer.cloud/user-account") || !strings.Contains(mapReq.Body.String(), "deactivateUser") {
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
