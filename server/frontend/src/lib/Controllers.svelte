<script>
  import { onMount, onDestroy } from 'svelte';

  let controllers = [];
  let statuses = {};       // id → status map
  let schemas = {};        // id → config schema array
  let configs = {};        // id → current config
  let editConfigs = {};    // id → editable copy of config
  let expanded = {};       // id → bool (config panel open)
  let saving = {};         // id → bool
  let errorMsg = {};       // id → error string
  let successMsg = {};     // id → success string
  let interval;

  onMount(async () => {
    await refresh();
    interval = setInterval(refresh, 5000);
  });

  onDestroy(() => clearInterval(interval));

  async function refresh() {
    try {
      controllers = await window.go.main.App.GetControllers();
      for (const c of controllers) {
        statuses[c.id] = await window.go.main.App.GetControllerStatus(c.id);
      }
      statuses = statuses; // trigger reactivity
    } catch (err) {
      console.error('Controllers refresh failed:', err);
    }
  }

  async function toggleExpand(id) {
    expanded[id] = !expanded[id];
    if (expanded[id] && !schemas[id]) {
      // Load schema + config on first open
      try {
        schemas[id] = await window.go.main.App.GetControllerConfigSchema(id);
        configs[id] = await window.go.main.App.GetControllerConfig(id);
        editConfigs[id] = { ...configs[id] };
      } catch (err) {
        errorMsg[id] = 'Failed to load config: ' + err;
      }
    }
    expanded = expanded;
  }

  async function save(id) {
    saving[id] = true;
    errorMsg[id] = '';
    successMsg[id] = '';
    try {
      await window.go.main.App.ConnectController(id, editConfigs[id]);
      successMsg[id] = 'Saved and applied.';
      // Refresh status
      await new Promise(r => setTimeout(r, 600));
      statuses[id] = await window.go.main.App.GetControllerStatus(id);
      statuses = statuses;
    } catch (err) {
      errorMsg[id] = '' + err;
    } finally {
      saving[id] = false;
    }
  }

  async function disconnect(id) {
    errorMsg[id] = '';
    successMsg[id] = '';
    try {
      await window.go.main.App.DisconnectController(id);
      await new Promise(r => setTimeout(r, 400));
      statuses[id] = await window.go.main.App.GetControllerStatus(id);
      statuses = statuses;
    } catch (err) {
      errorMsg[id] = '' + err;
    }
  }

  function stateIcon(val) {
    return val ? '✓' : '✗';
  }

  function controllerIcon(id) {
    if (id === 'obs') return 'video';
    if (id === 'zoomosc') return 'users';
    return 'cpu';
  }
</script>

<div class="controllers">
  <header>
    <h2>Controllers</h2>
    <p>Manage connected applications and installed plugins</p>
  </header>

  {#if controllers.length === 0}
    <div class="empty-state">
      <p>No controllers registered yet.</p>
    </div>
  {/if}

  {#each controllers as ctrl (ctrl.id)}
    {@const status = statuses[ctrl.id] || {}}
    {@const isConnected = status.connected === true}

    <div class="card">
      <!-- Card header -->
      <div class="card-header">
        <div class="ctrl-info">
          <i data-lucide={controllerIcon(ctrl.id)} class="ctrl-icon"></i>
          <div>
            <h3>{ctrl.name}</h3>
            <p class="ctrl-desc">{ctrl.description}</p>
          </div>
        </div>
        <div class="card-actions">
          <div class="badge {isConnected ? 'connected' : 'disconnected'}">
            {isConnected ? 'Connected' : 'Disconnected'}
          </div>
          <button class="btn-icon" on:click={() => toggleExpand(ctrl.id)} title="Configure">
            <i data-lucide={expanded[ctrl.id] ? 'chevron-up' : 'settings'}></i>
          </button>
        </div>
      </div>

      <!-- Status details row (shown when connected) -->
      {#if isConnected && ctrl.id === 'zoomosc'}
        <div class="state-row">
          <span class="state-chip {status.muted ? 'active' : ''}">
            🎙 {status.muted ? 'Muted' : 'Unmuted'}
          </span>
          <span class="state-chip {status.video ? 'active' : ''}">
            📷 Video {status.video ? 'On' : 'Off'}
          </span>
          <span class="state-chip {status.sharing ? 'active' : ''}">
            🖥 {status.sharing ? 'Sharing' : 'Not Sharing'}
          </span>
        </div>
      {/if}

      {#if isConnected && ctrl.id === 'obs'}
        <div class="state-row">
          {#if status.streaming}
            <span class="state-chip active">🔴 Streaming</span>
          {/if}
          {#if status.recording}
            <span class="state-chip active">⏺ Recording</span>
          {/if}
          {#if status.scene}
            <span class="state-chip">Scene: {status.scene}</span>
          {/if}
        </div>
      {/if}

      <!-- Error from status -->
      {#if !isConnected && status.error}
        <p class="status-error">{status.error}</p>
      {/if}

      <!-- Collapsible config panel -->
      {#if expanded[ctrl.id]}
        <div class="config-panel">
          <hr class="divider" />

          {#if errorMsg[ctrl.id]}
            <div class="msg error">{errorMsg[ctrl.id]}</div>
          {/if}
          {#if successMsg[ctrl.id]}
            <div class="msg success">{successMsg[ctrl.id]}</div>
          {/if}

          {#if schemas[ctrl.id]}
            {#each schemas[ctrl.id] as field}
              <div class="form-group">
                <label for="{ctrl.id}-{field.key}">{field.label}</label>
                {#if field.type === 'password'}
                  <input id="{ctrl.id}-{field.key}" type="password"
                    bind:value={editConfigs[ctrl.id][field.key]}
                    placeholder={field.default || ''}
                  />
                {:else if field.type === 'number'}
                  <input id="{ctrl.id}-{field.key}" type="number"
                    bind:value={editConfigs[ctrl.id][field.key]}
                    placeholder={field.default || ''}
                  />
                {:else}
                  <input id="{ctrl.id}-{field.key}" type="text"
                    bind:value={editConfigs[ctrl.id][field.key]}
                    placeholder={field.default || ''}
                  />
                {/if}
                {#if field.help}
                  <p class="help">{field.help}</p>
                {/if}
              </div>
            {/each}
          {:else}
            <p class="loading">Loading config…</p>
          {/if}

          <!-- ZoomOSC-specific setup instructions -->
          {#if ctrl.id === 'zoomosc'}
            <div class="instructions">
              <h4>ZoomOSC Setup</h4>
              <p>In ZoomOSC → Settings, set:</p>
              <ul>
                <li><strong>Transmission IP:</strong> 127.0.0.1</li>
                <li><strong>Transmission Port:</strong> 1234 (matches Feedback Listen Port above)</li>
                <li><strong>Receiving Port:</strong> 9090 (matches ZoomOSC Receiving Port above)</li>
              </ul>
              <p class="warn">⚠️ Always start or join meetings <strong>through ZoomOSC</strong> (its built-in Start/Join buttons), not through the Zoom app. ZoomOSC must own the meeting session to control audio and video.</p>
            </div>
          {/if}

          <div class="config-footer">
            <button class="btn-primary" on:click={() => save(ctrl.id)} disabled={saving[ctrl.id]}>
              {saving[ctrl.id] ? 'Saving…' : 'Save & Connect'}
            </button>
            {#if isConnected}
              <button class="btn-danger" on:click={() => disconnect(ctrl.id)}>
                Disconnect
              </button>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  {/each}

  <div class="help-card">
    <h3>Installing Plugins</h3>
    <p>Place a plugin binary in the plugins directory and restart the server:</p>
    <code>~/.robo-stream-server/plugins/&lt;plugin-id&gt;/&lt;plugin-binary&gt;</code>
    <p class="mt">The server discovers and starts plugins automatically on launch.</p>
  </div>
</div>

<style>
  .controllers {
    padding: 32px;
    max-width: 860px;
  }

  header {
    margin-bottom: 28px;
  }

  header h2 {
    font-size: 28px;
    margin-bottom: 6px;
  }

  header p {
    color: #94a3b8;
    font-size: 14px;
  }

  .card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    padding: 20px 24px;
    margin-bottom: 16px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .ctrl-info {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .ctrl-icon {
    width: 28px;
    height: 28px;
    color: #3b82f6;
  }

  .ctrl-info h3 {
    font-size: 17px;
    margin: 0 0 3px;
  }

  .ctrl-desc {
    font-size: 12px;
    color: #94a3b8;
    margin: 0;
    max-width: 480px;
  }

  .card-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .badge {
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 600;
  }

  .badge.connected {
    background: #10b981;
    color: white;
  }

  .badge.disconnected {
    background: #374151;
    color: #9ca3af;
  }

  .btn-icon {
    background: transparent;
    border: 1px solid #0f3460;
    border-radius: 6px;
    color: #94a3b8;
    cursor: pointer;
    padding: 6px;
    display: flex;
    align-items: center;
  }

  .btn-icon:hover {
    background: #0f3460;
    color: #eaeaea;
  }

  .btn-icon i {
    width: 16px;
    height: 16px;
  }

  /* State chips (muted/video/sharing) */
  .state-row {
    display: flex;
    gap: 8px;
    margin-top: 12px;
    flex-wrap: wrap;
  }

  .state-chip {
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 12px;
    background: #1e293b;
    color: #64748b;
    border: 1px solid #0f3460;
  }

  .state-chip.active {
    background: #1e3a5f;
    color: #93c5fd;
    border-color: #3b82f6;
  }

  .status-error {
    margin-top: 10px;
    font-size: 12px;
    color: #f87171;
  }

  /* Config panel */
  .divider {
    border: none;
    border-top: 1px solid #0f3460;
    margin: 16px 0;
  }

  .config-panel {
    margin-top: 4px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 16px;
  }

  .form-group label {
    font-size: 13px;
    font-weight: 500;
    color: #cbd5e1;
  }

  .form-group input {
    padding: 9px 12px;
    background: #0f1419;
    border: 1px solid #0f3460;
    border-radius: 6px;
    color: #eaeaea;
    font-size: 14px;
    max-width: 320px;
  }

  .form-group input:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .help {
    font-size: 11px;
    color: #64748b;
    margin: 0;
  }

  .loading {
    color: #64748b;
    font-size: 13px;
    font-style: italic;
  }

  /* Instructions block */
  .instructions {
    background: #0f1419;
    border: 1px solid #0f3460;
    border-radius: 8px;
    padding: 14px 16px;
    margin-bottom: 16px;
    font-size: 13px;
  }

  .instructions h4 {
    margin: 0 0 8px;
    font-size: 13px;
    color: #3b82f6;
  }

  .instructions p {
    margin: 0 0 6px;
    color: #94a3b8;
  }

  .instructions ul {
    margin: 0;
    padding-left: 16px;
    color: #94a3b8;
  }

  .instructions li {
    margin-bottom: 4px;
  }

  .instructions strong {
    color: #cbd5e1;
  }

  .instructions .warn {
    margin: 8px 0 0;
    color: #fbbf24;
  }

  .config-footer {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-top: 4px;
  }

  .btn-primary {
    padding: 10px 18px;
    background: #3b82f6;
    border: none;
    border-radius: 8px;
    color: white;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
  }

  .btn-primary:hover:not(:disabled) { background: #2563eb; }
  .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

  .btn-danger {
    padding: 10px 16px;
    background: transparent;
    border: 1px solid #ef4444;
    border-radius: 8px;
    color: #ef4444;
    font-size: 13px;
    cursor: pointer;
  }

  .btn-danger:hover { background: #ef4444; color: white; }

  .msg {
    padding: 10px 14px;
    border-radius: 6px;
    font-size: 13px;
    margin-bottom: 14px;
  }

  .msg.error { background: #7f1d1d; color: #fecaca; border: 1px solid #ef4444; }
  .msg.success { background: #064e3b; color: #a7f3d0; border: 1px solid #10b981; }

  /* Help card at bottom */
  .help-card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    padding: 20px 24px;
    margin-top: 8px;
  }

  .help-card h3 {
    font-size: 15px;
    margin-bottom: 10px;
  }

  .help-card p {
    font-size: 13px;
    color: #94a3b8;
    margin: 0 0 8px;
  }

  .help-card code {
    display: block;
    background: #0f1419;
    border: 1px solid #0f3460;
    border-radius: 6px;
    padding: 10px 14px;
    font-size: 12px;
    color: #93c5fd;
    font-family: monospace;
  }

  .mt { margin-top: 10px !important; }

  .empty-state {
    text-align: center;
    color: #64748b;
    padding: 48px;
  }
</style>
