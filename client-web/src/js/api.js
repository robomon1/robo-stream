// API Client - Pure JavaScript HTTP client (replaces Go backend)

export class APIClient {
  constructor(serverURL) {
    this.serverURL = serverURL;
    this.sessionID = null;
    this.clientID = this.loadOrCreateClientID();
  }

  // Load or create persistent client ID
  loadOrCreateClientID() {
    let id = localStorage.getItem('client_id');
    if (!id) {
      const hostname = window.location.hostname || 'unknown';
      const timestamp = Date.now();
      const random = Math.random().toString(36).substr(2, 9);
      id = `web-${hostname}-${timestamp}-${random}`;
      localStorage.setItem('client_id', id);
    }
    return id;
  }

  // Get server info (health check)
  async getServerInfo() {
    try {
      const response = await fetch(`${this.serverURL}/api/health`);
      if (!response.ok) throw new Error(`Server returned ${response.status}`);
      return await response.json();
    } catch (err) {
      throw new Error(`Failed to connect to server: ${err.message}`);
    }
  }

  // Register with server and get session ID
  async register() {
    try {
      const response = await fetch(`${this.serverURL}/api/client/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: this.clientID,
          client_name: 'Web Client'
        })
      });
      
      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Registration failed: ${error}`);
      }
      
      const data = await response.json();
      this.sessionID = data.session_id;
      
      // Save session ID to localStorage for persistence
      localStorage.setItem('session_id', this.sessionID);
      
      console.log('✓ Registered with server - Session:', this.sessionID);
      return data.config;
    } catch (err) {
      throw new Error(`Failed to register: ${err.message}`);
    }
  }

  // Get all configurations
  async getConfigurations() {
    try {
      const response = await fetch(`${this.serverURL}/api/configurations`);
      if (!response.ok) throw new Error(`Server returned ${response.status}`);
      return await response.json();
    } catch (err) {
      throw new Error(`Failed to get configurations: ${err.message}`);
    }
  }

  // Get specific configuration and switch to it
  async getConfiguration(configID) {
    try {
      if (!this.sessionID) {
        throw new Error('Not registered - no session ID');
      }

      const response = await fetch(`${this.serverURL}/api/client/config/${configID}`, {
        method: 'PUT',
        headers: { 'X-Session-ID': this.sessionID }
      });
      
      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Failed to load configuration: ${error}`);
      }
      
      return await response.json();
    } catch (err) {
      throw new Error(`Failed to get configuration: ${err.message}`);
    }
  }

  // Execute button action
  async executeAction(action) {
    try {
      if (!this.sessionID) {
        throw new Error('Not registered - no session ID');
      }
      
      const response = await fetch(`${this.serverURL}/api/action`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Session-ID': this.sessionID
        },
        body: JSON.stringify(action)
      });
      
      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Action failed: ${error}`);
      }
    } catch (err) {
      throw new Error(`Failed to execute action: ${err.message}`);
    }
  }

  // Get OBS status
  async getOBSStatus() {
    try {
      const response = await fetch(`${this.serverURL}/api/obs/status`);
      if (!response.ok) throw new Error(`Server returned ${response.status}`);
      return await response.json();
    } catch (err) {
      console.error('Failed to get OBS status:', err);
      return {
        connected: false,
        streaming: false,
        recording: false,
        current_scene: '',
        virtual_cam_active: false,
        replay_buffer_active: false,
        studio_mode_active: false
      };
    }
  }

  // Get input mute state
  async getInputMute(inputName) {
    try {
      const params = new URLSearchParams({ input: inputName });
      const response = await fetch(`${this.serverURL}/api/obs/input-mute?${params}`);
      if (!response.ok) throw new Error(`Server returned ${response.status}`);
      const data = await response.json();
      return data.muted;
    } catch (err) {
      console.error('Failed to get input mute:', err);
      return false;
    }
  }

  // Get source filter enabled state
  async getSourceFilterEnabled(sourceName, filterName) {
    try {
      const params = new URLSearchParams({ source: sourceName, filter: filterName });
      const response = await fetch(`${this.serverURL}/api/obs/source-filter?${params}`);
      if (!response.ok) throw new Error(`Server returned ${response.status}`);
      const data = await response.json();
      return data.enabled;
    } catch (err) {
      console.error('Failed to get source filter state:', err);
      return false;
    }
  }

  // Get source visibility
  async getSourceVisibility(sceneName, sourceName) {
    try {
      const params = new URLSearchParams({ 
        scene: sceneName, 
        source: sourceName 
      });
      
      const response = await fetch(
        `${this.serverURL}/api/obs/source-visibility?${params}`
      );
      
      if (!response.ok) throw new Error(`Server returned ${response.status}`);
      
      const data = await response.json();
      return data.visible;
    } catch (err) {
      console.error('Failed to get source visibility:', err);
      return false;
    }
  }
}
