// Robo-Stream Web Client - Touchscreen Optimized
import { APIClient } from './api.js';
import { initializeNativeFeatures, hapticFeedback, isNativeApp } from './native.js';

let currentConfiguration = null;
let apiClient = null;
let obsStatus = {
  streaming: false,
  recording: false,
  currentScene: '',
  virtualCamActive: false,
  replayBufferActive: false,
  studioModeActive: false
};
let sourceVisibility = {};

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('Robo-Stream Web Client starting...');
    
    // Setup event listeners
    setupEventListeners();
    
    // Initialize app
    initializeApp();
    
    // Initialize Lucide icons
    lucide.createIcons();
});

// Initialize application
async function initializeApp() {
    try {
        // Load server URL from localStorage
        const serverURL = localStorage.getItem('server_url') || 'http://localhost:8080';
        document.getElementById('input-server-url').value = serverURL;
        
        // Create API client
        apiClient = new APIClient(serverURL);
        
        // Connect and load
        await connectAndLoad();
    } catch (err) {
        console.error('Initialization error:', err);
        showConnectionBanner('Failed to initialize: ' + err.message, 'error');
    }
}

// Connect to server and load configuration
async function connectAndLoad() {
    try {
        // Test connection
        showConnectionBanner('Connecting to server...', 'connecting');
        await apiClient.getServerInfo();
        showConnectionBanner('Connected to server', 'connected');
        
        // Register to get session ID
        console.log('Registering with server...');
        const config = await apiClient.register();
        handleConfigurationLoaded(config);
        
        // Start status polling
        startStatusPolling();
        
        setTimeout(() => hideConnectionBanner(), 2000);
    } catch (err) {
        console.error('Connection error:', err);
        showConnectionBanner('Failed to connect: ' + err.message, 'error');
        setTimeout(() => hideConnectionBanner(), 2000);
    }
}

// Setup event listeners
function setupEventListeners() {
    // Settings button
    document.getElementById('btn-settings').addEventListener('click', openSettings);
    document.getElementById('btn-close-settings-modal').addEventListener('click', closeSettings);
    document.getElementById('btn-update-server').addEventListener('click', updateServerURL);

    // Config selector
    document.getElementById('btn-select-config').addEventListener('click', openConfigSelector);
    document.getElementById('btn-close-config-modal').addEventListener('click', closeConfigSelector);

    // Fullscreen button
    document.getElementById('btn-fullscreen').addEventListener('click', toggleFullscreen);
}

// Toggle fullscreen
function toggleFullscreen() {
    if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen();
    } else {
        document.exitFullscreen();
    }
}

// Handle configuration loaded
function handleConfigurationLoaded(config) {
    console.log('Configuration loaded:', config.name, `(${config.grid.rows}x${config.grid.cols})`);
    console.log('Button count:', config.buttons.length);
    
    currentConfiguration = config;
    renderButtonGrid();
    document.getElementById('config-name').textContent = config.name;
    showConnectionBanner('Configuration loaded: ' + config.name, 'connected');
    setTimeout(() => hideConnectionBanner(), 2000);
}

// Render button grid
function renderButtonGrid() {
    if (!currentConfiguration) {
        console.log('No configuration to render');
        return;
    }

    const grid = document.getElementById('button-grid');
    
    // Clear the grid
    while (grid.firstChild) {
        grid.removeChild(grid.firstChild);
    }
    
    // Force reflow
    void grid.offsetHeight;

    const { rows, cols } = currentConfiguration.grid;
    grid.style.gridTemplateColumns = `repeat(${cols}, 1fr)`;
    grid.style.gridTemplateRows = `repeat(${rows}, 1fr)`;

    console.log(`Rendering ${rows}x${cols} grid with ${currentConfiguration.buttons.length} buttons`);

    // Create all cells in grid order
    for (let row = 0; row < rows; row++) {
        for (let col = 0; col < cols; col++) {
            const button = currentConfiguration.buttons.find(b => b.row === row && b.col === col);
            
            if (button) {
                renderButton(button);
            } else {
                renderEmptyCell();
            }
        }
    }

    // Reinitialize icons
    setTimeout(() => lucide.createIcons(), 50);
}

// Render a button
function renderButton(button) {
  const grid = document.getElementById('button-grid');
  const buttonEl = document.createElement('button');
  buttonEl.className = 'deck-button';
  buttonEl.style.backgroundColor = button.color;
  buttonEl.dataset.position = `btn-${button.row}-${button.col}`;
  buttonEl.dataset.buttonId = button.id;
  buttonEl.dataset.actionType = button.action.type;
  
  // Store scene name for scene buttons
  if (button.action.type === 'switch_scene' && button.action.params?.scene_name) {
      buttonEl.dataset.sceneName = button.action.params.scene_name;
  }

  // Store params for source visibility buttons
  if ((button.action.type === 'toggle_source_visibility' ||
      button.action.type === 'show_source' ||
      button.action.type === 'hide_source') &&
      button.action.params) {
      buttonEl.dataset.sceneName = button.action.params.scene_name || '';
      buttonEl.dataset.sourceName = button.action.params.source_name || '';
  }

  buttonEl.innerHTML = `
      <i data-lucide="${button.icon || 'square'}"></i>
      <span class="button-text">${button.text}</span>
  `;

  // Check indicator
  updateButtonIndicator(buttonEl);

  // Click handler
  buttonEl.addEventListener('click', () => pressButton(`btn-${button.row}-${button.col}`, button.action));

  grid.appendChild(buttonEl);
}

// Check if button should show indicator
function shouldShowIndicator(buttonEl) {
  const actionType = buttonEl.dataset.actionType;
  const sceneName = buttonEl.dataset.sceneName;
  
  // Toggle actions
  if (isToggleAction(actionType)) {
      return isToggleActive(actionType);
  }
  
  // Scene buttons
  if (actionType === 'switch_scene' && sceneName) {
      return sceneName === obsStatus.currentScene;
  }
  
  // Source visibility buttons
  if (actionType === 'toggle_source_visibility' || 
      actionType === 'show_source' || 
      actionType === 'hide_source') {
    const buttonId = buttonEl.dataset.buttonId;
    return sourceVisibility[buttonId] === true;
  }

  // Virtual Camera indicators
  if (actionType === 'start_virtual_cam') return obsStatus.virtualCamActive;
  if (actionType === 'stop_virtual_cam') return !obsStatus.virtualCamActive;
  if (actionType === 'toggle_virtual_cam') return obsStatus.virtualCamActive;

  // Studio Mode indicators
  if (actionType === 'toggle_studio_mode' || 
      actionType === 'enable_studio_mode' || 
      actionType === 'disable_studio_mode') {
    return obsStatus.studioModeActive;
  }
  if (actionType === 'trigger_transition') return obsStatus.studioModeActive;

  return false;
}

// Check if action type is a toggle
function isToggleAction(actionType) {
  const toggleActions = [
      'toggle_stream', 'start_stream', 'stop_stream',
      'toggle_record', 'start_record', 'stop_record',
      'toggle_replay_buffer', 'start_replay_buffer', 'stop_replay_buffer',
      'toggle_virtual_cam', 'start_virtual_cam', 'stop_virtual_cam'
  ];
  return toggleActions.includes(actionType);
}

// Check if toggle is active
function isToggleActive(actionType) {
  // Start actions: show when active
  if (actionType === 'start_stream') return obsStatus.streaming;
  if (actionType === 'start_record') return obsStatus.recording;
  if (actionType === 'start_replay_buffer') return obsStatus.replayBufferActive;
  if (actionType === 'start_virtual_cam') return obsStatus.virtualCamActive;
  
  // Stop actions: show when NOT active
  if (actionType === 'stop_stream') return !obsStatus.streaming;
  if (actionType === 'stop_record') return !obsStatus.recording;
  if (actionType === 'stop_replay_buffer') return !obsStatus.replayBufferActive;
  if (actionType === 'stop_virtual_cam') return !obsStatus.virtualCamActive;
  
  // Toggle actions: show when active
  if (actionType === 'toggle_stream') return obsStatus.streaming;
  if (actionType === 'toggle_record') return obsStatus.recording;
  if (actionType === 'toggle_replay_buffer') return obsStatus.replayBufferActive;
  if (actionType === 'toggle_virtual_cam') return obsStatus.virtualCamActive;
  
  return false;
}

// Update indicator for specific button
function updateButtonIndicator(buttonEl) {
    if (shouldShowIndicator(buttonEl)) {
        buttonEl.classList.add('recording');
    } else {
        buttonEl.classList.remove('recording');
    }
}

// Update all button indicators
function updateAllIndicators() {
    const buttons = document.querySelectorAll('.deck-button');
    buttons.forEach(button => updateButtonIndicator(button));
}

// Render empty cell
function renderEmptyCell() {
    const grid = document.getElementById('button-grid');
    const emptyEl = document.createElement('div');
    emptyEl.className = 'empty-cell';
    grid.appendChild(emptyEl);
}

// Press button
async function pressButton(position, action) {
    // Visual feedback
    const button = document.querySelector(`[data-position="${position}"]`);
    if (button) {
        button.classList.add('pressed');
        setTimeout(() => button.classList.remove('pressed'), 200);
    }

    try {
        await apiClient.executeAction(action);
        
        // Update status after toggle or scene actions
        if (isToggleAction(action.type) || action.type === 'switch_scene') {
            setTimeout(() => updateStatusFromBackend(), 100);
        }

        // Update source visibility
        if (action.type === 'toggle_source_visibility' || 
            action.type === 'show_source' || 
            action.type === 'hide_source') {
          setTimeout(() => updateSourceVisibility(), 500);
        }
    } catch (err) {
        console.error('Failed to press button:', err);
        showConnectionBanner('Error: ' + err.message, 'error');
        setTimeout(() => hideConnectionBanner(), 3000);
    }
}

// Start status polling
function startStatusPolling() {
  setInterval(async () => {
      await updateStatusFromBackend();
      await updateSourceVisibility();
  }, 2000);
}

// Update status from backend
async function updateStatusFromBackend() {
  try {
      const status = await apiClient.getOBSStatus();
      
      // Track changes
      const streamingChanged = obsStatus.streaming !== (status.streaming || false);
      const recordingChanged = obsStatus.recording !== (status.recording || false);
      const sceneChanged = obsStatus.currentScene !== (status.current_scene || '');
      const virtualCamChanged = obsStatus.virtualCamActive !== (status.virtual_cam_active || false);
      const replayBufferChanged = obsStatus.replayBufferActive !== (status.replay_buffer_active || false);
      const studioModeChanged = obsStatus.studioModeActive !== (status.studio_mode_active || false);
      
      // Update state
      obsStatus.streaming = status.streaming || false;
      obsStatus.recording = status.recording || false;
      obsStatus.currentScene = status.current_scene || '';
      obsStatus.virtualCamActive = status.virtual_cam_active || false;
      obsStatus.replayBufferActive = status.replay_buffer_active || false;
      obsStatus.studioModeActive = status.studio_mode_active || false;
      
      // Update indicators if anything changed
      if (streamingChanged || recordingChanged || sceneChanged || 
          virtualCamChanged || replayBufferChanged || studioModeChanged) {
          updateAllIndicators();
      }
  } catch (err) {
      console.error('Failed to get status:', err);
  }
}

// Update source visibility for all source buttons
async function updateSourceVisibility() {
  const buttons = document.querySelectorAll('.deck-button');
  
  for (const buttonEl of buttons) {
      const actionType = buttonEl.dataset.actionType;
      
      if (actionType === 'toggle_source_visibility' || 
          actionType === 'show_source' || 
          actionType === 'hide_source') {
          
          const buttonId = buttonEl.dataset.buttonId;
          
          if (currentConfiguration && currentConfiguration.buttons) {
              const button = currentConfiguration.buttons.find(b => b.id === buttonId);
              if (button && button.action && button.action.params) {
                  const sceneName = button.action.params.scene_name;
                  const sourceName = button.action.params.source_name;
                  
                  if (sceneName && sourceName) {
                      try {
                          const visible = await apiClient.getSourceVisibility(sceneName, sourceName);
                          sourceVisibility[buttonId] = visible;
                      } catch (err) {
                          sourceVisibility[buttonId] = false;
                      }
                  }
              }
          }
      }
  }
  
  updateAllIndicators();
}

// Open settings modal
function openSettings() {
    document.getElementById('settings-modal').classList.add('open');
    setTimeout(() => lucide.createIcons(), 100);
}

// Close settings modal
function closeSettings() {
    document.getElementById('settings-modal').classList.remove('open');
}

// Update server URL
async function updateServerURL() {
    const url = document.getElementById('input-server-url').value.trim();
    
    if (!url) {
        showConnectionBanner('Please enter a server URL', 'error');
        return;
    }

    // Save URL
    localStorage.setItem('server_url', url);
    
    // Reconnect
    closeSettings();
    apiClient = new APIClient(url);
    await connectAndLoad();
}

// Open configuration selector
async function openConfigSelector() {
    try {
        const configurations = await apiClient.getConfigurations();
        renderConfigList(configurations);
        document.getElementById('config-modal').classList.add('open');
        setTimeout(() => lucide.createIcons(), 100);
    } catch (err) {
        console.error('Failed to load configurations:', err);
        showConnectionBanner('Error loading configurations: ' + err.message, 'error');
        setTimeout(() => hideConnectionBanner(), 3000);
    }
}

// Close configuration selector
function closeConfigSelector() {
    document.getElementById('config-modal').classList.remove('open');
}

// Render configuration list
function renderConfigList(configurations) {
    const list = document.getElementById('config-list');
    list.innerHTML = '';

    if (configurations.length === 0) {
        list.innerHTML = '<p class="empty-message">No configurations available</p>';
        return;
    }

    configurations.forEach(config => {
        const item = document.createElement('div');
        item.className = 'config-item';
        if (currentConfiguration && config.id === currentConfiguration.id) {
            item.classList.add('active');
        }

        const buttonCount = config.buttons ? Object.keys(config.buttons).length : 0;

        item.innerHTML = `
            <div class="config-item-header">
                <span class="config-item-name">${config.name}</span>
                ${config.is_default ? '<span class="config-badge">Default</span>' : ''}
            </div>
            <div class="config-item-description">${config.description || ''}</div>
            <div class="config-item-meta">
                <span>${config.grid.rows}×${config.grid.cols} grid</span>
                <span>•</span>
                <span>${buttonCount} buttons</span>
            </div>
        `;

        item.addEventListener('click', async () => {
            try {
                const resolved = await apiClient.getConfiguration(config.id);
                handleConfigurationLoaded(resolved);
                closeConfigSelector();
            } catch (err) {
                console.error('Failed to load configuration:', err);
                showConnectionBanner('Error loading configuration: ' + err.message, 'error');
                setTimeout(() => hideConnectionBanner(), 3000);
            }
        });

        list.appendChild(item);
    });
}

// Show connection banner
function showConnectionBanner(message, type) {
    const banner = document.getElementById('connection-banner');
    const messageEl = document.getElementById('banner-message');
    
    messageEl.textContent = message;
    banner.className = 'banner show ' + type;
}

// Hide connection banner
function hideConnectionBanner() {
    const banner = document.getElementById('connection-banner');
    banner.classList.remove('show');
}

document.addEventListener('DOMContentLoaded', async () => {
    await initializeNativeFeatures();
  });
