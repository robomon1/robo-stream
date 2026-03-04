package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// PluginImplementation is the interface that plugin authors implement.
// Call RunPlugin() with a value that satisfies this interface to start the plugin.
type PluginImplementation interface {
	// Info returns static metadata. Called once on startup.
	Info() PluginInfo

	// Initialize is called by the host after it reads PLUGIN_READY.
	// cfg contains the persisted configuration values (e.g. connection credentials).
	// The plugin should connect to its target service here.
	Initialize(cfg map[string]interface{}) error

	// Status returns the current connection state.
	Status() PluginStatus

	// ConfigSchema returns the fields the host should show in its settings UI.
	ConfigSchema() []ConfigField

	// GetConfig returns the current configuration values.
	GetConfig() map[string]interface{}

	// UpdateConfig saves new configuration values and reconnects if needed.
	UpdateConfig(cfg map[string]interface{}) error

	// SupportedActions returns every action type this plugin can execute.
	SupportedActions() []ActionTypeDef

	// Execute runs an action.
	Execute(action ExecuteRequest) error

	// DefaultButtons returns buttons the host should add to its library
	// when this plugin is first installed.
	DefaultButtons() []Button
}

// RunPlugin starts the HTTP protocol server for a plugin and blocks until
// SIGTERM or SIGINT is received. Plugin authors call this from main().
func RunPlugin(impl PluginImplementation) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("plugin: failed to bind: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	info := impl.Info()

	mux := http.NewServeMux()
	registerHandlers(mux, impl)

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("plugin %s: server error: %v", info.ID, err)
		}
	}()

	// Signal readiness to the host. The host reads this line from stdout.
	fmt.Printf("PLUGIN_READY {\"id\":%q,\"port\":%d,\"version\":%q}\n", info.ID, port, info.Version)
	os.Stdout.Sync()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Printf("plugin %s: shutting down", info.ID)
}

func registerHandlers(mux *http.ServeMux, impl PluginImplementation) {
	respond := func(w http.ResponseWriter, status int, v interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(v)
	}
	respondErr := func(w http.ResponseWriter, status int, msg string) {
		respond(w, status, ErrorResponse{Error: msg})
	}

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, impl.Info())
	})

	mux.HandleFunc("/initialize", func(w http.ResponseWriter, r *http.Request) {
		var req InitializeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := impl.Initialize(req.Config); err != nil {
			respondErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		respond(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, impl.Status())
	})

	mux.HandleFunc("/config/schema", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, impl.ConfigSchema())
	})

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			respond(w, http.StatusOK, impl.GetConfig())
		case http.MethodPost:
			var cfg map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				respondErr(w, http.StatusBadRequest, "invalid request")
				return
			}
			if err := impl.UpdateConfig(cfg); err != nil {
				respondErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			respond(w, http.StatusOK, map[string]bool{"ok": true})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/actions", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, impl.SupportedActions())
	})

	mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		var req ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := impl.Execute(req); err != nil {
			respondErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		respond(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("/buttons", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, impl.DefaultButtons())
	})
}
