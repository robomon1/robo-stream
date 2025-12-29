<script>
  import { onMount } from 'svelte';
  import ButtonModal from './ButtonModal.svelte';

  let buttons = [];
  let loading = true;
  let showModal = false;
  let editingButton = null;

  onMount(async () => {
    await loadButtons();
    // Reinitialize icons after buttons load
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  });

  async function loadButtons() {
    try {
      buttons = await window.go.main.App.GetButtons();
      console.log('Loaded buttons:', buttons);
    } catch (err) {
      console.error('Failed to load buttons:', err);
    } finally {
      loading = false;
    }
  }

  function createButton() {
    editingButton = null;
    showModal = true;
  }

  function editButton(button) {
    editingButton = button;
    showModal = true;
  }

  async function handleSave(buttonData) {
    try {
      if (editingButton) {
        // Update existing button
        console.log('Updating button:', buttonData);
        await window.go.main.App.UpdateButton(buttonData);
      } else {
        // Create new button
        console.log('Creating button:', buttonData);
        await window.go.main.App.CreateButton(buttonData);
      }
      
      showModal = false;
      editingButton = null;
      await loadButtons();
      
      // Reinitialize icons
      setTimeout(() => {
        if (window.lucide) lucide.createIcons();
      }, 100);
    } catch (err) {
      console.error('Failed to save button:', err);
      alert('Error: ' + err);
    }
  }

  async function deleteButton(button) {
    // Note: confirm() doesn't work in Wails
    console.log('Deleting button:', button.id, button.name);
    try {
      await window.go.main.App.DeleteButton(button.id);
      await loadButtons();
    } catch (err) {
      console.error('Failed to delete button:', err);
      alert('Error: ' + err);
    }
  }

  function closeModal() {
    showModal = false;
    editingButton = null;
  }
</script>

<div class="button-library">
  <header>
    <div>
      <h2>Button Library</h2>
      <p>Reusable buttons for your configurations</p>
    </div>
    <button class="btn-primary" on:click={createButton}>
      <i data-lucide="plus"></i>
      New Button
    </button>
  </header>

  {#if loading}
    <div class="loading">Loading buttons...</div>
  {:else if buttons.length === 0}
    <div class="empty">
      <i data-lucide="square"></i>
      <h3>No buttons yet</h3>
      <p>Create your first button to get started</p>
      <button class="btn-primary" on:click={createButton}>
        <i data-lucide="plus"></i>
        Create Button
      </button>
    </div>
  {:else}
    <div class="button-grid">
      {#each buttons as button}
        <div class="button-item">
          <div class="button-preview" style="background: {button.color}">
            <i data-lucide={button.icon}></i>
            <span>{button.name}</span>
          </div>
          <div class="button-info">
            <h4>{button.name}</h4>
            <p>{button.description || 'No description'}</p>
            <div class="button-actions">
              <button class="btn-icon" on:click={() => editButton(button)} title="Edit">
                <i data-lucide="edit"></i>
              </button>
              <button class="btn-icon" on:click={() => deleteButton(button)} title="Delete">
                <i data-lucide="trash-2"></i>
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<ButtonModal 
  isOpen={showModal} 
  button={editingButton} 
  onSave={handleSave} 
  onClose={closeModal}
/>

<style>
  .button-library {
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
  }

  .btn-primary:hover {
    background: #2563eb;
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

  .button-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }

  .button-item {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    overflow: hidden;
  }

  .button-preview {
    aspect-ratio: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 20px;
  }

  .button-preview i {
    width: 32px;
    height: 32px;
    color: white;
  }

  .button-preview span {
    color: white;
    font-size: 14px;
    font-weight: 500;
  }

  .button-info {
    padding: 16px;
    border-top: 1px solid #0f3460;
  }

  .button-info h4 {
    font-size: 14px;
    margin-bottom: 4px;
  }

  .button-info p {
    font-size: 12px;
    color: #94a3b8;
    margin-bottom: 12px;
  }

  .button-actions {
    display: flex;
    gap: 8px;
  }

  .btn-icon {
    padding: 6px;
    background: transparent;
    border: 1px solid #0f3460;
    border-radius: 6px;
    color: #eaeaea;
    cursor: pointer;
  }

  .btn-icon:hover {
    background: #0f3460;
  }

  .btn-icon i {
    width: 16px;
    height: 16px;
  }
</style>
