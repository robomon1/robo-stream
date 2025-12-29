// Stream-Pi Deck Client

class StreamPiDeck {
    constructor() {
        this.config = null;
        this.currentView = 'deck';  // 'deck' or 'config'
        this.ws = null;
        this.currentButton = null;
        this.scenes = [];
        this.inputs = [];

        this.init();
    }

    async init() {
        await this.loadConfig();
        this.renderDeck();
        this.setupEventListeners();
        this.connectWebSocket();
        await this.loadScenes();
        await this.loadInputs();
        await this.updateStatus();
        
        // Start status polling
        setInterval(() => this.updateStatus(), 5000);
    }

    async loadConfig() {
        try {
            const response = await fetch('/api/buttons');
            this.config = await response.json();
        } catch (error) {
            console.error('Failed to load config:', error);
        }
    }

    async loadScenes() {
        try {
            const response = await fetch('/api/scenes');
            const data = await response.json();
            this.scenes = data.scenes || [];
            this.updateSceneDropdown();
        } catch (error) {
            console.error('Failed to load scenes:', error);
        }
    }

    async loadInputs() {
        try {
            const response = await fetch('/api/inputs');
            const data = await response.json();
            this.inputs = data.inputs || [];
            this.updateInputDropdown();
        } catch (error) {
            console.error('Failed to load inputs:', error);
        }
    }

    updateSceneDropdown() {
        const select = document.getElementById('scene-name');
        select.innerHTML = '<option value="">-- Select Scene --</option>';
        this.scenes.forEach(scene => {
            const option = document.createElement('option');
            option.value = scene;
            option.textContent = scene;
            select.appendChild(option);
        });
    }

    updateInputDropdown() {
        const select = document.getElementById('input-name');
        select.innerHTML = '<option value="">-- Select Input --</option>';
        this.inputs.forEach(input => {
            const option = document.createElement('option');
            option.value = input;
            option.textContent = input;
            select.appendChild(option);
        });
    }

    renderDeck() {
        const grid = document.getElementById('button-grid');
        grid.innerHTML = '';
        grid.style.gridTemplateColumns = `repeat(${this.config.grid.cols}, 80px)`;

        // Create all grid positions
        for (let row = 0; row < this.config.grid.rows; row++) {
            for (let col = 0; col < this.config.grid.cols; col++) {
                const buttonId = `btn-${row}-${col}`;
                const button = this.config.buttons.find(b => b.id === buttonId);

                const btnElement = document.createElement('button');
                btnElement.className = 'deck-button';
                btnElement.dataset.id = buttonId;

                if (button) {
                    btnElement.style.background = button.color;
                    const textDiv = document.createElement('div');
                    textDiv.className = 'text';
                    textDiv.textContent = button.text;
                    btnElement.appendChild(textDiv);
                } else {
                    btnElement.classList.add('empty');
                }

                btnElement.addEventListener('click', () => this.handleButtonClick(buttonId));
                grid.appendChild(btnElement);
            }
        }
    }

    renderConfigView() {
        // Update grid inputs
        document.getElementById('grid-rows').value = this.config.grid.rows;
        document.getElementById('grid-cols').value = this.config.grid.cols;

        // Render button list
        const buttonList = document.getElementById('button-list');
        const emptyState = document.getElementById('button-empty-state');
        
        buttonList.innerHTML = '';

        if (this.config.buttons.length === 0) {
            buttonList.style.display = 'none';
            emptyState.style.display = 'block';
            return;
        }

        buttonList.style.display = 'grid';
        emptyState.style.display = 'none';

        this.config.buttons.forEach(button => {
            const card = document.createElement('div');
            card.className = 'button-card';
            card.onclick = () => this.showConfigModal(button.id);

            card.innerHTML = `
                <div class="button-card-header">
                    <div class="button-card-id">${button.id}</div>
                    <div class="button-card-color" style="background-color: ${button.color}"></div>
                </div>
                <div class="button-card-title">${button.text}</div>
                <div class="button-card-action">${this.formatActionName(button.action.type)}</div>
            `;

            buttonList.appendChild(card);
        });
    }

    formatActionName(actionType) {
        return actionType
            .split('_')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');
    }

    handleButtonClick(buttonId) {
        const button = this.config.buttons.find(b => b.id === buttonId);
        if (!button) return;

        this.pressButton(buttonId);
    }

    async pressButton(buttonId) {
        try {
            const response = await fetch(`/api/buttons/${buttonId}/press`, {
                method: 'POST'
            });
            const result = await response.json();
            if (!result.success) {
                console.error('Action failed:', result.message);
            }
        } catch (error) {
            console.error('Failed to press button:', error);
        }
    }

    showConfigModal(buttonId) {
        this.currentButton = buttonId;
        const button = this.config.buttons.find(b => b.id === buttonId);

        const modal = document.getElementById('config-modal');
        const form = document.getElementById('button-config-form');
        const deleteBtn = document.getElementById('delete-btn');

        // Reset form
        form.reset();

        if (button) {
            document.getElementById('button-text').value = button.text;
            document.getElementById('button-color').value = button.color;
            document.getElementById('action-type').value = button.action.type;
            this.updateActionParams(button.action.type, button.action.params);
            deleteBtn.style.display = 'block';
        } else {
            deleteBtn.style.display = 'none';
        }

        modal.classList.add('show');
    }

    hideConfigModal() {
        const modal = document.getElementById('config-modal');
        modal.classList.remove('show');
        this.currentButton = null;
    }

    updateActionParams(actionType, params = {}) {
        // Hide all param fields
        document.getElementById('scene-param').style.display = 'none';
        document.getElementById('input-param').style.display = 'none';
        document.getElementById('volume-param').style.display = 'none';
        document.getElementById('source-param').style.display = 'none';
        document.getElementById('visibility-param').style.display = 'none';
        document.getElementById('screenshot-param').style.display = 'none';

        // Show relevant param fields
        if (actionType === 'switch_scene') {
            document.getElementById('scene-param').style.display = 'block';
            if (params.scene_name) {
                document.getElementById('scene-name').value = params.scene_name;
            }
        } else if (actionType.includes('input') || actionType.includes('mute')) {
            document.getElementById('input-param').style.display = 'block';
            if (params.input_name) {
                document.getElementById('input-name').value = params.input_name;
            }
        } else if (actionType === 'set_input_volume') {
            document.getElementById('input-param').style.display = 'block';
            document.getElementById('volume-param').style.display = 'block';
            if (params.input_name) {
                document.getElementById('input-name').value = params.input_name;
            }
            if (params.volume_db !== undefined) {
                document.getElementById('volume-db').value = params.volume_db;
            }
        } else if (actionType === 'set_source_visibility') {
            document.getElementById('source-param').style.display = 'block';
            document.getElementById('visibility-param').style.display = 'block';
            if (params.source_name) {
                document.getElementById('source-name').value = params.source_name;
            }
            if (params.visible !== undefined) {
                document.getElementById('visibility').value = params.visible.toString();
            }
        } else if (actionType === 'take_screenshot') {
            document.getElementById('source-param').style.display = 'block';
            document.getElementById('screenshot-param').style.display = 'block';
            if (params.source_name) {
                document.getElementById('source-name').value = params.source_name;
            }
            if (params.file_path) {
                document.getElementById('file-path').value = params.file_path;
            }
        }
    }

    async saveButton() {
        const form = document.getElementById('button-config-form');
        const formData = new FormData(form);

        const text = formData.get('text');
        const color = formData.get('color');
        const actionType = formData.get('action_type');

        if (!text || !actionType) {
            alert('Please fill in all required fields');
            return;
        }

        const params = {};
        if (actionType === 'switch_scene') {
            params.scene_name = formData.get('scene_name');
        } else if (actionType.includes('input') || actionType.includes('mute')) {
            params.input_name = formData.get('input_name');
        } else if (actionType === 'set_input_volume') {
            params.input_name = formData.get('input_name');
            params.volume_db = parseFloat(formData.get('volume_db'));
        } else if (actionType === 'set_source_visibility') {
            params.source_name = formData.get('source_name');
            params.visible = formData.get('visible') === 'true';
        } else if (actionType === 'take_screenshot') {
            params.source_name = formData.get('source_name');
            params.file_path = formData.get('file_path');
        }

        const [, row, col] = this.currentButton.split('-').map(Number);

        const button = {
            id: this.currentButton,
            row: row,
            col: col,
            text: text,
            color: color,
            action: {
                type: actionType,
                params: params
            }
        };

        try {
            const response = await fetch(`/api/buttons/${this.currentButton}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(button)
            });

            if (response.ok) {
                await this.loadConfig();
                this.renderDeck();
                this.renderConfigView();
                this.hideConfigModal();
            }
        } catch (error) {
            console.error('Failed to save button:', error);
        }
    }

    async deleteButton() {
        if (!confirm('Are you sure you want to delete this button?')) {
            return;
        }

        try {
            const response = await fetch(`/api/buttons/${this.currentButton}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                await this.loadConfig();
                this.renderDeck();
                this.renderConfigView();
                this.hideConfigModal();
            }
        } catch (error) {
            console.error('Failed to delete button:', error);
        }
    }

    async updateGrid() {
        const rows = parseInt(document.getElementById('grid-rows').value);
        const cols = parseInt(document.getElementById('grid-cols').value);

        if (rows < 1 || rows > 10 || cols < 1 || cols > 10) {
            alert('Rows and columns must be between 1 and 10');
            return;
        }

        this.config.grid.rows = rows;
        this.config.grid.cols = cols;

        try {
            const response = await fetch('/api/buttons', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.config)
            });

            if (response.ok) {
                await this.loadConfig();
                this.renderDeck();
                this.renderConfigView();
            }
        } catch (error) {
            console.error('Failed to update grid:', error);
        }
    }

    switchView(view) {
        const deckView = document.getElementById('deck-view');
        const configView = document.getElementById('config-view');

        if (view === 'config') {
            deckView.classList.add('hidden');
            configView.classList.add('active');
            this.renderConfigView();
            this.currentView = 'config';
        } else {
            deckView.classList.remove('hidden');
            configView.classList.remove('active');
            this.currentView = 'deck';
        }
    }

    setupEventListeners() {
        // View switching
        document.getElementById('btn-configure').addEventListener('click', () => {
            this.switchView('config');
        });

        document.getElementById('btn-back-to-deck').addEventListener('click', () => {
            this.switchView('deck');
        });

        // Grid configuration
        document.getElementById('btn-update-grid').addEventListener('click', () => {
            this.updateGrid();
        });

        // Add button
        document.getElementById('btn-add-button').addEventListener('click', () => {
            // Find first empty slot
            for (let row = 0; row < this.config.grid.rows; row++) {
                for (let col = 0; col < this.config.grid.cols; col++) {
                    const buttonId = `btn-${row}-${col}`;
                    if (!this.config.buttons.find(b => b.id === buttonId)) {
                        this.showConfigModal(buttonId);
                        return;
                    }
                }
            }
            alert('Grid is full! Increase grid size or delete a button.');
        });

        // Modal form
        document.getElementById('button-config-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveButton();
        });

        document.getElementById('cancel-btn').addEventListener('click', () => {
            this.hideConfigModal();
        });

        document.getElementById('delete-btn').addEventListener('click', () => {
            this.deleteButton();
        });

        // Action type change
        document.getElementById('action-type').addEventListener('change', (e) => {
            this.updateActionParams(e.target.value);
        });

        // Close modal on background click
        document.getElementById('config-modal').addEventListener('click', (e) => {
            if (e.target.id === 'config-modal') {
                this.hideConfigModal();
            }
        });
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            if (message.type === 'status_update') {
                this.updateStatusDisplay(message.data);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected, reconnecting...');
            setTimeout(() => this.connectWebSocket(), 3000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    async updateStatus() {
        try {
            const response = await fetch('/api/status');
            const status = await response.json();
            this.updateStatusDisplay(status);
        } catch (error) {
            console.error('Failed to get status:', error);
        }
    }

    updateStatusDisplay(status) {
        const streamIndicator = document.getElementById('stream-indicator');
        const recordIndicator = document.getElementById('record-indicator');
        const currentScene = document.getElementById('current-scene');

        if (status.streaming) {
            streamIndicator.classList.add('active');
        } else {
            streamIndicator.classList.remove('active');
        }

        if (status.recording) {
            recordIndicator.classList.add('active', 'recording');
        } else {
            recordIndicator.classList.remove('active', 'recording');
        }

        if (status.current_scene) {
            currentScene.textContent = status.current_scene;
        }
    }
}

// Initialize the deck when the page loads
document.addEventListener('DOMContentLoaded', () => {
    window.streamPiDeck = new StreamPiDeck();
});
