// Robo-Stream Client - Touchscreen Optimized

let currentConfiguration = null;
let obsStatus = {
    streaming: false,
    recording: false
};
let recordingButtons = new Set(); // Track which buttons should show recording indicator

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('Robo-Stream Client starting...');
    
    // Setup event listeners FIRST (before backend emits events)
    setupEventListeners();
    
    // Then initialize app
    initializeApp();
    
    lucide.createIcons();
});

// Initialize application
async function initializeApp() {
    try {
        // Load server URL into settings
        const serverURL = await window.go.main.App.GetServerURL();
        document.getElementById('input-server-url').value = serverURL;

        // Get current configuration from backend
        // The backend's startup() already called connectAndLoad()
        // which should have emitted configuration_loaded event
        // But we'll also explicitly load it here in case we missed the event
        const config = await window.go.main.App.GetCurrentConfiguration();
        if (config) {
            handleConfigurationLoaded(config);
        }

        // Start status polling
        startStatusPolling();
    } catch (err) {
        console.error('Initialization error:', err);
        showConnectionBanner('Failed to initialize: ' + err, 'error');
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

    // Fullscreen (top bar)
    document.getElementById('btn-fullscreen').addEventListener('click', () => {
        window.go.main.App.ToggleFullscreen();
    });

    // Listen for backend events
    window.runtime.EventsOn('connected', handleConnected);
    window.runtime.EventsOn('connection_error', handleConnectionError);
    window.runtime.EventsOn('configuration_loaded', handleConfigurationLoaded);
    window.runtime.EventsOn('config_error', handleConfigError);
}

// Handle connected event
function handleConnected(info) {
    console.log('Connected:', info);
    showConnectionBanner('Connected to server', 'connected');
    setTimeout(() => hideConnectionBanner(), 2000);
}

// Handle connection error
function handleConnectionError(error) {
    console.error('Connection error:', error);
    showConnectionBanner('Connection error: ' + error, 'error');
}

// Handle configuration loaded
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

// Handle config error
function handleConfigError(error) {
    console.error('Config error:', error);
    showConnectionBanner('Configuration error: ' + error, 'error');
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
    buttonEl.dataset.buttonId = button.id;
    buttonEl.dataset.actionType = button.action.type; // Store action type for recording detection

    buttonEl.innerHTML = `
        <i data-lucide="${button.icon || 'square'}"></i>
        <span class="button-text">${button.text}</span>
    `;

    // Check if this is a recording button and we're currently recording
    if (isRecordingAction(button.action.type) && obsStatus.recording) {
        buttonEl.classList.add('recording');
    }

    // Press by position
    buttonEl.addEventListener('click', () => pressButton(`btn-${button.row}-${button.col}`, button.action.type));

    grid.appendChild(buttonEl);
}

// Check if action type is related to recording
function isRecordingAction(actionType) {
    return actionType === 'start_record' || 
           actionType === 'stop_record' || 
           actionType === 'toggle_record';
}

// Render empty cell
function renderEmptyCell() {
    const grid = document.getElementById('button-grid');
    const emptyEl = document.createElement('div');
    emptyEl.className = 'empty-cell';
    grid.appendChild(emptyEl);
}

// Press button by position
async function pressButton(position, actionType) {
    console.log('Button pressed:', position, 'action:', actionType);

    // Visual feedback
    const button = document.querySelector(`[data-position="${position}"]`);
    if (button) {
        button.classList.add('pressed');
        setTimeout(() => button.classList.remove('pressed'), 200);
    }

    try {
        await window.go.main.App.PressButton(position);
        
        // Update recording indicators immediately after pressing recording buttons
        if (isRecordingAction(actionType)) {
            setTimeout(() => updateStatusFromBackend(), 100);
        }
    } catch (err) {
        console.error('Failed to press button:', err);
        alert('Error: ' + err);
    }
}

// Update recording indicators on all buttons
function updateRecordingIndicators() {
    const buttons = document.querySelectorAll('.deck-button');
    let recordingButtonCount = 0;
    
    buttons.forEach(button => {
        const actionType = button.dataset.actionType;
        
        if (isRecordingAction(actionType)) {
            recordingButtonCount++;
            if (obsStatus.recording) {
                console.log('Adding recording indicator to button:', actionType);
                button.classList.add('recording');
            } else {
                console.log('Removing recording indicator from button:', actionType);
                button.classList.remove('recording');
            }
        }
    });
    
    console.log(`Updated ${recordingButtonCount} recording buttons. Recording: ${obsStatus.recording}`);
}

// Start status polling
function startStatusPolling() {
    // Poll every 2 seconds
    setInterval(async () => {
        await updateStatusFromBackend();
    }, 2000);
}

// Update status from backend
async function updateStatusFromBackend() {
    try {
        const status = await window.go.main.App.GetOBSStatus();
        
        // Track if recording state changed
        const wasRecording = obsStatus.recording;
        
        obsStatus.streaming = status.streaming || false;
        obsStatus.recording = status.recording || false;
        
        // Debug log
        console.log('Status update - Streaming:', obsStatus.streaming, 'Recording:', obsStatus.recording);
        
        // If recording state changed, update button indicators
        if (wasRecording !== obsStatus.recording) {
            console.log('Recording state changed:', obsStatus.recording);
            updateRecordingIndicators();
        }
    } catch (err) {
        console.error('Failed to get status:', err);
    }
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
        alert('Please enter a server URL');
        return;
    }

    try {
        await window.go.main.App.SetServerURL(url);
        closeSettings();
        showConnectionBanner('Connecting to ' + url + '...', 'connecting');
        
        // Reload configuration after URL change
        setTimeout(() => loadConfiguration(), 1000);
    } catch (err) {
        console.error('Failed to update server URL:', err);
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
