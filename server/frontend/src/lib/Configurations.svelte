<script>
  import { onMount } from 'svelte';
  import ConfigModal from './ConfigModal.svelte';

  let configurations = [];
  let loading = true;
  let showModal = false;
  let editingConfig = null;

  onMount(async () => {
    await loadConfigurations();
    // Reinitialize icons after configs load
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  });

  async function loadConfigurations() {
    try {
      configurations = await window.go.main.App.GetConfigurations();
      console.log('Loaded configurations:', configurations);
    } catch (err) {
      console.error('Failed to load configurations:', err);
    } finally {
      loading = false;
    }
  }

  function createConfiguration() {
    editingConfig = null;
    showModal = true;
  }

  function editConfiguration(config) {
    editingConfig = config;
    showModal = true;
  }

  async function handleSave(configData) {
    try {
      if (editingConfig) {
        // Update existing config
        console.log('Updating configuration:', configData);
        await window.go.main.App.UpdateConfiguration(configData);
      } else {
        // Create new config
        console.log('Creating configuration:', configData);
        await window.go.main.App.CreateConfiguration(configData);
      }
      
      showModal = false;
      editingConfig = null;
      await loadConfigurations();
      
      // Reinitialize icons
      setTimeout(() => {
        if (window.lucide) lucide.createIcons();
      }, 100);
    } catch (err) {
      console.error('Failed to save configuration:', err);
      alert('Error: ' + err);
    }
  }

  async function deleteConfiguration(config) {
    if (confirm(`Delete configuration "${config.name}"?`)) {
      try {
        console.log('Deleting configuration:', config.id);
        await window.go.main.App.DeleteConfiguration(config.id);
        await loadConfigurations();
      } catch (err) {
        console.error('Failed to delete configuration:', err);
        alert('Error: ' + err);
      }
    }
  }

  async function setDefault(config) {
    try {
      console.log('Setting default configuration:', config.id);
      await window.go.main.App.SetDefaultConfiguration(config.id);
      await loadConfigurations();
    } catch (err) {
      console.error('Failed to set default:', err);
      alert('Error: ' + err);
    }
  }

  function closeModal() {
    showModal = false;
    editingConfig = null;
  }
</script>

<div class="configurations">
  <header>
    <div>
      <h2>Configurations</h2>
      <p>Manage button layouts for different roles</p>
    </div>
    <button class="btn-primary" on:click={createConfiguration}>
      <i data-lucide="plus"></i>
      New Configuration
    </button>
  </header>

  {#if loading}
    <div class="loading">Loading configurations...</div>
  {:else if configurations.length === 0}
    <div class="empty">
      <i data-lucide="grid-3x3"></i>
      <h3>No configurations yet</h3>
      <p>Create your first configuration to get started</p>
      <button class="btn-primary" on:click={createConfiguration}>
        <i data-lucide="plus"></i>
        Create Configuration
      </button>
    </div>
  {:else}
    <div class="config-grid">
      {#each configurations as config}
        <div class="config-card">
          <div class="config-header">
            <h3>{config.name}</h3>
            {#if config.is_default}
              <span class="badge-default">Default</span>
            {/if}
          </div>
          <p class="config-description">{config.description || 'No description'}</p>
          <div class="config-stats">
            <span>{config.grid.rows}x{config.grid.cols} grid</span>
            <span>•</span>
            <span>{Object.keys(config.buttons || {}).length} buttons</span>
          </div>
          <div class="config-actions">
            <button class="btn-sm" on:click={() => editConfiguration(config)} title="Edit">Edit</button>
            <button class="btn-sm" on:click={() => deleteConfiguration(config)} title="Delete">Delete</button>
            {#if !config.is_default}
              <button class="btn-sm" on:click={() => setDefault(config)} title="Set as default">Set Default</button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<ConfigModal 
  isOpen={showModal} 
  config={editingConfig} 
  onSave={handleSave} 
  onClose={closeModal}
/>

<style>
  .configurations {
    padding: 32px;
    max-width: 1200px;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 32px;
  }

  header h2 {
    font-size: 28px;
    margin-bottom: 8px;
  }

  header p {
    color: #94a3b8;
    font-size: 14px;
  }

  .btn-primary {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 20px;
    background: #3b82f6;
    border: none;
    border-radius: 8px;
    color: white;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }

  .btn-primary:hover {
    background: #2563eb;
  }

  .btn-primary i {
    width: 18px;
    height: 18px;
  }

  .loading, .empty {
    text-align: center;
    padding: 60px 20px;
    color: #94a3b8;
  }

  .empty i {
    width: 64px;
    height: 64px;
    margin-bottom: 20px;
    opacity: 0.5;
  }

  .empty h3 {
    font-size: 20px;
    color: #eaeaea;
    margin-bottom: 8px;
  }

  .empty p {
    margin-bottom: 24px;
  }

  .config-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 20px;
  }

  .config-card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    padding: 24px;
    transition: transform 0.2s, box-shadow 0.2s;
  }

  .config-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  }

  .config-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .config-header h3 {
    font-size: 18px;
  }

  .badge-default {
    padding: 4px 12px;
    background: #10b981;
    color: white;
    font-size: 11px;
    font-weight: 600;
    border-radius: 12px;
  }

  .config-description {
    color: #94a3b8;
    font-size: 14px;
    margin-bottom: 16px;
  }

  .config-stats {
    display: flex;
    gap: 8px;
    font-size: 13px;
    color: #94a3b8;
    margin-bottom: 20px;
  }

  .config-actions {
    display: flex;
    gap: 8px;
  }

  .btn-sm {
    flex: 1;
    padding: 8px 12px;
    background: transparent;
    border: 1px solid #0f3460;
    border-radius: 6px;
    color: #eaeaea;
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-sm:hover {
    background: #0f3460;
    border-color: #3b82f6;
  }
</style>
