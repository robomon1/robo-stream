<script>
  export let isOpen = false;
  export let config = null; // null for create, object for edit
  export let onSave = () => {};
  export let onClose = () => {};

  let formData = {
    name: '',
    description: '',
    rows: 3,
    cols: 4
  };

  $: if (isOpen && config) {
    // Edit mode - load config data
    formData = {
      name: config.name || '',
      description: config.description || '',
      rows: config.grid?.rows || 3,
      cols: config.grid?.cols || 4
    };
  } else if (isOpen && !config) {
    // Create mode - reset form
    formData = {
      name: '',
      description: '',
      rows: 3,
      cols: 4
    };
  }

  function handleSave() {
    const configData = {
      name: formData.name,
      description: formData.description,
      grid: {
        rows: formData.rows,
        cols: formData.cols
      },
      buttons: config?.buttons || {},
      is_default: config?.is_default || false
    };

    if (config) {
      configData.id = config.id;
    }

    onSave(configData);
  }

  let mouseDownOnOverlay = false;

  function handleOverlayMouseDown(e) {
    if (e.target.classList.contains('modal-overlay')) {
      mouseDownOnOverlay = true;
    }
  }

  function handleOverlayClick(e) {
    if (mouseDownOnOverlay && e.target.classList.contains('modal-overlay')) {
      onClose();
    }
    mouseDownOnOverlay = false;
  }
</script>

{#if isOpen}
  <div 
    class="modal-overlay" 
    on:mousedown={handleOverlayMouseDown}
    on:click={handleOverlayClick}
  >
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h2>{config ? 'Edit Configuration' : 'Create Configuration'}</h2>
        <button class="close-btn" on:click={onClose}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label>Name *</label>
          <input type="text" bind:value={formData.name} placeholder="Streamer Layout" />
        </div>

        <div class="form-group">
          <label>Description</label>
          <input type="text" bind:value={formData.description} placeholder="Full control for the main streamer" />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Rows</label>
            <input type="number" min="1" max="10" bind:value={formData.rows} />
          </div>

          <div class="form-group">
            <label>Columns</label>
            <input type="number" min="1" max="10" bind:value={formData.cols} />
          </div>
        </div>

        <div class="grid-preview">
          <p>Grid preview: {formData.rows}×{formData.cols} ({formData.rows * formData.cols} buttons)</p>
          <div class="preview-grid" style="grid-template-columns: repeat({formData.cols}, 1fr); grid-template-rows: repeat({formData.rows}, 1fr);">
            {#each Array(formData.rows * formData.cols) as _, i}
              <div class="preview-cell"></div>
            {/each}
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" on:click={onClose}>Cancel</button>
        <button class="btn-primary" on:click={handleSave} disabled={!formData.name}>
          {config ? 'Save Changes' : 'Create Configuration'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #0f3460;
  }

  .modal-header h2 {
    font-size: 20px;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 32px;
    cursor: pointer;
    line-height: 1;
    padding: 0;
    width: 32px;
    height: 32px;
  }

  .close-btn:hover {
    color: #eaeaea;
  }

  .modal-body {
    padding: 24px;
  }

  .form-group {
    margin-bottom: 20px;
  }

  .form-group label {
    display: block;
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 8px;
  }

  .form-group input {
    width: 100%;
    padding: 10px 12px;
    background: #0f1419;
    border: 1px solid #0f3460;
    border-radius: 6px;
    color: #eaeaea;
    font-size: 14px;
  }

  .form-group input::placeholder {
    color: #64748b;
    opacity: 1;
  }

  .form-group input:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .grid-preview {
    margin-top: 24px;
    padding: 20px;
    background: #0f1419;
    border-radius: 8px;
  }

  .grid-preview p {
    margin-bottom: 16px;
    color: #94a3b8;
    font-size: 14px;
  }

  .preview-grid {
    display: grid;
    gap: 8px;
    max-width: 400px;
    margin: 0 auto;
  }

  .preview-cell {
    aspect-ratio: 1;
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 6px;
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid #0f3460;
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn-primary,
  .btn-secondary {
    padding: 10px 20px;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    border: none;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: transparent;
    border: 1px solid #0f3460;
    color: #eaeaea;
  }

  .btn-secondary:hover {
    background: #0f3460;
  }
</style>
