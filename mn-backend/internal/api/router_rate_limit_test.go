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

	for i := 0; i < 10; i++ {
		rec := sendUserLoginRequest(r)

		var resp controller.ResponseData
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal login response: %v, body=%s", err, rec.Body.String())
		}
		if resp.Code != controller.CodeInvalidPassword {
			t.Fatalf("expected login %d to reach controller, got %d, body=%s", i+1, resp.Code, rec.Body.String())
		}
	}

	blockedRec := sendUserLoginRequest(r)

	if blockedRec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, blockedRec.Code, blockedRec.Body.String())
	}

	var blockedResp controller.ResponseData
	if err := json.Unmarshal(blockedRec.Body.Bytes(), &blockedResp); err != nil {
		t.Fatalf("unmarshal blocked response: %v, body=%s", err, blockedRec.Body.String())
	}
	if blockedResp.Code != controller.CodeInvalidParam {
		t.Fatalf("expected eleventh login to be rate limited, got %d, body=%s", blockedResp.Code, blockedRec.Body.String())
	}
	if blockedResp.Message != "请勿频繁操作" {
		t.Fatalf("expected rate limit message, got %q", blockedResp.Message)
	}
}

func sendUserLoginRequest(r http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"wrong-password"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}
