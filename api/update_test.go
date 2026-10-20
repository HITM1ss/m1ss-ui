package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPanelUpdateRequiresLoginAndSameOrigin(t *testing.T) {
	engine := loginEngine(t)
	for _, action := range []struct{ method, path string }{
		{http.MethodGet, "/api/checkUpdate"},
		{http.MethodGet, "/api/updateStatus"},
		{http.MethodPost, "/api/update"},
	} {
		req := httptest.NewRequest(action.method, action.path, strings.NewReader("version=1.6.7"))
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("unauthenticated %s returned %d", action.path, recorder.Code)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "http://panel.test/api/update", strings.NewReader("version=1.6.7"))
	req.Header.Set("Origin", "https://another-site.test")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden {
		t.Errorf("cross-origin update returned %d", recorder.Code)
	}
}
