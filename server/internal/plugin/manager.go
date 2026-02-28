// Package plugin manages external controller plugins.
// Each plugin is a standalone binary that implements the Robo-Stream plugin
// protocol (defined in the sdk package). The manager discovers binaries in
// {dataDir}/plugins/, spawns them as subprocesses, and registers them in the
// controller registry so the rest of the server can route actions to them.
package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/robomon1/robo-stream/server/internal/controller"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

// pluginReadyMsg is the JSON payload printed to stdout by a plugin binary
// once it is listening and ready for connections.
type pluginReadyMsg struct {
	ID      string `json:"id"`
	Port    int    `json:"port"`
	Version string `json:"version"`
}

// PluginProcess represents a running plugin subprocess.
type PluginProcess struct {
	ID      string
	cmd     *exec.Cmd
	port    int
	ctrl    *HostController
	stopCh  chan struct{}
	execPath string
}

// Manager discovers, starts, and monitors plugin subprocesses.
// It is safe for concurrent use.
type Manager struct {
	pluginsDir string
	storage    *storage.Storage
	registry   *controller.Registry
	processes  map[string]*PluginProcess
	mu         sync.RWMutex
}

// New creates a plugin manager. pluginsDir is the root directory scanned for
// plugin subdirectories (typically {dataDir}/plugins/).
func New(pluginsDir string, st *storage.Storage, reg *controller.Registry) *Manager {
	return &Manager{
		pluginsDir: pluginsDir,
		storage:    st,
		registry:   reg,
		processes:  make(map[string]*PluginProcess),
	}
}

// DiscoverAndStart scans pluginsDir for plugin subdirectories and starts any
// found that are not already running. Errors for individual plugins are logged
// but do not prevent other plugins from starting.
func (m *Manager) DiscoverAndStart() error {
	if err := os.MkdirAll(m.pluginsDir, 0755); err != nil {
		return fmt.Errorf("plugin: failed to create plugins directory: %w", err)
	}

	entries, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		return fmt.Errorf("plugin: failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginDir := filepath.Join(m.pluginsDir, entry.Name())
		execPath, err := findExecutable(pluginDir)
		if err != nil {
			log.Printf("plugin: skipping %s: %v", entry.Name(), err)
			continue
		}
		if err := m.startPlugin(pluginDir, execPath); err != nil {
			log.Printf("plugin: failed to start %s: %v", entry.Name(), err)
		}
	}
	return nil
}

// Install copies a binary into pluginsDir/{pluginID}/ and starts it.
// srcPath is the path to the binary on the local filesystem.
func (m *Manager) Install(srcPath string) error {
	// Probe the binary to discover its plugin ID
	id, err := probePluginID(srcPath)
	if err != nil {
		return fmt.Errorf("plugin: cannot read plugin info from %s: %w", srcPath, err)
	}

	destDir := filepath.Join(m.pluginsDir, id)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, filepath.Base(srcPath))
	if err := copyFile(srcPath, destPath); err != nil {
		return fmt.Errorf("plugin: failed to copy binary: %w", err)
	}
	if err := os.Chmod(destPath, 0755); err != nil {
		return err
	}

	return m.startPlugin(destDir, destPath)
}

// Uninstall stops the plugin subprocess and removes its directory.
func (m *Manager) Uninstall(id string) error {
	m.mu.Lock()
	proc, ok := m.processes[id]
	if ok {
		close(proc.stopCh)
		if proc.cmd.Process != nil {
			proc.cmd.Process.Signal(os.Interrupt)
			// Give it a moment to exit cleanly
			time.Sleep(500 * time.Millisecond)
			proc.cmd.Process.Kill()
		}
		delete(m.processes, id)
	}
	m.mu.Unlock()

	m.registry.Unregister(id)

	pluginDir := filepath.Join(m.pluginsDir, id)
	return os.RemoveAll(pluginDir)
}

// Restart stops a running plugin and starts it again.
func (m *Manager) Restart(id string) error {
	m.mu.RLock()
	proc, ok := m.processes[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("plugin %q not found", id)
	}
	execPath := proc.execPath
	pluginDir := filepath.Dir(execPath)

	if err := m.stopProcess(id); err != nil {
		log.Printf("plugin: error stopping %s for restart: %v", id, err)
	}
	return m.startPlugin(pluginDir, execPath)
}

// List returns info about all running plugins.
func (m *Manager) List() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]map[string]interface{}, 0, len(m.processes))
	for id, proc := range m.processes {
		result = append(result, map[string]interface{}{
			"id":   id,
			"port": proc.port,
		})
	}
	return result
}

// Shutdown stops all plugin subprocesses.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.processes))
	for id := range m.processes {
		ids = append(ids, id)
	}
	m.mu.Unlock()

	for _, id := range ids {
		m.stopProcess(id)
	}
}

// UpdateConfig saves new config for a plugin and calls /initialize to apply it.
func (m *Manager) UpdateConfig(id string, cfg map[string]interface{}) error {
	m.mu.RLock()
	proc, ok := m.processes[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("plugin %q not running", id)
	}
	return proc.ctrl.SaveConfig(cfg)
}

// ==================== internal helpers ====================

func (m *Manager) startPlugin(pluginDir, execPath string) error {
	cmd := exec.Command(execPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", execPath, err)
	}

	// Drain stderr to the log continuously
	go drainToLog(stderr, "[plugin/stderr]")

	// Wait for PLUGIN_READY on stdout
	msg, remainingStdout, err := readPluginReady(stdout, 10*time.Second)
	if err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("waiting for PLUGIN_READY from %s: %w", execPath, err)
	}

	log.Printf("plugin: %s v%s started on port %d", msg.ID, msg.Version, msg.Port)

	// Load persisted config
	cfg := m.loadConfig(msg.ID)

	ctrl := newHostController(msg.ID, msg.Port, filepath.Join(m.pluginsDir, msg.ID), m.storage)

	// Initialize the plugin with its saved config
	if err := ctrl.initialize(cfg); err != nil {
		log.Printf("plugin: %s initialize failed: %v (plugin is running but may not be connected)", msg.ID, err)
	}

	stopCh := make(chan struct{})
	proc := &PluginProcess{
		ID:       msg.ID,
		cmd:      cmd,
		port:     msg.Port,
		ctrl:     ctrl,
		stopCh:   stopCh,
		execPath: execPath,
	}

	m.mu.Lock()
	m.processes[msg.ID] = proc
	m.mu.Unlock()

	// Register in the controller registry
	if err := m.registry.Register(ctrl); err != nil {
		log.Printf("plugin: %s already registered, replacing", msg.ID)
		m.registry.Unregister(msg.ID)
		m.registry.Register(ctrl)
	}

	// Drain remaining stdout to log
	go drainToLog(remainingStdout, fmt.Sprintf("[plugin/%s]", msg.ID))

	// Monitor for unexpected exit and restart
	go m.monitorProcess(proc)

	return nil
}

func (m *Manager) stopProcess(id string) error {
	m.mu.Lock()
	proc, ok := m.processes[id]
	if ok {
		close(proc.stopCh)
		delete(m.processes, id)
	}
	m.mu.Unlock()

	if !ok {
		return nil
	}

	m.registry.Unregister(id)

	if proc.cmd.Process != nil {
		proc.cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() {
			proc.cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			proc.cmd.Process.Kill()
		}
	}
	return nil
}

func (m *Manager) monitorProcess(proc *PluginProcess) {
	doneCh := make(chan error, 1)
	go func() { doneCh <- proc.cmd.Wait() }()

	select {
	case <-proc.stopCh:
		// Normal shutdown, do nothing
		return
	case err := <-doneCh:
		log.Printf("plugin: %s exited unexpectedly: %v — will restart in 10s", proc.ID, err)
	}

	// Clean up registry entry
	m.mu.Lock()
	delete(m.processes, proc.ID)
	m.mu.Unlock()
	m.registry.Unregister(proc.ID)

	// Restart after backoff
	select {
	case <-proc.stopCh:
		return
	case <-time.After(10 * time.Second):
	}

	log.Printf("plugin: restarting %s", proc.ID)
	if err := m.startPlugin(filepath.Dir(proc.execPath), proc.execPath); err != nil {
		log.Printf("plugin: failed to restart %s: %v", proc.ID, err)
	}
}

func (m *Manager) loadConfig(id string) map[string]interface{} {
	var cfg map[string]interface{}
	filename := filepath.Join(id, "config.json")
	m.storage.LoadJSON(filename, &cfg)
	if cfg == nil {
		cfg = make(map[string]interface{})
	}
	return cfg
}

// ==================== package-level helpers ====================

// findExecutable scans dir for an executable file and returns its path.
func findExecutable(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		if runtime.GOOS == "windows" {
			if strings.HasSuffix(strings.ToLower(name), ".exe") {
				return fullPath, nil
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0111 != 0 {
				return fullPath, nil
			}
		}
	}
	return "", fmt.Errorf("no executable found in %s", dir)
}

// probePluginID starts the binary with --probe flag and reads the plugin ID.
// This is used during Install before the plugin has a home directory.
// The plugin binary should print PLUGIN_READY and then exit when --probe is passed.
func probePluginID(execPath string) (string, error) {
	cmd := exec.Command(execPath, "--probe")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	defer cmd.Process.Kill()

	msg, _, err := readPluginReady(stdout, 5*time.Second)
	if err != nil {
		return "", err
	}
	return msg.ID, nil
}

// readPluginReady reads lines from r until it finds a PLUGIN_READY line or
// the timeout expires. It returns the parsed message and an io.Reader for the
// remaining (unread) bytes of r so the caller can continue draining stdout.
func readPluginReady(r io.Reader, timeout time.Duration) (*pluginReadyMsg, io.Reader, error) {
	type result struct {
		msg *pluginReadyMsg
		err error
	}

	// pipe so we can return remaining output
	pr, pw := io.Pipe()
	resultCh := make(chan result, 1)

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PLUGIN_READY ") {
				jsonStr := strings.TrimPrefix(line, "PLUGIN_READY ")
				var msg pluginReadyMsg
				if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
					resultCh <- result{nil, fmt.Errorf("invalid PLUGIN_READY payload: %w", err)}
					pw.Close()
					return
				}
				resultCh <- result{&msg, nil}
				// Forward remaining stdout through the pipe
				for scanner.Scan() {
					pw.Write(scanner.Bytes())
					pw.Write([]byte("\n"))
				}
				pw.Close()
				return
			}
			// Pre-ready log output — forward it too
			pw.Write(scanner.Bytes())
			pw.Write([]byte("\n"))
		}
		resultCh <- result{nil, fmt.Errorf("plugin exited without sending PLUGIN_READY")}
		pw.Close()
	}()

	select {
	case res := <-resultCh:
		return res.msg, pr, res.err
	case <-time.After(timeout):
		pr.Close()
		return nil, nil, fmt.Errorf("timed out waiting for PLUGIN_READY")
	}
}

func drainToLog(r io.Reader, prefix string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("%s %s", prefix, scanner.Text())
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
