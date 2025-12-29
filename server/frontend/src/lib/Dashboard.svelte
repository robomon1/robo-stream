<script>
  export let serverInfo = {};
  export let obsStatus = {};
  export let onSwitchView = () => {};

  function copyToClipboard(text) {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(text).then(() => {
        // Could add a toast notification here
        console.log('Copied to clipboard:', text);
      }).catch(err => {
        console.error('Failed to copy:', err);
      });
    }
  }
</script>

<div class="dashboard">
  <header>
    <h2>Dashboard</h2>
    <p>Server overview and status</p>
  </header>

  <div class="cards">
    <div class="card">
      <div class="card-header">
        <i data-lucide="activity"></i>
        <h3>Server Status</h3>
      </div>
      <div class="card-body">
        <div class="info-row">
          <span>Version:</span>
          <span>{serverInfo.version || '1.0.0'}</span>
        </div>
        <div class="info-row">
          <span>API Port:</span>
          <span>{serverInfo.api_port || 8080}</span>
        </div>
        <div class="info-row">
          <span>Active Sessions:</span>
          <span>{serverInfo.active_sessions || 0}</span>
        </div>
        {#if serverInfo.ip_addresses && serverInfo.ip_addresses.length > 0}
          <div class="info-section">
            <h4>Client URLs:</h4>
            {#each serverInfo.client_urls || [] as url}
              <div class="client-url">
                <code>{url}</code>
                <button class="copy-btn" on:click={() => copyToClipboard(url)} title="Copy to clipboard">
                  <i data-lucide="copy"></i>
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <i data-lucide="video"></i>
        <h3>OBS Status</h3>
      </div>
      <div class="card-body">
        <div class="info-row">
          <span>Connection:</span>
          <span class="{obsStatus.connected ? 'status-ok' : 'status-error'}">
            {obsStatus.connected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
        {#if obsStatus.connected}
          <div class="info-row">
            <span>Streaming:</span>
            <span class="{obsStatus.streaming ? 'status-ok' : 'status-inactive'}">
              {obsStatus.streaming ? 'Active' : 'Inactive'}
            </span>
          </div>
          <div class="info-row">
            <span>Recording:</span>
            <span class="{obsStatus.recording ? 'status-ok' : 'status-inactive'}">
              {obsStatus.recording ? 'Active' : 'Inactive'}
            </span>
          </div>
          <div class="info-row">
            <span>Current Scene:</span>
            <span>{obsStatus.current_scene || 'N/A'}</span>
          </div>
        {/if}
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <i data-lucide="database"></i>
        <h3>Configuration</h3>
      </div>
      <div class="card-body">
        <div class="info-row">
          <span>Total Configs:</span>
          <span>{serverInfo.configurations || 0}</span>
        </div>
        <div class="info-row">
          <span>Total Buttons:</span>
          <span>{serverInfo.buttons || 0}</span>
        </div>
      </div>
    </div>
  </div>

  <div class="quick-actions">
    <h3>Quick Actions</h3>
    <div class="actions-grid">
      <button class="action-btn" on:click={() => onSwitchView('configurations')}>
        <i data-lucide="plus"></i>
        New Configuration
      </button>
      <button class="action-btn" on:click={() => onSwitchView('buttons')}>
        <i data-lucide="square"></i>
        New Button
      </button>
      <button class="action-btn" on:click={() => onSwitchView('obs')}>
        <i data-lucide="settings"></i>
        OBS Settings
      </button>
    </div>
  </div>
</div>

<style>
  .dashboard {
    padding: 32px;
    max-width: 1200px;
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

  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 20px;
    margin-bottom: 32px;
  }

  .card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    overflow: hidden;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 20px;
    border-bottom: 1px solid #0f3460;
  }

  .card-header i {
    width: 20px;
    height: 20px;
    color: #3b82f6;
  }

  .card-header h3 {
    font-size: 16px;
    font-weight: 600;
  }

  .card-body {
    padding: 20px;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    font-size: 14px;
  }

  .info-row span:first-child {
    color: #94a3b8;
  }

  .status-ok {
    color: #10b981;
    font-weight: 500;
  }

  .status-error {
    color: #ef4444;
    font-weight: 500;
  }

  .status-inactive {
    color: #94a3b8;
  }

  .info-section {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #0f3460;
  }

  .info-section h4 {
    font-size: 13px;
    color: #94a3b8;
    margin-bottom: 12px;
  }

  .client-url {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: #0f1419;
    border: 1px solid #0f3460;
    border-radius: 6px;
    margin-bottom: 8px;
  }

  .client-url code {
    flex: 1;
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 13px;
    color: #3b82f6;
  }

  .copy-btn {
    padding: 4px 8px;
    background: transparent;
    border: 1px solid #0f3460;
    border-radius: 4px;
    color: #94a3b8;
    cursor: pointer;
    display: flex;
    align-items: center;
    transition: all 0.2s;
  }

  .copy-btn:hover {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }

  .copy-btn i {
    width: 14px;
    height: 14px;
  }

  .quick-actions h3 {
    font-size: 18px;
    margin-bottom: 16px;
  }

  .actions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
  }

  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 16px;
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 8px;
    color: #eaeaea;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .action-btn:hover {
    background: #3b82f6;
    border-color: #3b82f6;
  }

  .action-btn i {
    width: 18px;
    height: 18px;
  }
</style>
