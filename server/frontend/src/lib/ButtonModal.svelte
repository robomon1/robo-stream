<script>
  import { onMount } from 'svelte';
  
  export let isOpen = false;
  export let button = null; // null for create, object for edit
  export let onSave = () => {};
  export let onClose = () => {};

  let formData = {
    name: '',
    description: '',
    icon: 'square',
    color: '#3b82f6',
    actionType: 'switch_scene',
    actionParams: {}
  };

  let testing = false;
  let testResult = '';
  let scenes = [];
  let inputs = [];
  let loadingOBSData = false;

  $: if (isOpen) {
    loadOBSData();
    // Reinitialize icons when modal opens
    setTimeout(() => {
      if (window.lucide) lucide.createIcons();
    }, 100);
  }

  async function loadOBSData() {
    if (loadingOBSData) return;
    loadingOBSData = true;
    
    try {
      if (window.go && window.go.main && window.go.main.App) {
        scenes = await window.go.main.App.GetScenes() || [];
        inputs = await window.go.main.App.GetInputs() || [];
        console.log('Loaded scenes:', scenes.length, scenes);
        console.log('Loaded inputs:', inputs.length, inputs);
      }
    } catch (err) {
      console.error('Failed to load OBS data:', err);
      scenes = [];
      inputs = [];
    } finally {
      loadingOBSData = false;
    }
  }

  $: if (isOpen && button) {
    // Edit mode - load button data
    formData = {
      name: button.name || '',
      description: button.description || '',
      icon: button.icon || 'square',
      color: button.color || '#3b82f6',
      actionType: button.action?.type || 'switch_scene',
      actionParams: { ...button.action?.params } || {}
    };
    testResult = '';
  } else if (isOpen && !button) {
    // Create mode - completely empty to show placeholders
    formData = {
      name: '',
      description: '',
      icon: 'square',
      color: '#3b82f6',
      actionType: 'switch_scene',
      actionParams: {}
    };
    testResult = '';
  }

  // Watch action type and update params accordingly
  $: {
    // When action type changes, ensure params match
    const requiredParams = getRequiredParams();
    
    console.log('Action type:', formData.actionType, 'Required params:', requiredParams);
    
    // Create new params object with only the required params
    const newParams = {};
    
    for (const param of requiredParams) {
      if (param === 'scene_name') {
        // Use existing value or default to first scene
        newParams.scene_name = formData.actionParams.scene_name || (scenes.length > 0 ? scenes[0] : '');
      } else if (param === 'input_name') {
        // Use existing value or default to first input
        newParams.input_name = formData.actionParams.input_name || (inputs.length > 0 ? inputs[0] : '');
      } else {
        // Keep other params
        newParams[param] = formData.actionParams[param] || '';
      }
    }
    
    // Update params if action type is defined (avoid initialization issues)
    if (formData.actionType) {
      formData.actionParams = newParams;
      console.log('Updated params:', formData.actionParams);
    }
  }

  const icons = [
    { value: 'video', label: '📹 Video' },
    { value: 'play', label: '▶️ Play' },
    { value: 'pause', label: '⏸️ Pause' },
    { value: 'stop-circle', label: '⏹️ Stop' },
    { value: 'circle', label: '⏺️ Record' },
    { value: 'mic', label: '🎤 Mic' },
    { value: 'mic-off', label: '🎤 Mic Off' },
    { value: 'volume-2', label: '🔊 Volume' },
    { value: 'volume-x', label: '🔇 Mute' },
    { value: 'camera', label: '📷 Camera' },
    { value: 'layout', label: '🎬 Scene' },
    { value: 'monitor', label: '🖥️ Monitor' },
    { value: 'square', label: '⬜ Square' },
    { value: 'circle', label: '⭕ Circle' },
    { value: 'star', label: '⭐ Star' },
  ];

  const actionTypes = [
    { value: 'switch_scene', label: 'Switch Scene', params: ['scene_name'] },
    { value: 'start_stream', label: 'Start Stream', params: [] },
    { value: 'stop_stream', label: 'Stop Stream', params: [] },
    { value: 'toggle_stream', label: 'Toggle Stream', params: [] },
    { value: 'start_record', label: 'Start Recording', params: [] },
    { value: 'stop_record', label: 'Stop Recording', params: [] },
    { value: 'toggle_record', label: 'Toggle Recording', params: [] },
    { value: 'toggle_input_mute', label: 'Toggle Input Mute', params: ['input_name'] },
    { value: 'mute_input', label: 'Mute Input', params: ['input_name'] },
    { value: 'unmute_input', label: 'Unmute Input', params: ['input_name'] },
  ];

  function handleSave() {
    // Build button object
    const buttonData = {
      name: formData.name,
      description: formData.description,
      icon: formData.icon,
      color: formData.color,
      action: {
        type: formData.actionType,
        params: formData.actionParams
      }
    };

    if (button) {
      // Edit mode - include ID
      buttonData.id = button.id;
    }

    onSave(buttonData);
  }

  async function testAction() {
    testing = true;
    testResult = '';
    
    try {
      console.log('Testing action:', formData.actionType, formData.actionParams);
      
      // Execute the action via the OBS manager
      await window.go.main.App.ExecuteAction({
        type: formData.actionType,
        params: formData.actionParams
      });
      
      testResult = '✅ Action executed successfully!';
      console.log('Test successful');
    } catch (err) {
      testResult = '❌ Error: ' + err;
      console.error('Test failed:', err);
    } finally {
      testing = false;
      
      // Clear result after 3 seconds
      setTimeout(() => {
        testResult = '';
      }, 3000);
    }
  }

  function getRequiredParams() {
    const actionType = actionTypes.find(a => a.value === formData.actionType);
    return actionType ? actionType.params : [];
  }

  let mouseDownOnOverlay = false;

  function handleOverlayMouseDown(e) {
    // Only mark as down if clicking directly on overlay, not on modal content
    if (e.target.classList.contains('modal-overlay')) {
      mouseDownOnOverlay = true;
    }
  }

  function handleOverlayClick(e) {
    // Only close if both mousedown and click were on overlay
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
        <h2>{button ? 'Edit Button' : 'Create Button'}</h2>
        <button class="close-btn" on:click={onClose}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label>Name *</label>
          <input type="text" bind:value={formData.name} placeholder="Go Live" />
        </div>

        <div class="form-group">
          <label>Description</label>
          <input type="text" bind:value={formData.description} placeholder="Start streaming to Twitch" />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Icon</label>
            <select bind:value={formData.icon}>
              {#each icons as icon}
                <option value={icon.value}>{icon.label}</option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label>Color</label>
            <input type="color" bind:value={formData.color} />
          </div>
        </div>

        <div class="form-group">
          <label>Action Type</label>
          <select bind:value={formData.actionType}>
            {#each actionTypes as action}
              <option value={action.value}>{action.label}</option>
            {/each}
          </select>
        </div>

        {#key formData.actionType}
          {#each getRequiredParams() as param}
            <div class="form-group">
              {#if param === 'scene_name'}
                <label>Scene Name</label>
                {#if scenes.length > 0}
                  <select bind:value={formData.actionParams[param]}>
                    {#each scenes as scene}
                      <option value={scene}>{scene}</option>
                    {/each}
                  </select>
                {:else}
                  <input 
                    type="text" 
                    bind:value={formData.actionParams[param]} 
                    placeholder="Main"
                  />
                  <p class="help-text">OBS not connected - enter scene name manually</p>
                {/if}
              {:else if param === 'input_name'}
                <label>Input Name</label>
                {#if inputs.length > 0}
                  <select bind:value={formData.actionParams[param]}>
                    {#each inputs as input}
                      <option value={input}>{input}</option>
                    {/each}
                  </select>
                {:else}
                  <input 
                    type="text" 
                    bind:value={formData.actionParams[param]} 
                    placeholder="Mic/Aux"
                  />
                  <p class="help-text">OBS not connected - enter input name manually</p>
                {/if}
              {:else}
                <label>{param.replace('_', ' ')}</label>
                <input 
                  type="text" 
                  bind:value={formData.actionParams[param]} 
                  placeholder={param}
                />
              {/if}
            </div>
          {/each}
        {/key}

        <div class="button-preview" style="background: {formData.color}">
          <i data-lucide={formData.icon}></i>
          <span>{formData.name || 'Preview'}</span>
        </div>

        {#if testResult}
          <div class="test-result" class:success={testResult.startsWith('✅')} class:error={testResult.startsWith('❌')}>
            {testResult}
          </div>
        {/if}
      </div>

      <div class="modal-footer">
        <button class="btn-test" on:click={testAction} disabled={testing || !formData.name}>
          {testing ? 'Testing...' : '🧪 Test Action'}
        </button>
        <div class="spacer"></div>
        <button class="btn-secondary" on:click={onClose}>Cancel</button>
        <button class="btn-primary" on:click={handleSave} disabled={!formData.name}>
          {button ? 'Save Changes' : 'Create Button'}
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

  .form-group input,
  .form-group select {
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

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .help-text {
    margin-top: 4px;
    font-size: 12px;
    color: #94a3b8;
    font-style: italic;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .button-preview {
    margin-top: 24px;
    padding: 24px;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
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

  .test-result {
    margin-top: 16px;
    padding: 12px 16px;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
  }

  .test-result.success {
    background: #10b98120;
    border: 1px solid #10b981;
    color: #10b981;
  }

  .test-result.error {
    background: #ef444420;
    border: 1px solid #ef4444;
    color: #ef4444;
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid #0f3460;
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .spacer {
    flex: 1;
  }

  .btn-primary,
  .btn-secondary,
  .btn-test {
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

  .btn-test {
    background: #10b981;
    color: white;
  }

  .btn-test:hover:not(:disabled) {
    background: #059669;
  }

  .btn-test:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
