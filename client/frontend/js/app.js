// Robo-Stream Client Application (for YOUR server)

let currentConfiguration = null;

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('Robo-Stream Client starting...');
    initializeApp();
    setupEventListeners();
    lucide.createIcons();
});

// Initialize application
async function initializeApp() {
    // Load server URL
    const serverURL = await window.go.main.App.GetServerURL();
    document.getElementById('server-url').value = serverURL;

    // Load current configuration
    await loadConfiguration();

    // Start status polling
    startStatusPolling();
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

    // Fullscreen
    document.getElementById('btn-fullscreen').addEventListener('click', () => {
        window.go.main.App.ToggleFullscreen();
    });

    // Reconnect
    document.getElementById('btn-reconnect').addEventListener('click', async () => {
        showConnectionBanner('Reconnecting...', 'connecting');
        await window.go.main.App.Reconnect();
    });

    // Listen for backend events
    window.runtime.EventsOn('connected', handleConnected);
    window.runtime.EventsOn('connection_error', handleConnectionError);
    window.runtime.EventsOn('configuration_loaded', handleConfigurationLoaded);
    window.runtime.EventsOn('status_update', handleStatusUpdate);
}

// Load current configuration
async function loadConfiguration() {
    try {
        currentConfiguration = await window.go.main.App.GetConfiguration();
        if (currentConfiguration) {
            renderButtonGrid();
            document.getElementById('config-name').textContent = currentConfiguration.name;
        }
    } catch (err) {
        console.error('Failed to load configuration:', err);
    }
}

// Render button grid
function renderButtonGrid() {
    if (!currentConfiguration) {
        console.log('No configuration to render');
        return;
    }

    const grid = document.getElementById('button-grid');
    
    // Thoroughly clear the grid
    while (grid.firstChild) {
        grid.removeChild(grid.firstChild);
    }
    
    // Force a reflow to ensure DOM is cleared
    void grid.offsetHeight;

    const { rows, cols } = currentConfiguration.grid;
    grid.style.gridTemplateColumns = `repeat(${cols}, 1fr)`;
    grid.style.gridTemplateRows = `repeat(${rows}, 1fr)`;

    console.log(`Rendering ${rows}x${cols} grid with ${currentConfiguration.buttons.length} buttons`);

    // Create all cells in grid order
    for (let row = 0; row < rows; row++) {
        for (let col = 0; col < cols; col++) {
            // Find button at this position
            const button = currentConfiguration.buttons.find(b => b.row === row && b.col === col);
            
            if (button) {
                renderButton(button);
            } else {
                renderEmptyCell();
            }
        }
    }

    // Reinitialize icons after a brief delay to ensure DOM is ready
    setTimeout(() => lucide.createIcons(), 50);
}

// Render a button
function renderButton(button) {
    const grid = document.getElementById('button-grid');
    const buttonEl = document.createElement('button');
    buttonEl.className = 'deck-button';
    buttonEl.style.backgroundColor = button.color;
    buttonEl.dataset.position = `btn-${button.row}-${button.col}`;
    buttonEl.dataset.buttonId = button.id; // Add unique ID for tracking

    buttonEl.innerHTML = `
        <i data-lucide="${button.icon || 'square'}"></i>
        <span class="button-text">${button.text}</span>
    `;

    // Press by position (not ID)
    buttonEl.addEventListener('click', () => pressButton(`btn-${button.row}-${button.col}`));

    grid.appendChild(buttonEl);
}

// Render empty cell
function renderEmptyCell() {
    const grid = document.getElementById('button-grid');
    const cell = document.createElement('div');
    cell.className = 'empty-cell';
    cell.textContent = '';
    grid.appendChild(cell);
}

// Press button by position
async function pressButton(position) {
    console.log('Button pressed:', position);

    // Visual feedback
    const button = document.querySelector(`[data-position="${position}"]`);
    if (button) {
        button.classList.add('pressed');
        setTimeout(() => button.classList.remove('pressed'), 200);
    }

    try {
        await window.go.main.App.PressButton(position);
    } catch (err) {
        console.error('Failed to press button:', err);
        alert('Error: ' + err);
    }
}

// Open configuration selector
async function openConfigSelector() {
    try {
        const configurations = await window.go.main.App.GetConfigurations();
        renderConfigList(configurations);
        document.getElementById('config-modal').classList.add('open');
        setTimeout(() => lucide.createIcons(), 100);
    } catch (err) {
        console.error('Failed to load configurations:', err);
        alert('Error loading configurations: ' + err);
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

        // Count buttons from the map
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
                await window.go.main.App.LoadConfiguration(config.id);
                closeConfigSelector();
            } catch (err) {
                console.error('Failed to load configuration:', err);
                alert('Error loading configuration: ' + err);
            }
        });

        list.appendChild(item);
    });
}

// Open settings
function openSettings() {
    document.getElementById('settings-modal').classList.add('open');
    setTimeout(() => lucide.createIcons(), 100);
}

// Close settings
function closeSettings() {
    document.getElementById('settings-modal').classList.remove('open');
}

// Update server URL
async function updateServerURL() {
    const url = document.getElementById('server-url').value.trim();
    if (!url) {
        alert('Please enter a server URL');
        return;
    }

    try {
        showConnectionBanner('Connecting to new server...', 'connecting');
        await window.go.main.App.SetServerURL(url);
        closeSettings();
    } catch (err) {
        console.error('Failed to update server URL:', err);
        showConnectionBanner('Connection failed', 'error');
        alert('Error: ' + err);
    }
}

// Event handlers
function handleConnected(info) {
    console.log('Connected to server:', info);
    showConnectionBanner('Connected', 'connected');
    setTimeout(() => hideConnectionBanner(), 2000);
}

function handleConnectionError(error) {
    console.error('Connection error:', error);
    showConnectionBanner('Connection failed: ' + error, 'error');
}

function handleConfigurationLoaded(config) {
    console.log('Configuration loaded:', config.name, `(${config.grid.rows}x${config.grid.cols})`);
    console.log('Button count:', config.buttons.length);
    console.log('Buttons:', config.buttons.map(b => `${b.text} at (${b.row},${b.col})`));
    
    currentConfiguration = config;
    renderButtonGrid();
    document.getElementById('config-name').textContent = config.name;
    showConnectionBanner('Configuration loaded: ' + config.name, 'connected');
    setTimeout(() => hideConnectionBanner(), 2000);
}

function handleStatusUpdate(status) {
    // Update streaming indicator
    const streamIndicator = document.getElementById('stream-indicator');
    if (status.streaming) {
        streamIndicator.classList.add('active');
    } else {
        streamIndicator.classList.remove('active');
    }

    // Update recording indicator
    const recordIndicator = document.getElementById('record-indicator');
    if (status.recording) {
        recordIndicator.classList.add('active');
    } else {
        recordIndicator.classList.remove('active');
    }

    // Update current scene
    const currentSceneEl = document.getElementById('current-scene');
    if (status.current_scene) {
        currentSceneEl.textContent = status.current_scene;
    }
}

// Connection banner
function showConnectionBanner(message, type) {
    const banner = document.getElementById('connection-banner');
    banner.querySelector('.connection-status').textContent = message;
    banner.className = 'connection-banner ' + type;
    banner.style.display = 'flex';
}

function hideConnectionBanner() {
    document.getElementById('connection-banner').style.display = 'none';
}

// Status polling
function startStatusPolling() {
    // Poll status every 2 seconds
    setInterval(async () => {
        try {
            const status = await window.go.main.App.GetStatus();
            handleStatusUpdate(status);
        } catch (err) {
            // Silently fail - connection might be down
        }
    }, 2000);
}
