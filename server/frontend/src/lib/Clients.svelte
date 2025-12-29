<script>
  import { onMount } from 'svelte';

  let sessions = [];
  let configurations = [];
  let loading = true;

  onMount(async () => {
    await loadData();
    setInterval(loadData, 5000); // Refresh every 5s
  });

  async function loadData() {
    try {
      sessions = await window.go.main.App.GetSessions();
      configurations = await window.go.main.App.GetConfigurations();
    } catch (err) {
      console.error('Failed to load data:', err);
    } finally {
      loading = false;
    }
  }

  function getConfigName(configId) {
    const config = configurations.find(c => c.id === configId);
    return config ? config.name : 'Unknown';
  }

  function formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = Math.floor((now - date) / 1000); // seconds

    if (diff < 60) return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return `${Math.floor(diff / 86400)}d ago`;
  }
</script>

<div class="clients">
  <header>
    <div>
      <h2>Connected Clients</h2>
      <p>Monitor and manage client connections</p>
    </div>
  </header>

  {#if loading}
    <div class="loading">Loading clients...</div>
  {:else if sessions.length === 0}
    <div class="empty">
      <i data-lucide="users"></i>
      <h3>No clients connected</h3>
      <p>Clients will appear here when they connect to the server</p>
    </div>
  {:else}
    <div class="client-list">
      {#each sessions as session}
        <div class="client-card">
          <div class="client-header">
            <div class="client-avatar">
              <i data-lucide="monitor"></i>
            </div>
            <div class="client-info">
              <h3>{session.client_name || session.client_id}</h3>
              <p class="client-id">{session.client_id}</p>
            </div>
            <div class="client-status">
              <span class="indicator active"></span>
              <span>Active</span>
            </div>
          </div>
          <div class="client-details">
            <div class="detail-row">
              <span class="detail-label">IP Address:</span>
              <span class="detail-value">{session.ip_address}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Configuration:</span>
              <span class="detail-value">{getConfigName(session.config_id)}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Last Active:</span>
              <span class="detail-value">{formatTime(session.last_active)}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Connected:</span>
              <span class="detail-value">{formatTime(session.last_connected)}</span>
            </div>
          </div>
          <div class="client-actions">
            <select class="config-select">
              <option value="">Change Configuration</option>
              {#each configurations as config}
                <option value={config.id} selected={config.id === session.config_id}>
                  {config.name}
                </option>
              {/each}
            </select>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .clients {
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

  .client-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .client-card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 12px;
    overflow: hidden;
  }

  .client-header {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    border-bottom: 1px solid #0f3460;
  }

  .client-avatar {
    width: 48px;
    height: 48px;
    background: #3b82f6;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .client-avatar i {
    width: 24px;
    height: 24px;
    color: white;
  }

  .client-info {
    flex: 1;
  }

  .client-info h3 {
    font-size: 16px;
    margin-bottom: 4px;
  }

  .client-id {
    font-size: 12px;
    color: #94a3b8;
  }

  .client-status {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: #10b981;
  }

  .indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #10b981;
  }

  .client-details {
    padding: 20px;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
  }

  .detail-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .detail-label {
    font-size: 12px;
    color: #94a3b8;
  }

  .detail-value {
    font-size: 14px;
  }

  .client-actions {
    padding: 16px 20px;
    border-top: 1px solid #0f3460;
    background: #0f1419;
  }

  .config-select {
    width: 100%;
    padding: 8px 12px;
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 6px;
    color: #eaeaea;
    font-size: 14px;
  }
</style>
