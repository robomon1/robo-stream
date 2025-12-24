// Stream-Pi Client JavaScript

class StreamPiClient {
    constructor() {
        this.config = null;
        this.editMode = false;
        this.ws = null;
        this.currentButton = null;
        this.scenes = [];
        this.inputs = [];

        this.init();
    }

    async init() {
        await this.loadConfig();
        this.renderGrid();
        this.setupEventListeners();
        this.connectWebSocket();
        await this.loadScenes();
        await this.loadInputs();
        await this.updateStatus();
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

    renderGrid() {
        const grid = document.getElementById('button-grid');
        grid.innerHTML = '';
        grid.style.gridTemplateColumns = `repeat(${this.config.grid.cols}, 1fr)`;

        // Create all grid positions
        for (let row = 0; row < this.config.grid.rows; row++) {
            for (let col = 0; col < this.config.grid.cols; col++) {
                const buttonId = `btn-${row}-${col}`;
                const button = this.config.buttons.find(b => b.id === buttonId);

                const btnElement = document.createElement('button');
                btnElement.className = 'stream-button';
                btnElement.dataset.id = buttonId;
                btnElement.dataset.row = row;
                btnElement.dataset.col = col;

                if (button) {
                    btnElement.textContent = button.text;
                    btnElement.style.backgroundColor = button.color;
                } else {
                    btnElement.classList.add('empty');
                    btnElement.textContent = '+';
                }

                if (this.editMode) {
                    btnElement.classList.add('edit-mode');
                }

                btnElement.addEventListener('click', () => this.handleButtonClick(buttonId));
                grid.appendChild(btnElement);
            }
        }
    }

    handleButtonClick(buttonId) {
        if (this.editMode) {
            this.showConfigModal(buttonId);
        } else {
            this.pressButton(buttonId);
        }
    }

    async pressButton(buttonId) {
        const button = this.config.buttons.find(b => b.id === buttonId);
        if (!button) return;

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
        document.getElementById('source-param').style.display = 'none';
        document.getElementById('visibility-param').style.display = 'none';

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
        } else if (actionType === 'set_source_visibility') {
            document.getElementById('source-param').style.display = 'block';
            document.getElementById('visibility-param').style.display = 'block';
            if (params.source_name) {
                document.getElementById('source-name').value = params.source_name;
            }
            if (params.visible !== undefined) {
                document.getElementById('visibility').value = params.visible.toString();
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
        } else if (actionType === 'set_source_visibility') {
            params.source_name = formData.get('source_name');
            params.visible = formData.get('visible') === 'true';
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
                this.renderGrid();
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
                this.renderGrid();
                this.hideConfigModal();
            }
        } catch (error) {
            console.error('Failed to delete button:', error);
        }
    }

    toggleEditMode() {
        this.editMode = !this.editMode;
        const btn = document.getElementById('edit-mode-btn');
        const saveBtn = document.getElementById('save-config-btn');

        if (this.editMode) {
            btn.textContent = 'Exit Edit Mode';
            btn.classList.add('btn-primary');
            btn.classList.remove('btn-secondary');
            saveBtn.style.display = 'inline-block';
        } else {
            btn.textContent = 'Edit Mode';
            btn.classList.add('btn-secondary');
            btn.classList.remove('btn-primary');
            saveBtn.style.display = 'none';
        }

        this.renderGrid();
    }

    setupEventListeners() {
        // Edit mode toggle
        document.getElementById('edit-mode-btn').addEventListener('click', () => {
            this.toggleEditMode();
        });

        // Add button
        document.getElementById('add-button-btn').addEventListener('click', () => {
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
            alert('Grid is full!');
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
        const streamStatus = document.getElementById('stream-status');
        const recordStatus = document.getElementById('record-status');
        const currentScene = document.getElementById('current-scene');

        if (status.streaming) {
            streamStatus.classList.add('active');
        } else {
            streamStatus.classList.remove('active');
        }

        if (status.recording) {
            recordStatus.classList.add('active', 'recording');
        } else {
            recordStatus.classList.remove('active', 'recording');
        }

        if (status.current_scene) {
            currentScene.textContent = status.current_scene;
        }
    }
}

// Initialize the client when the page loads
document.addEventListener('DOMContentLoaded', () => {
    window.streamPi = new StreamPiClient();
});
