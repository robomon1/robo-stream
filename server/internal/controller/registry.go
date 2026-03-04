package controller

import (
	"fmt"
	"sync"

	"github.com/robomon1/robo-stream/server/internal/models"
)

// Registry holds all registered controllers and routes button actions to
// the correct one. It is safe for concurrent use.
//
// Controllers are always returned in the order they were first registered,
// regardless of intermediate unregister/re-register cycles (e.g. a plugin
// that crashes and restarts keeps its original slot in the list).
type Registry struct {
	controllers map[string]Controller
	order       []string // first-registration order; never shrinks
	mu          sync.RWMutex
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		controllers: make(map[string]Controller),
	}
}

// Register adds a controller. Returns an error if the ID is already present
// in the map (i.e. registered without a preceding Unregister call).
// Calling Register after Unregister for the same ID re-inserts the controller
// at its original position in the list.
func (r *Registry) Register(c Controller) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.controllers[c.ID()]; exists {
		return fmt.Errorf("controller %q already registered", c.ID())
	}
	r.controllers[c.ID()] = c
	// Append to order only on the very first registration; subsequent
	// re-registrations (after a crash/unregister) reuse the existing slot.
	found := false
	for _, id := range r.order {
		if id == c.ID() {
			found = true
			break
		}
	}
	if !found {
		r.order = append(r.order, c.ID())
	}
	return nil
}

// Unregister removes a controller from the active map but preserves its
// position in the order slice so that re-registration lands in the same slot.
func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.controllers, id)
	// Intentionally NOT removing from r.order — see Register.
}

// Get returns the controller with the given ID.
func (r *Registry) Get(id string) (Controller, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.controllers[id]
	return c, ok
}

// List returns all active controllers in stable registration order.
func (r *Registry) List() []Controller {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Controller, 0, len(r.controllers))
	for _, id := range r.order {
		if c, ok := r.controllers[id]; ok {
			result = append(result, c)
		}
	}
	return result
}

// ExecuteAction routes an action to the appropriate controller.
// If action.Controller is empty it defaults to "obs" for backwards
// compatibility with buttons created before the plugin system existed.
func (r *Registry) ExecuteAction(action models.ButtonAction) error {
	target := action.Controller
	if target == "" {
		target = "obs"
	}
	r.mu.RLock()
	c, ok := r.controllers[target]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("controller %q not found or not running", target)
	}
	return c.ExecuteAction(action)
}

// ShutdownAll calls Shutdown on every registered controller.
func (r *Registry) ShutdownAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, c := range r.controllers {
		if err := c.Shutdown(); err != nil {
			fmt.Printf("warning: error shutting down controller %q: %v\n", id, err)
		}
	}
}
