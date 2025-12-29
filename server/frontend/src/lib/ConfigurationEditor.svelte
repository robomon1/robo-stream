<script>
  import { onMount } from 'svelte';
  import ConfigModal from './ConfigModal.svelte';
  import ButtonModal from './ButtonModal.svelte';

  let configurations = [];
  let buttons = [];
  let selectedConfig = null;
  let editMode = false;
  let loading = true;
  let showConfigModal = false;
  let showButtonModal = false;
  let editingConfig = null;
  let editingButton = null;
  let draggedButton = null;

  // Reinitialize icons when edit mode changes or buttons change
  $: if (editMode !== undefined || buttons.length) {
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  }

  onMount(async () => {
    await loadData();
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  });

  async function loadData() {
    try {
      [configurations, buttons] = await Promise.all([
        window.go.main.App.GetConfigurations(),
        window.go.main.App.GetButtons()
      ]);
      
      console.log('Loaded configurations:', configurations);
      console.log('Loaded buttons:', buttons);
      
      // Select first config by default
      if (configurations.length > 0 && !selectedConfig) {
        selectedConfig = configurations[0];
      }
    } catch (err) {
      console.error('Failed to load data:', err);
    } finally {
      loading = false;
    }
  }

  function selectConfig(config) {
    selectedConfig = config;
    editMode = false;
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  }

  function createConfiguration() {
    editingConfig = null;
    showConfigModal = true;
  }

  function editConfiguration() {
    if (!selectedConfig) return;
    editingConfig = selectedConfig;
    showConfigModal = true;
  }

  async function duplicateConfiguration() {
    if (!selectedConfig) return;
    
    try {
      const copy = {
        name: selectedConfig.name + ' (Copy)',
        description: selectedConfig.description,
        grid: { ...selectedConfig.grid },
        buttons: { ...selectedConfig.buttons },
        is_default: false  // Duplicates are never default - user must explicitly set it
      };
      
      await window.go.main.App.CreateConfiguration(copy);
      await loadData();
      
      setTimeout(() => {
        if (window.lucide) lucide.createIcons();
      }, 100);
    } catch (err) {
      console.error('Failed to duplicate configuration:', err);
      alert('Error: ' + err);
    }
  }

  async function deleteConfiguration() {
    if (!selectedConfig) return;
    
    // Protect the hard-coded default configuration
    if (selectedConfig.is_default && selectedConfig.name === 'Default') {
      console.log('Cannot delete the hard-coded default configuration');
      return;
    }
    
    // Note: confirm() doesn't work in Wails, so we'll just delete
    // In production, you might want a custom modal for confirmation
    console.log('Deleting configuration:', selectedConfig.name);
    
    try {
      await window.go.main.App.DeleteConfiguration(selectedConfig.id);
      selectedConfig = null;
      await loadData();
    } catch (err) {
      console.error('Failed to delete configuration:', err);
      alert('Error: ' + err);
    }
  }

  async function setDefault() {
    if (!selectedConfig) return;
    
    try {
      await window.go.main.App.SetDefaultConfiguration(selectedConfig.id);
      await loadData();
    } catch (err) {
      console.error('Failed to set default:', err);
      alert('Error: ' + err);
    }
  }

  async function handleSave(configData) {
    try {
      if (editingConfig) {
        await window.go.main.App.UpdateConfiguration(configData);
      } else {
        await window.go.main.App.CreateConfiguration(configData);
      }
      
      showConfigModal = false;
      editingConfig = null;
      await loadData();
      
      setTimeout(() => {
        if (window.lucide) lucide.createIcons();
      }, 100);
    } catch (err) {
      console.error('Failed to save configuration:', err);
      alert('Error: ' + err);
    }
  }

  function closeConfigModal() {
    showConfigModal = false;
    editingConfig = null;
  }

  // Button modal functions
  function createButton() {
    editingButton = null;
    showButtonModal = true;
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  }

  function editButton(button) {
    editingButton = button;
    showButtonModal = true;
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  }

  async function handleButtonSave(buttonData) {
    try {
      if (editingButton) {
        console.log('Updating button:', buttonData);
        await window.go.main.App.UpdateButton(buttonData);
      } else {
        console.log('Creating button:', buttonData);
        await window.go.main.App.CreateButton(buttonData);
      }
      
      showButtonModal = false;
      editingButton = null;
      await loadData();
      
      setTimeout(() => {
        if (window.lucide) lucide.createIcons();
      }, 100);
    } catch (err) {
      console.error('Failed to save button:', err);
      alert('Error: ' + err);
    }
  }

  function closeButtonModal() {
    showButtonModal = false;
    editingButton = null;
  }

  async function deleteButton(button) {
    console.log('Deleting button:', button.id, button.name);
    try {
      await window.go.main.App.DeleteButton(button.id);
      await loadData();
      
      setTimeout(() => {
        if (window.lucide) lucide.createIcons();
      }, 100);
    } catch (err) {
      console.error('Failed to delete button:', err);
      alert('Error: ' + err);
    }
  }

  function toggleEditMode() {
    editMode = !editMode;
  }

  function handleDragStart(event, button) {
    draggedButton = button;
    event.dataTransfer.effectAllowed = 'copy';
  }

  function handleDragEnd() {
    draggedButton = null;
  }

  function handleDrop(event, row, col) {
    event.preventDefault();
    if (!draggedButton || !selectedConfig || !editMode) return;
    
    const position = `btn-${row}-${col}`;
    console.log('Dropping button', draggedButton.id, 'at position', position);
    
    // Update configuration
    selectedConfig.buttons[position] = draggedButton.id;
    
    // Save to backend
    saveConfigurationButtons();
    
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  }

  function handleDragOver(event) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }

  function removeButton(position) {
    if (!selectedConfig || !editMode) return;
    
    delete selectedConfig.buttons[position];
    saveConfigurationButtons();
    
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  }

  async function saveConfigurationButtons() {
    try {
      await window.go.main.App.UpdateConfiguration(selectedConfig);
      console.log('Saved configuration buttons');
    } catch (err) {
      console.error('Failed to save configuration:', err);
      alert('Error saving: ' + err);
    }
  }

  function getButtonAtPosition(row, col) {
    if (!selectedConfig) return null;
    const position = `btn-${row}-${col}`;
    const buttonId = selectedConfig.buttons[position];
    if (!buttonId) return null;
    return buttons.find(b => b.id === buttonId);
  }

  async function duplicateButton(button) {
    try {
      const copy = {
        name: button.name + ' (Copy)',
        description: button.description,
        icon: button.icon,
        color: button.color,
        action: { ...button.action, params: { ...button.action.params } }
      };
      
      await window.go.main.App.CreateButton(copy);
      await loadData();
      
      setTimeout(() => {
        if (window.lucide) lucide.createIcons();
      }, 100);
    } catch (err) {
      console.error('Failed to duplicate button:', err);
      alert('Error: ' + err);
    }
  }

  async function executeButtonAction(button) {
    if (!button || editMode) return;  // Don't execute in edit mode
    
    try {
      console.log('Executing button action:', button.name, button.action);
      await window.go.main.App.ExecuteAction(button.action);
      console.log('Action executed successfully');
    } catch (err) {
      console.error('Failed to execute action:', err);
      // Show error to user
      alert('Error executing action: ' + err);
    }
  }
</script>

<div class="config-editor">
  <!-- Left Sidebar - Configuration List -->
  <div class="sidebar">
    <div class="sidebar-header">
      <h3>Configurations</h3>
      <button class="btn-icon" on:click={createConfiguration} title="New Configuration">
        <i data-lucide="plus"></i>
      </button>
    </div>

    {#if loading}
      <div class="loading">Loading...</div>
    {:else if configurations.length === 0}
      <div class="empty-sidebar">
        <p>No configurations</p>
        <button class="btn-sm" on:click={createConfiguration}>Create One</button>
      </div>
    {:else}
      <div class="config-list">
        {#each configurations as config}
          <div 
            class="config-list-item" 
            class:active={selectedConfig?.id === config.id}
            on:click={() => selectConfig(config)}
          >
            <div class="config-list-item-header">
              <span>{config.name}</span>
              {#if config.is_default}
                <span class="badge-small">Default</span>
              {/if}
            </div>
            <div class="config-list-item-info">
              <span>{config.grid.rows}×{config.grid.cols}</span>
              <span>•</span>
              <span>{Object.keys(config.buttons || {}).length} buttons</span>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Main Content - Configuration View -->
  <div class="main">
    {#if !selectedConfig}
      <div class="empty-state">
        <i data-lucide="layout"></i>
        <h2>No Configuration Selected</h2>
        <p>Select a configuration from the list or create a new one</p>
        <button class="btn-primary" on:click={createConfiguration}>
          <i data-lucide="plus"></i>
          Create Configuration
        </button>
      </div>
    {:else}
      <!-- Configuration Header -->
      <div class="config-header">
        <div class="config-info">
          <h2>{selectedConfig.name}</h2>
          <p>{selectedConfig.description || 'No description'}</p>
          <div class="config-meta">
            <span>{selectedConfig.grid.rows}×{selectedConfig.grid.cols} grid</span>
            <span>•</span>
            <span>{Object.keys(selectedConfig.buttons || {}).length} buttons assigned</span>
            {#if selectedConfig.is_default}
              <span class="badge-default">Default</span>
            {/if}
          </div>
        </div>
        <div class="config-actions">
          <button class="btn-secondary" on:click={toggleEditMode}>
            <i data-lucide={editMode ? 'eye' : 'edit'}></i>
            {editMode ? 'Preview' : 'Edit'}
          </button>
          <button class="btn-secondary" on:click={editConfiguration}>
            <i data-lucide="settings"></i>
            Settings
          </button>
          <button class="btn-secondary" on:click={duplicateConfiguration}>
            <i data-lucide="copy"></i>
            Duplicate
          </button>
          {#if !selectedConfig.is_default}
            <button class="btn-secondary" on:click={setDefault}>
              <i data-lucide="star"></i>
              Set Default
            </button>
          {/if}
          {#if !(selectedConfig.is_default && selectedConfig.name === 'Default')}
            <button class="btn-danger" on:click={deleteConfiguration}>
              <i data-lucide="trash-2"></i>
              Delete
            </button>
          {/if}
        </div>
      </div>

      <div class="config-content">
        <!-- Button Grid -->
        <div class="grid-container">
          <div 
            class="button-grid" 
            style="grid-template-columns: repeat({selectedConfig.grid.cols}, 1fr); grid-template-rows: repeat({selectedConfig.grid.rows}, 1fr);"
          >
            {#each Array(selectedConfig.grid.rows) as _, row}
              {#each Array(selectedConfig.grid.cols) as _, col}
                {@const button = getButtonAtPosition(row, col)}
                <div 
                  class="grid-cell"
                  class:drop-target={editMode}
                  on:drop={(e) => handleDrop(e, row, col)}
                  on:dragover={handleDragOver}
                >
                  {#if button}
                    <div 
                      class="button-display" 
                      class:clickable={!editMode}
                      style="background: {button.color}"
                      on:click={() => executeButtonAction(button)}
                    >
                      <i data-lucide={button.icon}></i>
                      <span>{button.name}</span>
                      {#if editMode}
                        <button class="remove-btn" on:click|stopPropagation={() => removeButton(`btn-${row}-${col}`)}>
                          ×
                        </button>
                      {/if}
                    </div>
                  {:else if editMode}
                    <div class="empty-cell">Drop here</div>
                  {:else}
                    <div class="empty-cell-preview">Empty</div>
                  {/if}
                </div>
              {/each}
            {/each}
          </div>
        </div>

        <!-- Button Library (Edit Mode) -->
        {#if editMode}
          <div class="button-library-panel">
            <div class="library-header">
              <h3>Button Library</h3>
              <button class="btn-icon" on:click={createButton} title="New Button">
                <i data-lucide="plus"></i>
              </button>
            </div>
            <p class="hint">Drag buttons onto the grid to assign them</p>
            
            {#if buttons.length === 0}
              <div class="empty">
                <p>No buttons available</p>
                <button class="btn-sm" on:click={createButton}>
                  Create Button
                </button>
              </div>
            {:else}
              <div class="button-library-list">
                {#each buttons as button}
                  <div 
                    class="library-button"
                    draggable="true"
                    on:dragstart={(e) => handleDragStart(e, button)}
                    on:dragend={handleDragEnd}
                  >
                    <div class="library-button-preview" style="background: {button.color}" draggable="false">
                      <i data-lucide={button.icon}></i>
                    </div>
                    <div class="library-button-info" draggable="false">
                      <span class="library-button-name">{button.name}</span>
                      <span class="library-button-action">{button.action.type}</span>
                    </div>
                    <div class="library-button-actions" draggable="false">
                      <button 
                        class="btn-icon-small" 
                        on:click|stopPropagation={() => editButton(button)}
                        on:mousedown|stopPropagation
                        title="Edit"
                      >
                        <i data-lucide="edit"></i>
                      </button>
                      <button 
                        class="btn-icon-small" 
                        on:click|stopPropagation={() => duplicateButton(button)}
                        on:mousedown|stopPropagation
                        title="Duplicate"
                      >
                        <i data-lucide="copy"></i>
                      </button>
                      <button 
                        class="btn-icon-small btn-danger-small" 
                        on:click|stopPropagation={() => deleteButton(button)}
                        on:mousedown|stopPropagation
                        title="Delete"
                      >
                        <i data-lucide="trash-2"></i>
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<ConfigModal 
  isOpen={showConfigModal} 
  config={editingConfig} 
  onSave={handleSave} 
  onClose={closeConfigModal}
/>

<ButtonModal 
  isOpen={showButtonModal} 
  button={editingButton} 
  onSave={handleButtonSave} 
  onClose={closeButtonModal}
/>

<style>
  .config-editor {
    display: flex;
    height: calc(100vh - 60px);
    background: #0f1419;
  }

  /* Sidebar */
  .sidebar {
    width: 300px;
    background: #16213e;
    border-right: 1px solid #0f3460;
    display: flex;
    flex-direction: column;
  }

  .sidebar-header {
    padding: 20px;
    border-bottom: 1px solid #0f3460;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .sidebar-header h3 {
    font-size: 18px;
    margin: 0;
  }

  .config-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }

  .config-list-item {
    padding: 12px 16px;
    margin-bottom: 4px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .config-list-item:hover {
    background: #0f3460;
  }

  .config-list-item.active {
    background: #3b82f6;
  }

  .config-list-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }

  .config-list-item-header span:first-child {
    font-weight: 500;
  }

  .config-list-item-info {
    font-size: 12px;
    color: #94a3b8;
    display: flex;
    gap: 6px;
  }

  .badge-small {
    padding: 2px 8px;
    background: #10b981;
    color: white;
    border-radius: 10px;
    font-size: 10px;
    font-weight: 600;
  }

  .empty-sidebar {
    padding: 20px;
    text-align: center;
    color: #94a3b8;
  }

  /* Main Content */
  .main {
    flex: 1;
    overflow-y: auto;
    padding: 32px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 16px;
    color: #94a3b8;
  }

  .empty-state i {
    width: 64px;
    height: 64px;
    color: #94a3b8;
  }

  .config-header {
    margin-bottom: 32px;
  }

  .config-info h2 {
    font-size: 28px;
    margin-bottom: 8px;
  }

  .config-info p {
    color: #94a3b8;
    margin-bottom: 12px;
  }

  .config-meta {
    display: flex;
    gap: 8px;
    align-items: center;
    font-size: 14px;
    color: #94a3b8;
  }

  .badge-default {
    padding: 4px 12px;
    background: #10b981;
    color: white;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
  }

  .config-actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
    flex-wrap: wrap;
  }

  .config-content {
    display: grid;
    grid-template-columns: 1fr 350px;
    gap: 32px;
  }

  .config-content:not(:has(.button-library-panel)) {
    grid-template-columns: 1fr;
  }

  /* Button Grid */
  .grid-container {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    padding: 24px;
  }

  .button-grid {
    display: grid;
    gap: 12px;
    max-width: 800px;
    margin: 0 auto;
  }

  .grid-cell {
    position: relative;
    aspect-ratio: 1;
    border-radius: 8px;
  }

  .grid-cell.drop-target {
    border: 2px dashed #0f3460;
  }

  .button-display {
    width: 100%;
    height: 100%;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    position: relative;
    transition: all 0.2s;
  }

  .button-display.clickable {
    cursor: pointer;
  }

  .button-display.clickable:hover {
    transform: scale(1.05);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .button-display:not(.clickable) {
    cursor: default;
  }

  .button-display i {
    width: 32px;
    height: 32px;
    color: white;
  }

  .button-display span {
    color: white;
    font-size: 14px;
    font-weight: 500;
    text-align: center;
  }

  .remove-btn {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.6);
    color: white;
    border: none;
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
    display: none;
  }

  .button-display:hover .remove-btn {
    display: block;
  }

  .remove-btn:hover {
    background: #ef4444;
  }

  .empty-cell,
  .empty-cell-preview {
    width: 100%;
    height: 100%;
    background: #0f1419;
    border: 2px dashed #0f3460;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #94a3b8;
    font-size: 14px;
  }

  .empty-cell-preview {
    border-style: solid;
    opacity: 0.3;
  }

  /* Button Library Panel */
  .button-library-panel {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    padding: 20px;
    max-height: calc(100vh - 300px);
    overflow-y: auto;
  }

  .library-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .button-library-panel h3 {
    font-size: 18px;
    margin: 0;
  }

  .hint {
    font-size: 12px;
    color: #94a3b8;
    margin-bottom: 16px;
  }

  .button-library-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .library-button {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px;
    background: #0f1419;
    border: 1px solid #0f3460;
    border-radius: 6px;
    cursor: grab;
  }

  .library-button:active {
    cursor: grabbing;
  }

  .library-button-preview {
    width: 48px;
    height: 48px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .library-button-preview i {
    width: 24px;
    height: 24px;
    color: white;
  }

  .library-button-info {
    flex: 1;
    min-width: 0;
  }

  .library-button-name {
    display: block;
    font-size: 14px;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .library-button-action {
    display: block;
    font-size: 12px;
    color: #94a3b8;
  }

  .library-button-actions {
    display: flex;
    gap: 4px;
  }

  .btn-icon-small {
    padding: 6px;
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .btn-icon-small:hover {
    color: #eaeaea;
    background: #0f3460;
  }

  .btn-icon-small.btn-danger-small {
    color: #ef4444;
  }

  .btn-icon-small.btn-danger-small:hover {
    background: #7f1d1d;
  }

  .btn-icon-small i {
    width: 14px;
    height: 14px;
  }

  /* Buttons */
  .btn-primary,
  .btn-secondary,
  .btn-danger,
  .btn-sm,
  .btn-icon {
    padding: 10px 16px;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    border: none;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-secondary {
    background: transparent;
    border: 1px solid #0f3460;
    color: #eaeaea;
  }

  .btn-danger {
    background: transparent;
    border: 1px solid #ef4444;
    color: #ef4444;
  }

  .btn-icon {
    padding: 8px;
    background: transparent;
    border: 1px solid #0f3460;
    color: #eaeaea;
  }

  .loading,
  .empty {
    padding: 20px;
    text-align: center;
    color: #94a3b8;
  }
</style>
