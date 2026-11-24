package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
)

func setupAccountRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	m := engine.NewAccountManager(nil, engine.NewLiveAccount("0", "0.001")) // in-memory
	origAccounts, origAccount := accounts, account
	accounts = m
	account = mustDefaultPaper(t, m)
	t.Cleanup(func() { accounts, account = origAccounts, origAccount })

	r := gin.New()
	RegisterAccountRoutes(r.Group("/api/account"))
	return r
}

func mustDefaultPaper(t *testing.T, m *engine.AccountManager) engine.Account {
	t.Helper()
	acc, err := m.Paper("default")
	if err != nil {
		t.Fatalf("default paper: %v", err)
	}
	return acc
}

func doJSON(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAccountCreateListGet(t *testing.T) {
	r := setupAccountRouter(t)

	w := doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"acc1","balance":"500","fee":"0.002"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}

	// duplicate -> 409
	w = doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"acc1"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("dup: %d %s", w.Code, w.Body.String())
	}

	// bad body -> 400
	w = doJSON(t, r, http.MethodPost, "/api/account/paper", `{"balance":"5"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing name: %d", w.Code)
	}
	w = doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"x","balance":"abc"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad balance: %d", w.Code)
	}

	// list contains acc1 + default + live
	w = doJSON(t, r, http.MethodGet, "/api/account/all", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list: %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{`"acc1"`, `"default"`, `"live"`, `"500"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("list body missing %s: %s", want, body)
		}
	}

	// get one
	w = doJSON(t, r, http.MethodGet, "/api/account/acc1", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"500"`) {
		t.Fatalf("get acc1: %d %s", w.Code, w.Body.String())
	}

	// unknown -> 404
	w = doJSON(t, r, http.MethodGet, "/api/account/nope", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown: %d", w.Code)
	}
}

func TestAccountTopUpAndReset(t *testing.T) {
	r := setupAccountRouter(t)
	doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"acc2","balance":"100"}`)

	w := doJSON(t, r, http.MethodPost, "/api/account/paper/acc2/topup", `{"amount":"25.5"}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"125.5"`) {
		t.Fatalf("topup: %d %s", w.Code, w.Body.String())
	}

	// invalid amount -> 400
	w = doJSON(t, r, http.MethodPost, "/api/account/paper/acc2/topup", `{"amount":"-5"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("neg topup: %d", w.Code)
	}
	// unknown account -> 404
	w = doJSON(t, r, http.MethodPost, "/api/account/paper/ghost/topup", `{"amount":"5"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("topup unknown: %d", w.Code)
	}

	// reset
	w = doJSON(t, r, http.MethodPost, "/api/account/paper/acc2/reset", `{"balance":"1000"}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"1000"`) {
		t.Fatalf("reset: %d %s", w.Code, w.Body.String())
	}
}

func TestLegacyAccountRoute(t *testing.T) {
	r := setupAccountRouter(t)
	w := doJSON(t, r, http.MethodGet, "/api/account", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"10000"`) {
		t.Fatalf("legacy: %d %s", w.Code, w.Body.String())
	}
}
