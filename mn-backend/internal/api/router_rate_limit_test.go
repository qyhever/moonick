package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"moonick/internal/controller"
)

func TestSetupRouter_UserLoginRejectsFrequentRequestsFromSameIP(t *testing.T) {
	restore := useValidJWTConfig(t)
	defer restore()

	r := SetupRouter()

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"wrong-password"}`))
	firstReq.Header.Set("Content-Type", "application/json")
	firstReq.Header.Set("X-Forwarded-For", "203.0.113.10")
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)

	var firstResp controller.ResponseData
	if err := json.Unmarshal(firstRec.Body.Bytes(), &firstResp); err != nil {
		t.Fatalf("unmarshal first response: %v, body=%s", err, firstRec.Body.String())
	}
	if firstResp.Code != controller.CodeInvalidPassword {
		t.Fatalf("expected first login to reach controller, got %d, body=%s", firstResp.Code, firstRec.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"wrong-password"}`))
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.Header.Set("X-Forwarded-For", "203.0.113.10")
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, secondRec.Code, secondRec.Body.String())
	}

	var secondResp controller.ResponseData
	if err := json.Unmarshal(secondRec.Body.Bytes(), &secondResp); err != nil {
		t.Fatalf("unmarshal second response: %v, body=%s", err, secondRec.Body.String())
	}
	if secondResp.Code != controller.CodeInvalidParam {
		t.Fatalf("expected second login to be rate limited, got %d, body=%s", secondResp.Code, secondRec.Body.String())
	}
	if secondResp.Message != "请勿频繁操作" {
		t.Fatalf("expected rate limit message, got %q", secondResp.Message)
	}
}
