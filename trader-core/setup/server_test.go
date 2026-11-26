package setup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpointUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := InitServer(Config{ApiKey: "secret", Env: "test"})

	// /health must NOT require the API key (load balancers / docker healthchecks)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /health = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body = %v", body)
	}

	// /api/* still requires the key
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/metrics without key = %d, want 401", w2.Code)
	}
	w3 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	req.Header.Set("Authorization", "ApiKey secret")
	r.ServeHTTP(w3, req)
	if w3.Code != http.StatusOK {
		t.Fatalf("GET /api/metrics with key = %d, want 200; body: %s", w3.Code, w3.Body.String())
	}
}
