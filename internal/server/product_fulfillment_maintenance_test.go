package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductFulfillmentMaintenanceEndpoints(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	create := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/blueprints", strings.NewReader(`{"id":"PROD-MAINT-API-001","type":"product_blueprint","name":"Maintenance API","version":"0.1.0","status":"draft","owner":"Team","summary":"API maintenance"}`))
	createReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("create product status=%d body=%s", create.Code, create.Body.String())
	}
	add := httptest.NewRecorder()
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/products/PROD-MAINT-API-001/fulfillment-services", strings.NewReader(`{"service_ref":"identity.blumer.cloud/user-account","role":"primary","required":true,"description":"Original"}`))
	addReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(add, addReq)
	if add.Code != http.StatusOK {
		t.Fatalf("add fulfillment status=%d body=%s", add.Code, add.Body.String())
	}

	edit := httptest.NewRecorder()
	editReq := httptest.NewRequest(http.MethodPut, "/api/v1/products/PROD-MAINT-API-001/fulfillment-services/0", strings.NewReader(`{"service_ref":"collaboration.blumer.cloud/mailbox","role":"supporting","required":false,"description":"Edited mailbox"}`))
	editReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(edit, editReq)
	if edit.Code != http.StatusOK || !strings.Contains(edit.Body.String(), "collaboration.blumer.cloud/mailbox") || !strings.Contains(edit.Body.String(), "Edited mailbox") || !strings.Contains(edit.Body.String(), `"required":false`) {
		t.Fatalf("edit fulfillment status=%d body=%s", edit.Code, edit.Body.String())
	}

	dupAdd := httptest.NewRecorder()
	dupAddReq := httptest.NewRequest(http.MethodPost, "/api/v1/products/PROD-MAINT-API-001/fulfillment-services", strings.NewReader(`{"service_ref":"identity.blumer.cloud/user-account","role":"primary","required":true}`))
	dupAddReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(dupAdd, dupAddReq)
	if dupAdd.Code != http.StatusOK {
		t.Fatalf("add second fulfillment status=%d body=%s", dupAdd.Code, dupAdd.Body.String())
	}
	dupEdit := httptest.NewRecorder()
	dupEditReq := httptest.NewRequest(http.MethodPut, "/api/v1/products/PROD-MAINT-API-001/fulfillment-services/1", strings.NewReader(`{"service_ref":"collaboration.blumer.cloud/mailbox","role":"primary","required":true}`))
	dupEditReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(dupEdit, dupEditReq)
	if dupEdit.Code != http.StatusConflict {
		t.Fatalf("expected duplicate edit conflict, got %d body=%s", dupEdit.Code, dupEdit.Body.String())
	}

	del := httptest.NewRecorder()
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/products/PROD-MAINT-API-001/fulfillment-services/0", nil)
	h.ServeHTTP(del, delReq)
	if del.Code != http.StatusOK || strings.Contains(del.Body.String(), "collaboration.blumer.cloud/mailbox") {
		t.Fatalf("delete fulfillment status=%d body=%s", del.Code, del.Body.String())
	}
	invalid := httptest.NewRecorder()
	invalidReq := httptest.NewRequest(http.MethodDelete, "/api/v1/products/PROD-MAINT-API-001/fulfillment-services/9", nil)
	h.ServeHTTP(invalid, invalidReq)
	if invalid.Code != http.StatusNotFound {
		t.Fatalf("expected invalid index not found, got %d body=%s", invalid.Code, invalid.Body.String())
	}
}
