// Robo-Stream Web Client - Touchscreen Optimized
import { APIClient } from './api.js';
import { initializeNativeFeatures, hapticFeedback, isNativeApp, getAppInfo } from './native.js';
import { createIcons, icons } from 'lucide';

let currentConfiguration = null;
let apiClient = null;
let obsStatus = {
  streaming: false,
  recording: false,
  recordingPaused: false,
  currentScene: '',
  virtualCamActive: false,
  replayBufferActive: false,
  studioModeActive: false
};
let sourceVisibility = {};
let inputMuted = {};
let filterEnabled = {};

// Config caches — avoid redundant fetches
let configListCache = null;        // lightweight summaries for the selector
let resolvedConfigCache = {};      // configId → full resolved config

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('Robo-Stream Web Client starting...');
    
    // Setup event listeners
    setupEventListeners();
    
    // Initialize app
    initializeApp();
    
    // Initialize Lucide icons
    createIcons({ icons });
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
        showConnectionBanner('Connecting to server...', 'connecting');

        // Register fetches config in one round trip (replaces separate health check)
        console.log('Registering with server...');
        const config = await apiClient.register();
        showConnectionBanner('Connected to server', 'connected');

        // Cache this config so switching back to it is instant
        if (config && config.id) {
            resolvedConfigCache[config.id] = config;
        }

        // Pre-fetch the config list in the background so the selector opens instantly
        apiClient.getConfigurations()
            .then(list => { configListCache = list; })
            .catch(() => {});

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

    // Build a position-indexed map so each cell lookup is O(1) instead of O(N)
    const buttonMap = new Map(
        currentConfiguration.buttons.map(b => [`${b.row}-${b.col}`, b])
    );

    // Create all cells in grid order
    for (let row = 0; row < rows; row++) {
        for (let col = 0; col < cols; col++) {
            const button = buttonMap.get(`${row}-${col}`);

            if (button) {
                renderButton(button);
            } else {
                renderEmptyCell();
            }
        }
    }

    // Reinitialize icons
    setTimeout(() => createIcons({ icons }), 50);
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

  // Store params for input mute buttons
  if ((button.action.type === 'toggle_input_mute' ||
      button.action.type === 'mute_input' ||
      button.action.type === 'unmute_input') &&
      button.action.params) {
      buttonEl.dataset.inputName = button.action.params.input_name || '';
  }

  // Store params for source filter buttons
  if ((button.action.type === 'toggle_source_filter' ||
      button.action.type === 'enable_source_filter' ||
      button.action.type === 'disable_source_filter') &&
      button.action.params) {
      buttonEl.dataset.sourceName = button.action.params.source_name || '';
      buttonEl.dataset.filterName = button.action.params.filter_name || '';
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
  if (actionType === 'toggle_studio_mode' || actionType === 'enable_studio_mode') {
    return obsStatus.studioModeActive;
  }
  if (actionType === 'disable_studio_mode') return !obsStatus.studioModeActive;
  if (actionType === 'trigger_transition') return obsStatus.studioModeActive;

  // Recording pause/resume indicators
  if (actionType === 'pause_record') return obsStatus.recording && !obsStatus.recordingPaused;
  if (actionType === 'resume_record') return obsStatus.recording && obsStatus.recordingPaused;

  // Input mute indicators
  if (actionType === 'toggle_input_mute' ||
      actionType === 'mute_input' ||
      actionType === 'unmute_input') {
    const buttonId = buttonEl.dataset.buttonId;
    const muted = inputMuted[buttonId] === true;
    if (actionType === 'unmute_input') return !muted;
    return muted;
  }

  // Source filter indicators
  if (actionType === 'toggle_source_filter' ||
      actionType === 'enable_source_filter' ||
      actionType === 'disable_source_filter') {
    const buttonId = buttonEl.dataset.buttonId;
    const enabled = filterEnabled[buttonId] === true;
    if (actionType === 'disable_source_filter') return !enabled;
    return enabled;
  }

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

        // Update input mute state
        if (action.type === 'toggle_input_mute' ||
            action.type === 'mute_input' ||
            action.type === 'unmute_input') {
          setTimeout(() => updateInputMute(), 500);
        }

        // Update source filter state
        if (action.type === 'toggle_source_filter' ||
            action.type === 'enable_source_filter' ||
            action.type === 'disable_source_filter') {
          setTimeout(() => updateFilterState(), 500);
        }

        // Update recording status for pause/resume
        if (action.type === 'pause_record' || action.type === 'resume_record') {
          setTimeout(() => updateStatusFromBackend(), 500);
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
      // Run all four updates in parallel - each is independent
      await Promise.all([
          updateStatusFromBackend(),
          updateSourceVisibility(),
          updateInputMute(),
          updateFilterState(),
      ]);
  }, 2000);
}

// Update status from backend
async function updateStatusFromBackend() {
  try {
      const status = await apiClient.getOBSStatus();
      
      // Track changes
      const streamingChanged = obsStatus.streaming !== (status.streaming || false);
      const recordingChanged = obsStatus.recording !== (status.recording || false);
      const recordingPausedChanged = obsStatus.recordingPaused !== (status.recording_paused || false);
      const sceneChanged = obsStatus.currentScene !== (status.current_scene || '');
      const virtualCamChanged = obsStatus.virtualCamActive !== (status.virtual_cam_active || false);
      const replayBufferChanged = obsStatus.replayBufferActive !== (status.replay_buffer_active || false);
      const studioModeChanged = obsStatus.studioModeActive !== (status.studio_mode_active || false);

      // Update state
      obsStatus.streaming = status.streaming || false;
      obsStatus.recording = status.recording || false;
      obsStatus.recordingPaused = status.recording_paused || false;
      obsStatus.currentScene = status.current_scene || '';
      obsStatus.virtualCamActive = status.virtual_cam_active || false;
      obsStatus.replayBufferActive = status.replay_buffer_active || false;
      obsStatus.studioModeActive = status.studio_mode_active || false;

      // Update indicators if anything changed
      if (streamingChanged || recordingChanged || recordingPausedChanged || sceneChanged ||
          virtualCamChanged || replayBufferChanged || studioModeChanged) {
          updateAllIndicators();
      }
  } catch (err) {
      console.error('Failed to get status:', err);
  }
}

// Update source visibility for all source buttons
async function updateSourceVisibility() {
  if (!currentConfiguration) return;

  // Build a buttonId→params map once so the loop below is O(1) per button
  const paramsByButtonId = new Map(
      (currentConfiguration.buttons || [])
          .filter(b => b.action && b.action.params)
          .map(b => [b.id, b.action.params])
  );

  const buttons = document.querySelectorAll('.deck-button');

  // Deduplicate: collect unique (scene, source) pairs to avoid querying OBS
  // multiple times for buttons that reference the same source in the same scene
  const keyToButtonIds = new Map(); // "scene||source" → { sceneName, sourceName, buttonIds[] }

  for (const buttonEl of buttons) {
      const actionType = buttonEl.dataset.actionType;

      if (actionType === 'toggle_source_visibility' ||
          actionType === 'show_source' ||
          actionType === 'hide_source') {

          const buttonId = buttonEl.dataset.buttonId;
          const params = paramsByButtonId.get(buttonId);
          const sceneName = params?.scene_name;
          const sourceName = params?.source_name;

          if (sceneName && sourceName) {
              const key = `${sceneName}||${sourceName}`;
              if (!keyToButtonIds.has(key)) {
                  keyToButtonIds.set(key, { sceneName, sourceName, buttonIds: [] });
              }
              keyToButtonIds.get(key).buttonIds.push(buttonId);
          }
      }
  }

  // One API call per unique (scene, source) pair
  const promises = [];
  for (const { sceneName, sourceName, buttonIds } of keyToButtonIds.values()) {
      promises.push(
          apiClient.getSourceVisibility(sceneName, sourceName)
              .then(visible => { for (const id of buttonIds) sourceVisibility[id] = visible; })
              .catch(() => { for (const id of buttonIds) sourceVisibility[id] = false; })
      );
  }

  await Promise.all(promises);
  updateAllIndicators();
}

// Update mute state for all input mute buttons
async function updateInputMute() {
  const buttons = document.querySelectorAll('.deck-button');

  // Deduplicate: one OBS call per unique input name, applied to all buttons
  // that reference that input (mute/unmute/toggle buttons for the same mic)
  const inputToButtonIds = new Map(); // inputName → buttonId[]

  for (const buttonEl of buttons) {
    const actionType = buttonEl.dataset.actionType;

    if (actionType === 'toggle_input_mute' ||
        actionType === 'mute_input' ||
        actionType === 'unmute_input') {
      const inputName = buttonEl.dataset.inputName;
      const buttonId = buttonEl.dataset.buttonId;

      if (inputName) {
        if (!inputToButtonIds.has(inputName)) {
          inputToButtonIds.set(inputName, []);
        }
        inputToButtonIds.get(inputName).push(buttonId);
      }
    }
  }

  // One API call per unique input name
  const promises = [];
  for (const [inputName, buttonIds] of inputToButtonIds) {
    promises.push(
        apiClient.getInputMute(inputName)
            .then(muted => { for (const id of buttonIds) inputMuted[id] = muted; })
            .catch(() => { for (const id of buttonIds) inputMuted[id] = false; })
    );
  }

  await Promise.all(promises);
  updateAllIndicators();
}

// Update filter state for all source filter buttons
async function updateFilterState() {
  const buttons = document.querySelectorAll('.deck-button');

  // Deduplicate: one OBS call per unique (source, filter) pair
  const keyToButtonIds = new Map(); // "source||filter" → { sourceName, filterName, buttonIds[] }

  for (const buttonEl of buttons) {
    const actionType = buttonEl.dataset.actionType;

    if (actionType === 'toggle_source_filter' ||
        actionType === 'enable_source_filter' ||
        actionType === 'disable_source_filter') {
      const sourceName = buttonEl.dataset.sourceName;
      const filterName = buttonEl.dataset.filterName;
      const buttonId = buttonEl.dataset.buttonId;

      if (sourceName && filterName) {
        const key = `${sourceName}||${filterName}`;
        if (!keyToButtonIds.has(key)) {
          keyToButtonIds.set(key, { sourceName, filterName, buttonIds: [] });
        }
        keyToButtonIds.get(key).buttonIds.push(buttonId);
      }
    }
  }

  // One API call per unique (source, filter) pair
  const promises = [];
  for (const { sourceName, filterName, buttonIds } of keyToButtonIds.values()) {
    promises.push(
        apiClient.getSourceFilterEnabled(sourceName, filterName)
            .then(enabled => { for (const id of buttonIds) filterEnabled[id] = enabled; })
            .catch(() => { for (const id of buttonIds) filterEnabled[id] = false; })
    );
  }

  await Promise.all(promises);
  updateAllIndicators();
}

// Open settings modal
async function openSettings() {
    document.getElementById('settings-modal').classList.add('open');
    setTimeout(() => createIcons({ icons }), 100);

    // Populate version
    try {
        const info = await getAppInfo();
        const versionEl = document.getElementById('app-version');
        if (versionEl && info) {
            const build = info.build && info.build !== 'web' ? ` (${info.build})` : '';
            versionEl.textContent = `v${info.version}${build}`;
        }
    } catch (e) { /* ignore */ }
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
        // Use cached list if available (pre-fetched at startup); otherwise fetch now
        const configurations = configListCache || await apiClient.getConfigurations();
        renderConfigList(configurations);
        document.getElementById('config-modal').classList.add('open');
        setTimeout(() => createIcons({ icons }), 100);
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

        // button_count is set by the new summary endpoint;
        // fall back to counting the raw buttons map for old server versions
        const buttonCount = config.button_count ?? (config.buttons ? Object.keys(config.buttons).length : 0);

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
                // Use cached resolved config if available; otherwise fetch and cache it
                let resolved = resolvedConfigCache[config.id];
                if (!resolved) {
                    resolved = await apiClient.getConfiguration(config.id);
                    resolvedConfigCache[config.id] = resolved;
                }
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
