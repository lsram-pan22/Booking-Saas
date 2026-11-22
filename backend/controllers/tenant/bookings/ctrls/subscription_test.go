package ctrls

import (
	"booking-saas/utils"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSubscriptionDenialIsNotSessionExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	respondError(ctx, utils.ErrSubscriptionRequired)
	if recorder.Code != 403 || !strings.Contains(recorder.Body.String(), `"code":"SUBSCRIPTION_REQUIRED"`) {
		t.Fatalf("unexpected denial: %d %s", recorder.Code, recorder.Body.String())
	}
}
