<script>
  import { onMount } from 'svelte';
  import Dashboard from './lib/Dashboard.svelte';
  import ConfigurationEditor from './lib/ConfigurationEditor.svelte';
  import ButtonLibrary from './lib/ButtonLibrary.svelte';
  import Clients from './lib/Clients.svelte';
  import OBSSettings from './lib/OBSSettings.svelte';

  let currentView = 'dashboard';
  let serverInfo = {};
  let obsStatus = {};

  onMount(async () => {
    loadServerInfo();
    loadOBSStatus();
    
    // Refresh every 5 seconds
    setInterval(loadServerInfo, 5000);
    setInterval(loadOBSStatus, 5000);
    
    // Initialize icons
    if (window.lucide) {
      lucide.createIcons();
    }
  });

  async function loadServerInfo() {
    try {
      serverInfo = await window.go.main.App.GetServerInfo();
    } catch (err) {
      console.error('Failed to load server info:', err);
    }
  }

  async function loadOBSStatus() {
    try {
      obsStatus = await window.go.main.App.GetOBSStatus();
    } catch (err) {
      console.error('Failed to load OBS status:', err);
    }
  }

  function switchView(view) {
    currentView = view;
    // Reinitialize icons after view change
    setTimeout(() => {
      if (window.lucide) {
        lucide.createIcons();
      }
    }, 100);
  }
</script>

<main>
  <div class="app-container">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="logo">
        <h1>Robo-Stream</h1>
        <p>Server</p>
      </div>

      <div class="status-panel">
        <div class="status-item">
          <span class="indicator {obsStatus.connected ? 'connected' : 'disconnected'}"></span>
          <span>{obsStatus.connected ? 'OBS Connected' : 'OBS Disconnected'}</span>
        </div>
        <div class="stats">
          <div class="stat">
            <span class="stat-value">{serverInfo.active_sessions || 0}</span>
            <span class="stat-label">Clients</span>
          </div>
          <div class="stat">
            <span class="stat-value">{serverInfo.configurations || 0}</span>
            <span class="stat-label">Configs</span>
          </div>
          <div class="stat">
            <span class="stat-value">{serverInfo.buttons || 0}</span>
            <span class="stat-label">Buttons</span>
          </div>
        </div>
      </div>

      <nav class="nav">
        <button 
          class="nav-item {currentView === 'dashboard' ? 'active' : ''}"
          on:click={() => switchView('dashboard')}
        >
          <i data-lucide="layout-dashboard"></i>
          Dashboard
        </button>
        <button 
          class="nav-item {currentView === 'configurations' ? 'active' : ''}"
          on:click={() => switchView('configurations')}
        >
          <i data-lucide="grid-3x3"></i>
          Configurations
        </button>
        <button 
          class="nav-item {currentView === 'buttons' ? 'active' : ''}"
          on:click={() => switchView('buttons')}
        >
          <i data-lucide="square"></i>
          Button Library
        </button>
        <button 
          class="nav-item {currentView === 'clients' ? 'active' : ''}"
          on:click={() => switchView('clients')}
        >
          <i data-lucide="users"></i>
          Clients
        </button>
        <button 
          class="nav-item {currentView === 'obs' ? 'active' : ''}"
          on:click={() => switchView('obs')}
        >
          <i data-lucide="settings"></i>
          OBS Settings
        </button>
      </nav>
    </aside>

    <!-- Main Content -->
    <div class="main-content">
      {#if currentView === 'dashboard'}
        <Dashboard {serverInfo} {obsStatus} onSwitchView={switchView} />
      {:else if currentView === 'configurations'}
        <ConfigurationEditor />
      {:else if currentView === 'buttons'}
        <ButtonLibrary />
      {:else if currentView === 'clients'}
        <Clients />
      {:else if currentView === 'obs'}
        <OBSSettings />
      {/if}
    </div>
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }

  main {
    width: 100vw;
    height: 100vh;
    overflow: hidden;
  }

  .app-container {
    display: flex;
    height: 100%;
  }

  .sidebar {
    width: 250px;
    background: #16213e;
    display: flex;
    flex-direction: column;
    border-right: 1px solid #0f3460;
  }

  .logo {
    padding: 24px 20px;
    border-bottom: 1px solid #0f3460;
  }

  .logo h1 {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .logo p {
    font-size: 12px;
    color: #94a3b8;
  }

  .status-panel {
    padding: 20px;
    border-bottom: 1px solid #0f3460;
  }

  .status-item {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
    font-size: 13px;
  }

  .indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    animation: pulse 2s infinite;
  }

  .indicator.connected {
    background: #10b981;
  }

  .indicator.disconnected {
    background: #ef4444;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .stats {
    display: flex;
    gap: 16px;
  }

  .stat {
    flex: 1;
    text-align: center;
  }

  .stat-value {
    display: block;
    font-size: 20px;
    font-weight: 600;
    color: #3b82f6;
    margin-bottom: 4px;
  }

  .stat-label {
    display: block;
    font-size: 11px;
    color: #94a3b8;
  }

  .nav {
    flex: 1;
    padding: 20px 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: transparent;
    border: none;
    color: #94a3b8;
    font-size: 14px;
    cursor: pointer;
    border-radius: 8px;
    transition: all 0.2s;
    text-align: left;
  }

  .nav-item:hover {
    background: #0f3460;
    color: #eaeaea;
  }

  .nav-item.active {
    background: #3b82f6;
    color: white;
  }

  .nav-item i {
    width: 18px;
    height: 18px;
  }

  .main-content {
    flex: 1;
    overflow-y: auto;
    background: #1a1a2e;
  }
</style>
