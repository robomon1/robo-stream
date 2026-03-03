<script>
  export let isOpen = false;
  export let button = null; // null for create, object for edit
  export let onSave = () => {};
  export let onClose = () => {};

  let formData = {
    name: '',
    description: '',
    icon: 'square',
    color: '#3b82f6',
    actionController: '',
    actionType: '',
    actionParams: {}
  };

  let testing = false;
  let testResult = '';
  let scenes = [];
  let inputs = [];
  let loadingOBSData = false;
  let initialized = false;

  // Controller groups loaded from backend: [{controller_id, controller_name, connected, actions:[]}]
  let actionGroups = [];
  // Flat lookup: [{value, label, controller, params:[string]}]
  let flatActionTypes = [];
  // Which controller tab is active in the picker
  let activeTab = '';

  // Fallback OBS actions shown while backend data loads
  const obsActionsFallback = [
    { value: 'switch_scene',            label: 'Switch Scene',            params: ['scene_name'] },
    { value: 'start_stream',            label: 'Start Stream',            params: [] },
    { value: 'stop_stream',             label: 'Stop Stream',             params: [] },
    { value: 'toggle_stream',           label: 'Toggle Stream',           params: [] },
    { value: 'start_record',            label: 'Start Recording',         params: [] },
    { value: 'stop_record',             label: 'Stop Recording',          params: [] },
    { value: 'toggle_record',           label: 'Toggle Recording',        params: [] },
    { value: 'pause_record',            label: 'Pause Recording',         params: [] },
    { value: 'resume_record',           label: 'Resume Recording',        params: [] },
    { value: 'toggle_source_visibility',label: 'Toggle Source Visibility',params: ['scene_name','source_name'] },
    { value: 'toggle_input_mute',       label: 'Toggle Input Mute',       params: ['input_name'] },
    { value: 'mute_input',              label: 'Mute Input',              params: ['input_name'] },
    { value: 'unmute_input',            label: 'Unmute Input',            params: ['input_name'] },
    { value: 'set_input_volume',        label: 'Set Input Volume',        params: ['input_name','volume'] },
    { value: 'start_virtual_cam',       label: 'Start Virtual Camera',    params: [] },
    { value: 'stop_virtual_cam',        label: 'Stop Virtual Camera',     params: [] },
    { value: 'toggle_virtual_cam',      label: 'Toggle Virtual Camera',   params: [] },
    { value: 'start_replay_buffer',     label: 'Start Replay Buffer',     params: [] },
    { value: 'stop_replay_buffer',      label: 'Stop Replay Buffer',      params: [] },
    { value: 'save_replay_buffer',      label: 'Save Replay Buffer',      params: [] },
    { value: 'toggle_studio_mode',      label: 'Toggle Studio Mode',      params: [] },
    { value: 'set_preview_scene',       label: 'Set Preview Scene',       params: ['scene_name'] },
  ];

  // ── Lifecycle ────────────────────────────────────────────────────────────────

  $: if (isOpen) {
    loadOBSData();
    loadActionTypes();
    setTimeout(() => { if (window.lucide) lucide.createIcons(); }, 100);
  }

  $: if (!isOpen) { initialized = false; }

  // Edit mode — populate form from existing button
  $: if (isOpen && button && !initialized) {
    formData = {
      name:             button.name        || '',
      description:      button.description || '',
      icon:             button.icon        || 'square',
      color:            button.color       || '#3b82f6',
      actionController: button.action?.controller || '',
      actionType:       button.action?.type       || '',
      actionParams:     { ...(button.action?.params || {}) }
    };
    initialized = true;
    testResult = '';
  }

  // Create mode — blank form
  $: if (isOpen && !button && !initialized) {
    formData = {
      name: '', description: '', icon: 'square', color: '#3b82f6',
      actionController: '', actionType: '', actionParams: {}
    };
    testResult = '';
    initialized = true;
  }

  // ── Data loading ─────────────────────────────────────────────────────────────

  async function loadActionTypes() {
    try {
      if (!window.go?.main?.App) return;
      const groups = await window.go.main.App.GetAllActionTypes() || [];
      actionGroups = groups;
      flatActionTypes = groups.flatMap(g =>
        (g.actions || []).map(a => ({
          value:      a.type,
          label:      a.name,
          controller: g.controller_id,
          params:     (a.params || []).map(p => p.key)
        }))
      );

      // If no tab is active yet, set it from current action or default to first group
      if (!activeTab) {
        if (formData.actionController && groups.some(g => g.controller_id === formData.actionController)) {
          activeTab = formData.actionController;
        } else if (groups.length > 0) {
          activeTab = groups[0].controller_id;
        }
      }

      // Default actionType to first action of the active tab if none selected
      if (!formData.actionType && activeTab) {
        const group = groups.find(g => g.controller_id === activeTab);
        if (group?.actions?.length > 0) {
          formData.actionType = group.actions[0].type;
          formData.actionController = activeTab;
        }
      }
    } catch (err) {
      console.error('Failed to load action types:', err);
    }
  }

  async function loadOBSData() {
    if (loadingOBSData) return;
    loadingOBSData = true;
    try {
      if (window.go?.main?.App) {
        scenes = await window.go.main.App.GetScenes() || [];
        inputs = await window.go.main.App.GetInputs() || [];
      }
    } catch (err) {
      scenes = []; inputs = [];
    } finally {
      loadingOBSData = false;
    }
  }

  // ── Tab switching ─────────────────────────────────────────────────────────────

  function switchTab(controllerId) {
    activeTab = controllerId;
    // If the currently-selected action doesn't belong to this tab, pick the first action
    const belongs = flatActionTypes.some(
      a => a.value === formData.actionType && a.controller === controllerId
    );
    if (!belongs) {
      const group = actionGroups.find(g => g.controller_id === controllerId);
      const first = group?.actions?.[0];
      if (first) {
        formData.actionType       = first.type;
        formData.actionController = controllerId;
        formData.actionParams     = {};
      }
    }
  }

  // Keep actionController in sync whenever the user picks a different action
  $: {
    if (flatActionTypes.length > 0 && formData.actionType) {
      const found = flatActionTypes.find(a => a.value === formData.actionType);
      if (found) {
        formData.actionController = found.controller;
        activeTab                 = found.controller; // keep tab in sync
      }
    }
  }

  // ── Param management ─────────────────────────────────────────────────────────

  $: {
    const requiredParams = getRequiredParams();
    const newParams = {};
    for (const param of requiredParams) {
      if (param === 'scene_name') {
        newParams.scene_name = formData.actionParams.scene_name || (scenes[0] ?? '');
      } else if (param === 'input_name') {
        newParams.input_name = formData.actionParams.input_name || (inputs[0] ?? '');
      } else {
        newParams[param] = formData.actionParams[param] || '';
      }
    }
    if (formData.actionType) formData.actionParams = newParams;
  }

  function getRequiredParams() {
    if (flatActionTypes.length > 0) {
      const def = flatActionTypes.find(a => a.value === formData.actionType);
      return def?.params ?? [];
    }
    const def = obsActionsFallback.find(a => a.value === formData.actionType);
    return def?.params ?? [];
  }

  // ── Returns the list of actions to show for the currently-active tab ─────────

  $: activeGroupActions = (() => {
    if (actionGroups.length === 0) return obsActionsFallback; // loading fallback
    const group = actionGroups.find(g => g.controller_id === activeTab);
    return group?.actions ?? [];
  })();

  // ── Save / Test ───────────────────────────────────────────────────────────────

  function handleSave() {
    const buttonData = {
      name:        formData.name,
      description: formData.description,
      icon:        formData.icon,
      color:       formData.color,
      action: {
        controller: formData.actionController,
        type:       formData.actionType,
        params:     formData.actionParams
      }
    };
    if (button) buttonData.id = button.id;
    onSave(buttonData);
  }

  async function testAction() {
    testing = true; testResult = '';
    try {
      await window.go.main.App.ExecuteAction({
        controller: formData.actionController,
        type:       formData.actionType,
        params:     formData.actionParams
      });
      testResult = '✅ Action executed successfully!';
    } catch (err) {
      testResult = '❌ Error: ' + err;
    } finally {
      testing = false;
      setTimeout(() => { testResult = ''; }, 3000);
    }
  }

  // ── Overlay click ─────────────────────────────────────────────────────────────

  let mouseDownOnOverlay = false;
  function handleOverlayMouseDown(e) {
    if (e.target.classList.contains('modal-overlay')) mouseDownOnOverlay = true;
  }
  function handleOverlayClick(e) {
    if (mouseDownOnOverlay && e.target.classList.contains('modal-overlay')) onClose();
    mouseDownOnOverlay = false;
  }

  const icons = [
    { value: 'video',       label: '📹 Video' },
    { value: 'play',        label: '▶️ Play' },
    { value: 'pause',       label: '⏸️ Pause' },
    { value: 'stop-circle', label: '⏹️ Stop' },
    { value: 'circle',      label: '⏺️ Record' },
    { value: 'mic',         label: '🎤 Mic' },
    { value: 'mic-off',     label: '🎤 Mic Off' },
    { value: 'volume-2',    label: '🔊 Volume' },
    { value: 'volume-x',    label: '🔇 Mute' },
    { value: 'camera',      label: '📷 Camera' },
    { value: 'layout',      label: '🎬 Scene' },
    { value: 'monitor',     label: '🖥️ Monitor' },
    { value: 'hand',        label: '✋ Hand' },
    { value: 'phone-off',   label: '📵 Leave' },
    { value: 'square',      label: '⬜ Square' },
    { value: 'star',        label: '⭐ Star' },
  ];
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

        <!-- Name -->
        <div class="form-group">
          <label>Name *</label>
          <input type="text" bind:value={formData.name} placeholder="Go Live" />
        </div>

        <!-- Description -->
        <div class="form-group">
          <label>Description</label>
          <input type="text" bind:value={formData.description} placeholder="Start streaming to Twitch" />
        </div>

        <!-- Icon + Color -->
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

        <!-- ── Action Picker ──────────────────────────────────────────── -->
        <div class="form-group">
          <label>Action</label>

          <!-- Controller pills -->
          <div class="ctrl-tabs">
            {#if actionGroups.length > 0}
              {#each actionGroups as group}
                <button
                  class="ctrl-tab"
                  class:active={activeTab === group.controller_id}
                  class:offline={!group.connected}
                  on:click={() => switchTab(group.controller_id)}
                  title={group.connected ? group.controller_name : group.controller_name + ' (offline)'}
                >
                  {group.controller_name}
                  {#if !group.connected}<span class="offline-dot" title="Not connected">●</span>{/if}
                </button>
              {/each}
            {:else}
              <!-- Fallback pill while loading -->
              <button class="ctrl-tab active">OBS Studio</button>
            {/if}
          </div>

          <!-- Actions for the active tab only -->
          <select bind:value={formData.actionType} class="action-select">
            {#if actionGroups.length > 0}
              {#each activeGroupActions as action}
                <option value={action.type}>{action.name}</option>
              {/each}
            {:else}
              {#each obsActionsFallback as action}
                <option value={action.value}>{action.label}</option>
              {/each}
            {/if}
          </select>
        </div>
        <!-- ── / Action Picker ─────────────────────────────────────────── -->

        <!-- Dynamic params for the chosen action -->
        {#key formData.actionType}
          {#each getRequiredParams() as param}
            <div class="form-group">
              {#if param === 'scene_name'}
                <label>Scene Name</label>
                {#if scenes.length > 0}
                  <select bind:value={formData.actionParams[param]}>
                    {#each scenes as scene}<option value={scene}>{scene}</option>{/each}
                  </select>
                {:else}
                  <input type="text" bind:value={formData.actionParams[param]} placeholder="Main" />
                  <p class="help-text">OBS not connected — enter scene name manually</p>
                {/if}

              {:else if param === 'input_name'}
                <label>Input Name</label>
                {#if inputs.length > 0}
                  <select bind:value={formData.actionParams[param]}>
                    {#each inputs as inp}<option value={inp}>{inp}</option>{/each}
                  </select>
                {:else}
                  <input type="text" bind:value={formData.actionParams[param]} placeholder="Mic/Aux" />
                  <p class="help-text">OBS not connected — enter input name manually</p>
                {/if}

              {:else if param === 'source_name'}
                <label>Source Name</label>
                <input type="text" bind:value={formData.actionParams[param]} placeholder="Webcam" />
                <p class="help-text">Exact source name from OBS (case-sensitive)</p>

              {:else if param === 'filter_name'}
                <label>Filter Name</label>
                <input type="text" bind:value={formData.actionParams[param]} placeholder="Color Correction" />

              {:else if param === 'transition_name'}
                <label>Transition Name</label>
                <input type="text" bind:value={formData.actionParams[param]} placeholder="Fade" />
                <p class="help-text">Common: Fade, Cut, Slide, Stinger</p>

              {:else if param === 'volume'}
                <label>Volume (%)</label>
                <input type="number" min="0" max="100" bind:value={formData.actionParams[param]} placeholder="50" />

              {:else if param === 'duration'}
                <label>Duration (ms)</label>
                <input type="number" min="0" bind:value={formData.actionParams[param]} placeholder="300" />

              {:else}
                <label>{param.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')}</label>
                <input type="text" bind:value={formData.actionParams[param]} placeholder={param} />
              {/if}
            </div>
          {/each}
        {/key}

        <!-- Button preview -->
        <div class="button-preview" style="background: {formData.color}">
          <i data-lucide={formData.icon}></i>
          <span>{formData.name || 'Preview'}</span>
        </div>

        {#if testResult}
          <div class="test-result" class:success={testResult.startsWith('✅')} class:error={testResult.startsWith('❌')}>
            {testResult}
          </div>
        {/if}
      </div><!-- /modal-body -->

      <div class="modal-footer">
        <button class="btn-test" on:click={testAction} disabled={testing || !formData.name}>
          {testing ? 'Testing…' : '🧪 Test Action'}
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
  /* ── Modal shell ─────────────────────────────────────────────────────────── */
  .modal-overlay {
    position: fixed; inset: 0;
    background: rgba(0,0,0,0.7);
    display: flex; align-items: center; justify-content: center;
    z-index: 1000;
  }
  .modal {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    width: 90%; max-width: 600px;
    max-height: 90vh; overflow-y: auto;
  }
  .modal-header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #0f3460;
  }
  .modal-header h2 { font-size: 20px; margin: 0; }
  .close-btn {
    background: none; border: none; color: #94a3b8;
    font-size: 32px; cursor: pointer; line-height: 1;
    padding: 0; width: 32px; height: 32px;
  }
  .close-btn:hover { color: #eaeaea; }

  /* ── Body / form ─────────────────────────────────────────────────────────── */
  .modal-body { padding: 24px; }
  .form-group { margin-bottom: 20px; }
  .form-group label {
    display: block; font-size: 14px; font-weight: 500; margin-bottom: 8px;
  }
  .form-group input,
  .form-group select {
    width: 100%; padding: 10px 12px;
    background: #0f1419; border: 1px solid #0f3460;
    border-radius: 6px; color: #eaeaea; font-size: 14px;
    box-sizing: border-box;
  }
  .form-group input::placeholder { color: #64748b; opacity: 1; }
  .form-group input:focus,
  .form-group select:focus { outline: none; border-color: #3b82f6; }
  .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

  /* ── Controller pill tabs ─────────────────────────────────────────────────── */
  .ctrl-tabs {
    display: flex; gap: 8px; flex-wrap: wrap;
    margin-bottom: 10px;
  }
  .ctrl-tab {
    display: flex; align-items: center; gap: 5px;
    padding: 5px 14px;
    border-radius: 20px;
    border: 1px solid #0f3460;
    background: transparent;
    color: #94a3b8;
    font-size: 12px; font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
    white-space: nowrap;
  }
  .ctrl-tab:hover { border-color: #3b82f6; color: #eaeaea; }
  .ctrl-tab.active { background: #3b82f6; border-color: #3b82f6; color: white; }
  .ctrl-tab.offline { opacity: 0.65; }
  .ctrl-tab.offline.active { background: #64748b; border-color: #64748b; }
  .offline-dot { color: #fbbf24; font-size: 8px; line-height: 1; }

  /* ── Action select ────────────────────────────────────────────────────────── */
  .action-select { margin-top: 0; }
  .action-select option { color: #eaeaea; font-size: 14px; }

  /* ── Misc ─────────────────────────────────────────────────────────────────── */
  .help-text { margin-top: 4px; font-size: 12px; color: #94a3b8; font-style: italic; }

  .button-preview {
    margin-top: 24px; padding: 24px; border-radius: 8px;
    display: flex; flex-direction: column; align-items: center; gap: 12px;
  }
  .button-preview i { width: 32px; height: 32px; color: white; }
  .button-preview span { color: white; font-size: 14px; font-weight: 500; }

  .test-result {
    margin-top: 16px; padding: 12px 16px;
    border-radius: 6px; font-size: 14px; font-weight: 500;
  }
  .test-result.success { background:#10b98120; border:1px solid #10b981; color:#10b981; }
  .test-result.error   { background:#ef444420; border:1px solid #ef4444; color:#ef4444; }

  /* ── Footer ──────────────────────────────────────────────────────────────── */
  .modal-footer {
    padding: 16px 24px; border-top: 1px solid #0f3460;
    display: flex; gap: 12px; align-items: center;
  }
  .spacer { flex: 1; }
  .btn-primary, .btn-secondary, .btn-test {
    padding: 10px 20px; border-radius: 6px;
    font-size: 14px; font-weight: 500; cursor: pointer; border: none;
  }
  .btn-primary { background: #3b82f6; color: white; }
  .btn-primary:hover:not(:disabled) { background: #2563eb; }
  .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-secondary { background: transparent; border: 1px solid #0f3460; color: #eaeaea; }
  .btn-secondary:hover { background: #0f3460; }
  .btn-test { background: #10b981; color: white; }
  .btn-test:hover:not(:disabled) { background: #059669; }
  .btn-test:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
