// Robo-Stream Web Client - Touchscreen Optimized
import { APIClient } from './api.js';
import { initializeNativeFeatures, hapticFeedback, isNativeApp, getAppInfo } from './native.js';
import { createIcons, icons } from 'lucide';

let currentConfiguration = null;
let apiClient = null;

// Indicator state fetched from the server — keyed by button position ID (e.g. "btn-0-0").
// The server computes all indicator logic; the client is fully plugin-agnostic.
// Values: "active" | "warn" | ""
let indicatorState = {};

// Optimistic overrides applied immediately on button press so indicators respond
// without waiting for the next poll cycle.  Each entry is cleared once the server
// confirms the new state (or after a short timeout).
// Values: "active" | "warn" | ""
let indicatorOverrides = {};
const buttonPressCooldownMs = 350;
const lastButtonPressAt = {};

const ZOOMOSC_ACTION_TYPES = new Set([
    'toggle_audio', 'mute_audio', 'unmute_audio',
    'toggle_video', 'start_video', 'stop_video',
    'toggle_share', 'start_share', 'stop_share',
    'raise_hand', 'lower_hand', 'toggle_hand',
    'leave_meeting', 'end_meeting',
    'spotlight_self', 'unspotlight_self'
]);

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
    // Clear stale indicator state whenever the configuration changes.
    indicatorState = {};
    indicatorOverrides = {};
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

  buttonEl.innerHTML = `
      <i data-lucide="${button.icon || 'square'}"></i>
      <span class="button-text">${button.text}</span>
  `;

  // Apply current indicator state
  updateButtonIndicator(buttonEl);

  // Click handler — pass button.id (UUID) so the override map is keyed the
  // same way as indicatorState (which the server keys by button UUID).
  buttonEl.addEventListener('click', () => pressButton(`btn-${button.row}-${button.col}`, button.id, button.action));

  grid.appendChild(buttonEl);
}

// ──────────────────────────────────────────────────────────────────────────────
// Indicator helpers
// ──────────────────────────────────────────────────────────────────────────────

// Returns the effective indicator class for a button, respecting any active
// optimistic override that was set when the button was pressed.
function getIndicatorClass(buttonId) {
    if (buttonId in indicatorOverrides) {
        return indicatorOverrides[buttonId];
    }
    return indicatorState[buttonId] || '';
}

// Apply the correct 'recording' CSS class to a single button element.
function updateButtonIndicator(buttonEl) {
    const cls = getIndicatorClass(buttonEl.dataset.buttonId);
    if (cls) {
        buttonEl.classList.add('recording');
    } else {
        buttonEl.classList.remove('recording');
    }
}

// Re-evaluate all buttons in the grid.
function updateAllIndicators() {
    document.querySelectorAll('.deck-button').forEach(btn => updateButtonIndicator(btn));
}

// Fetch server-computed indicator states and update the UI.
async function updateButtonIndicators() {
    if (!apiClient) return;
    try {
        const states = await apiClient.getButtonIndicators();
        if (!states || typeof states !== 'object') return;

        // Only repaint if something actually changed.
        let changed = false;
        for (const [id, cls] of Object.entries(states)) {
            if (indicatorState[id] !== cls) {
                changed = true;
                break;
            }
        }

        indicatorState = states;
        if (changed) updateAllIndicators();
    } catch (err) {
        // Silently ignore — server may be temporarily unreachable.
    }
}

// ──────────────────────────────────────────────────────────────────────────────
// Button press — optimistic indicator updates
// ──────────────────────────────────────────────────────────────────────────────

// Render empty cell
function renderEmptyCell() {
    const grid = document.getElementById('button-grid');
    const emptyEl = document.createElement('div');
    emptyEl.className = 'empty-cell';
    grid.appendChild(emptyEl);
}

// Press button
// - position: "btn-{row}-{col}" — used only for visual pressed class
// - buttonId: button UUID (matches indicatorState and indicatorOverrides keys)
// - action: the action descriptor sent to the server
async function pressButton(position, buttonId, action) {
    const now = Date.now();
    if (lastButtonPressAt[buttonId] && now-lastButtonPressAt[buttonId] < buttonPressCooldownMs) {
        return;
    }
    lastButtonPressAt[buttonId] = now;

    // Visual feedback
    const buttonEl = document.querySelector(`[data-position="${position}"]`);
    if (buttonEl) {
        buttonEl.classList.add('pressed');
        setTimeout(() => buttonEl.classList.remove('pressed'), 200);
    }

    try {
        await apiClient.executeAction(action);

        // Apply an optimistic indicator override so the UI responds immediately
        // without waiting for the next server poll.
        //
        // Convention:
        //   • toggle_* actions: flip the current indicator state.
        //   • All other actions: assume "active" (the intended state is now in
        //     effect).  This is correct for start/stop/mute/unmute pairs because
        //     the server uses "active" to mean "this button's target state IS
        //     currently the case" — e.g. stop_stream is active when NOT streaming.
        //
        // The override key uses buttonId (UUID) — the same key the server uses
        // in the button-indicators response so getIndicatorClass() finds it.
        if (buttonEl) {
            const actionType = action.type || '';
            const preClickState = indicatorState[buttonId] || '';
            let newClass;
            if (actionType.startsWith('toggle_')) {
                // Toggle: flip whatever the server last reported for this button.
                newClass = preClickState === 'active' ? '' : 'active';
            } else {
                newClass = 'active';
            }
            indicatorOverrides[buttonId] = newClass;
            updateAllIndicators();

            // Reconcile loop: poll the server every 500 ms until it confirms
            // the state has changed. For ZoomOSC we keep the optimistic override
            // while the plugin still reports unknown state.
            let attempts = 0;
            const controller = action.controller || (ZOOMOSC_ACTION_TYPES.has(actionType) ? 'zoomosc' : 'obs');
            const reconcile = async () => {
                attempts++;
                await updateButtonIndicators();
                const serverState = indicatorState[buttonId] || '';
                if (serverState !== preClickState) {
                    // Server confirmed the change (or we timed out) — hand off
                    // to the authoritative server state.
                    delete indicatorOverrides[buttonId];
                    updateAllIndicators();
                    return;
                }

                // Keep optimistic state while ZoomOSC status is still unknown.
                if (controller === 'zoomosc') {
                    const status = await apiClient.getControllerStatus('zoomosc');
                    if (status && status.state_known === false) {
                        setTimeout(reconcile, 500);
                        return;
                    }
                }

                if (attempts >= 5) {
                    delete indicatorOverrides[buttonId];
                    updateAllIndicators();
                } else {
                    setTimeout(reconcile, 500);
                }
            };
            setTimeout(reconcile, 500);
        }

    } catch (err) {
        console.error('Failed to press button:', err);
        showConnectionBanner('Error: ' + err.message, 'error');
        setTimeout(() => hideConnectionBanner(), 3000);
    }
}

// Start status polling — a single endpoint handles all controllers.
function startStatusPolling() {
    // Immediate fetch so indicators appear before the first interval tick.
    updateButtonIndicators();

    setInterval(() => {
        updateButtonIndicators();
    }, 2000);
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

    // Show version in title bar
    try {
        const info = await getAppInfo();
        if (info) {
            const titleVersion = document.getElementById('title-version');
            if (titleVersion) titleVersion.textContent = `v${info.version}`;
        }
    } catch (e) { /* ignore */ }
  });
