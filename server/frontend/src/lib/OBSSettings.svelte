<script>
  import { onMount } from 'svelte';

  let url = 'localhost:4455';
  let password = '';
  let connected = false;
  let connecting = false;
  let scenes = [];
  let inputs = [];
  let errorMessage = '';
  let wailsReady = false;

  onMount(async () => {
    // Check if Wails runtime is available
    if (window.go && window.go.main && window.go.main.App) {
      wailsReady = true;
      console.log('✅ Wails runtime detected');
      
      // Test the binding
      try {
        const testResult = await window.go.main.App.TestBinding('Hello from frontend');
        console.log('Binding test result:', testResult);
      } catch (err) {
        console.error('❌ Binding test failed:', err);
        errorMessage = 'Wails bindings not working: ' + err;
      }
      
      // Load saved credentials
      try {
        const savedConfig = await window.go.main.App.GetSavedOBSConfig();
        if (savedConfig) {
          url = savedConfig.url || 'localhost:4455';
          password = savedConfig.password || '';
          console.log('📋 Loaded saved OBS config:', url);
        }
      } catch (err) {
        console.error('Failed to load saved config:', err);
      }
      
      await checkStatus();
    } else {
      console.error('❌ Wails runtime not available');
      errorMessage = 'Wails runtime not loaded. Are you running in dev mode?';
    }
  });

  async function checkStatus() {
    if (!wailsReady) return;
    
    try {
      const status = await window.go.main.App.GetOBSStatus();
      console.log('OBS Status:', status);
      connected = status.connected;
      if (status.url) {
        url = status.url;
      }
      if (connected) {
        await loadOBSData();
      }
    } catch (err) {
      console.error('Failed to check OBS status:', err);
      errorMessage = 'Failed to check status: ' + err;
    }
  }

  async function loadOBSData() {
    if (!wailsReady) return;
    
    try {
      scenes = await window.go.main.App.GetScenes();
      inputs = await window.go.main.App.GetInputs();
      console.log('Loaded scenes:', scenes.length, 'inputs:', inputs.length);
    } catch (err) {
      console.error('Failed to load OBS data:', err);
    }
  }

  async function connect() {
    if (!wailsReady) {
      errorMessage = 'Wails runtime not ready';
      return;
    }

    connecting = true;
    errorMessage = '';
    console.log('🔌 Attempting to connect to OBS:', url);
    
    try {
      console.log('Calling ConnectOBS...');
      await window.go.main.App.ConnectOBS(url, password);
      console.log('✅ ConnectOBS call completed');
      
      // Wait a moment for connection to establish
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Verify connection
      console.log('Checking connection status...');
      const status = await window.go.main.App.GetOBSStatus();
      console.log('Status after connect:', status);
      
      if (status.connected) {
        connected = true;
        password = ''; // Clear password
        console.log('✅ Connected! Loading OBS data...');
        await loadOBSData();
        errorMessage = ''; // Clear any previous errors
      } else {
        throw new Error('Connection call succeeded but status shows not connected. Check OBS WebSocket settings.');
      }
    } catch (err) {
      console.error('❌ OBS connection error:', err);
      errorMessage = 'Connection failed: ' + err + '\n\nMake sure:\n• OBS is running\n• WebSocket server is enabled (Tools → WebSocket Server Settings)\n• URL is correct (default: ws://localhost:4455)\n• Password matches (if set)';
    } finally {
      connecting = false;
    }
  }

  async function disconnect() {
    if (!wailsReady) return;
    
    try {
      await window.go.main.App.DisconnectOBS();
      connected = false;
      scenes = [];
      inputs = [];
      errorMessage = '';
    } catch (err) {
      console.error('Failed to disconnect:', err);
      errorMessage = 'Disconnect failed: ' + err;
    }
  }
</script>

<div class="obs-settings">
  <header>
    <h2>OBS Settings</h2>
    <p>Configure OBS WebSocket connection</p>
  </header>

  <div class="connection-card">
    <div class="card-header">
      <h3>Connection</h3>
      <div class="status-badge {connected ? 'connected' : 'disconnected'}">
        {connected ? 'Connected' : 'Disconnected'}
      </div>
    </div>

    {#if errorMessage}
      <div class="error-message">
        <strong>⚠️ Error:</strong>
        <pre>{errorMessage}</pre>
      </div>
    {/if}

    {#if !wailsReady}
      <div class="warning-message">
        <strong>⚠️ Warning:</strong> Wails runtime not loaded. Make sure you're running with <code>wails dev</code>
      </div>
    {/if}

    {#if !connected}
      <div class="connection-form">
        <div class="form-group">
          <label for="obs-url">WebSocket URL</label>
          <input
            id="obs-url"
            type="text"
            bind:value={url}
            placeholder="localhost:4455"
            disabled={connecting}
          />
          <p class="help-text">Enter host:port (ws:// prefix not needed)</p>
        </div>

        <div class="form-group">
          <label for="obs-password">Password (if required)</label>
          <input
            id="obs-password"
            type="password"
            bind:value={password}
            placeholder="Enter password"
            disabled={connecting}
          />
        </div>

        <button class="btn-primary" on:click={connect} disabled={connecting}>
          {connecting ? 'Connecting...' : 'Connect to OBS'}
        </button>
      </div>
    {:else}
      <div class="connection-info">
        <div class="info-row">
          <span>URL:</span>
          <span>{url}</span>
        </div>
        <button class="btn-secondary" on:click={disconnect}>
          Disconnect
        </button>
      </div>

      <div class="obs-data">
        <div class="data-section">
          <h4>Scenes ({scenes.length})</h4>
          <div class="item-list">
            {#each scenes as scene}
              <div class="item">
                <i data-lucide="layout"></i>
                {scene}
              </div>
            {/each}
          </div>
        </div>

        <div class="data-section">
          <h4>Inputs ({inputs.length})</h4>
          <div class="item-list">
            {#each inputs as input}
              <div class="item">
                <i data-lucide="mic"></i>
                {input}
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>

  <div class="help-section">
    <h3>Setup Instructions</h3>
    <ol>
      <li>Open OBS Studio</li>
      <li>Go to Tools → WebSocket Server Settings</li>
      <li>Enable WebSocket server</li>
      <li>Note the server port (default: 4455)</li>
      <li>Set a password if desired</li>
      <li>Enter the details above and click Connect</li>
    </ol>
  </div>
</div>

<style>
  .obs-settings {
    padding: 32px;
    max-width: 800px;
  }

  header {
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

  .connection-card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    padding: 24px;
    margin-bottom: 24px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid #0f3460;
  }

  .card-header h3 {
    font-size: 18px;
  }

  .status-badge {
    padding: 6px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
  }

  .status-badge.connected {
    background: #10b981;
    color: white;
  }

  .status-badge.disconnected {
    background: #ef4444;
    color: white;
  }

  .error-message {
    margin-bottom: 16px;
    padding: 12px 16px;
    background: #7f1d1d;
    border: 1px solid #ef4444;
    border-radius: 8px;
    color: #fecaca;
  }

  .error-message strong {
    display: block;
    margin-bottom: 8px;
    color: #ef4444;
  }

  .error-message pre {
    margin: 0;
    white-space: pre-wrap;
    font-family: monospace;
    font-size: 12px;
  }

  .warning-message {
    margin-bottom: 16px;
    padding: 12px 16px;
    background: #78350f;
    border: 1px solid #f59e0b;
    border-radius: 8px;
    color: #fde68a;
  }

  .warning-message strong {
    display: block;
    margin-bottom: 4px;
    color: #f59e0b;
  }

  .warning-message code {
    background: #451a03;
    padding: 2px 6px;
    border-radius: 4px;
    font-family: monospace;
  }

  .connection-form {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .form-group label {
    font-size: 14px;
    font-weight: 500;
  }

  .form-group input {
    padding: 10px 12px;
    background: #0f1419;
    border: 1px solid #0f3460;
    border-radius: 6px;
    color: #eaeaea;
    font-size: 14px;
  }

  .form-group input:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .help-text {
    font-size: 12px;
    color: #94a3b8;
  }

  .btn-primary {
    padding: 12px 20px;
    background: #3b82f6;
    border: none;
    border-radius: 8px;
    color: white;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    padding: 10px 16px;
    background: transparent;
    border: 1px solid #ef4444;
    border-radius: 6px;
    color: #ef4444;
    font-size: 14px;
    cursor: pointer;
    margin-top: 16px;
  }

  .btn-secondary:hover {
    background: #ef4444;
    color: white;
  }

  .connection-info {
    display: flex;
    flex-direction: column;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    padding: 12px 0;
    font-size: 14px;
  }

  .info-row span:first-child {
    color: #94a3b8;
  }

  .obs-data {
    margin-top: 24px;
    padding-top: 24px;
    border-top: 1px solid #0f3460;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 24px;
  }

  .data-section h4 {
    font-size: 14px;
    margin-bottom: 12px;
    color: #94a3b8;
  }

  .item-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: #0f1419;
    border-radius: 6px;
    font-size: 13px;
  }

  .item i {
    width: 16px;
    height: 16px;
    color: #3b82f6;
  }

  .help-section {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    padding: 24px;
  }

  .help-section h3 {
    font-size: 16px;
    margin-bottom: 16px;
  }

  .help-section ol {
    padding-left: 20px;
  }

  .help-section li {
    margin-bottom: 8px;
    font-size: 14px;
    color: #94a3b8;
  }
</style>
