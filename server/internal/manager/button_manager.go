package manager

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

// ButtonManager manages the button library
type ButtonManager struct {
	storage *storage.Storage
	buttons map[string]*models.Button
}

// NewButtonManager creates a new ButtonManager
func NewButtonManager(storage *storage.Storage) *ButtonManager {
	bm := &ButtonManager{
		storage: storage,
		buttons: make(map[string]*models.Button),
	}
	bm.load()
	return bm
}

// load reads buttons from storage
func (bm *ButtonManager) load() error {
	var buttons []*models.Button
	if err := bm.storage.LoadJSON("buttons.json", &buttons); err != nil {
		return err
	}
	for _, btn := range buttons {
		bm.buttons[btn.ID] = btn
	}
	return nil
}

// save writes buttons to storage
func (bm *ButtonManager) save() error {
	buttons := make([]*models.Button, 0, len(bm.buttons))
	for _, btn := range bm.buttons {
		buttons = append(buttons, btn)
	}
	return bm.storage.SaveJSON("buttons.json", buttons)
}

// Create creates a new button
func (bm *ButtonManager) Create(btn *models.Button) error {
	btn.ID = uuid.New().String()
	btn.CreatedAt = time.Now()
	btn.UpdatedAt = time.Now()
	bm.buttons[btn.ID] = btn
	return bm.save()
}

// Get retrieves a button by ID
func (bm *ButtonManager) Get(id string) (*models.Button, error) {
	btn, ok := bm.buttons[id]
	if !ok {
		return nil, fmt.Errorf("button not found: %s", id)
	}
	return btn, nil
}

// List returns all buttons
func (bm *ButtonManager) List() []*models.Button {
	buttons := make([]*models.Button, 0, len(bm.buttons))
	for _, btn := range bm.buttons {
		buttons = append(buttons, btn)
	}
	return buttons
}

// Update updates an existing button
func (bm *ButtonManager) Update(btn *models.Button) error {
	if _, ok := bm.buttons[btn.ID]; !ok {
		return fmt.Errorf("button not found: %s", btn.ID)
	}
	btn.UpdatedAt = time.Now()
	bm.buttons[btn.ID] = btn
	return bm.save()
}

// Delete removes a button
func (bm *ButtonManager) Delete(id string) error {
	delete(bm.buttons, id)
	return bm.save()
}

// Search finds buttons matching a query
func (bm *ButtonManager) Search(query string) []*models.Button {
	// Simple search implementation
	results := make([]*models.Button, 0)
	// TODO: Implement proper search logic
	return results
}
