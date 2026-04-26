package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

func TestHandleTripMutationError_MapsTripStatusInvalidToInvalidParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	handleTripMutationError(ctx, service.ErrTripStatusInvalid)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v, body=%s", err, rec.Body.String())
	}
	if resp.Code != CodeInvalidParam {
		t.Fatalf("expected response code %d, got %d, body=%s", CodeInvalidParam, resp.Code, rec.Body.String())
	}
}
