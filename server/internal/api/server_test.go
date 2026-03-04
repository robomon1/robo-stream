package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robomon1/robo-stream/server/internal/controller"
	obsctrl "github.com/robomon1/robo-stream/server/internal/controller/obs"
	"github.com/robomon1/robo-stream/server/internal/manager"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

type fakeController struct {
	id       string
	executed int
}

func (f *fakeController) ID() string          { return f.id }
func (f *fakeController) Name() string        { return "Fake" }
func (f *fakeController) Description() string { return "Fake test controller" }
func (f *fakeController) Version() string     { return "1.0.0" }
func (f *fakeController) Shutdown() error     { return nil }
func (f *fakeController) ExecuteAction(action models.ButtonAction) error {
	f.executed++
	return nil
}
func (f *fakeController) SupportedActionTypes() []controller.ActionTypeDefinition {
	return nil
}
func (f *fakeController) IsConnected() bool                                  { return true }
func (f *fakeController) GetStatus() map[string]interface{}                  { return map[string]interface{}{} }
func (f *fakeController) ComputeIndicator(action models.ButtonAction) string { return "" }
func (f *fakeController) GetConfigSchema() []controller.ConfigField          { return nil }
func (f *fakeController) GetCurrentConfig() map[string]interface{}           { return map[string]interface{}{} }
func (f *fakeController) SaveConfig(cfg map[string]interface{}) error        { return nil }
func (f *fakeController) GetDefaultButtons() []*models.Button                { return nil }

func newTestServer(t *testing.T) (*Server, *manager.SessionManager, *fakeController) {
	t.Helper()

	dataDir := t.TempDir()
	st, err := storage.New(dataDir)
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}

	bm := manager.NewButtonManager(st)
	cm := manager.NewConfigManager(st, bm)
	sm := manager.NewSessionManager(st)
	reg := controller.NewRegistry()

	obs := obsctrl.New(st)
	if err := reg.Register(obs); err != nil {
		t.Fatalf("register obs failed: %v", err)
	}

	fake := &fakeController{id: "test"}
	if err := reg.Register(fake); err != nil {
		t.Fatalf("register fake failed: %v", err)
	}

	s := NewServer(cm, sm, reg, obs, nil)
	return s, sm, fake
}

func TestExecuteActionRejectsUnknownSession(t *testing.T) {
	s, _, fake := newTestServer(t)

	body, err := json.Marshal(models.ButtonAction{Controller: "test", Type: "do_thing"})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(body))
	req.Header.Set("X-Session-ID", "missing-session")

	rec := httptest.NewRecorder()
	s.executeAction(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
	if fake.executed != 0 {
		t.Fatalf("action executed for invalid session")
	}
}

func TestExecuteActionAcceptsValidSession(t *testing.T) {
	s, sm, fake := newTestServer(t)

	session, err := sm.RegisterOrUpdate("client-1", "Test Client", "default", "127.0.0.1")
	if err != nil {
		t.Fatalf("RegisterOrUpdate failed: %v", err)
	}

	body, err := json.Marshal(models.ButtonAction{Controller: "test", Type: "do_thing"})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(body))
	req.Header.Set("X-Session-ID", session.SessionID)

	rec := httptest.NewRecorder()
	s.executeAction(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	if fake.executed != 1 {
		t.Fatalf("expected action to execute once, got %d", fake.executed)
	}
}
