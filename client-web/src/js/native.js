// Native Features - Capacitor Plugin Integration
// Add to src/js/native.js
import { KeepAwake } from '@capacitor-community/keep-awake';

// Check if running as native app
export function isNativeApp() {
  return window.Capacitor?.isNativePlatform() || false;
}

// Get platform (ios, android, web)
export function getPlatform() {
  if (!window.Capacitor) return 'web';
  return window.Capacitor.getPlatform();
}

// Initialize native features
export async function initializeNativeFeatures() {
  if (!isNativeApp()) {
    console.log('Running as web app - native features disabled');
    return;
  }
  
  console.log('Initializing native features...');
  const platform = getPlatform();
  console.log('Platform:', platform);
  
  try {
    // Initialize status bar
    await initializeStatusBar();
    
    // Initialize splash screen
    await initializeSplashScreen();
    
    // Initialize keyboard
    await initializeKeyboard();
    
    // Initialize app listeners
    initializeAppListeners();

    // Keep screen awake while app is active
    await keepScreenAwake(true);

    console.log('✓ Native features initialized');
  } catch (error) {
    console.error('Failed to initialize native features:', error);
  }
}

// Status Bar
async function initializeStatusBar() {
  if (!window.Capacitor?.Plugins?.StatusBar) return;
  
  const { StatusBar } = window.Capacitor.Plugins;
  
  try {
    await StatusBar.setStyle({ style: 'DARK' });
    await StatusBar.setBackgroundColor({ color: '#0f1419' });
    console.log('✓ Status bar configured');
  } catch (error) {
    console.warn('Status bar setup failed:', error);
  }
}

// Splash Screen
async function initializeSplashScreen() {
  if (!window.Capacitor?.Plugins?.SplashScreen) return;
  
  const { SplashScreen } = window.Capacitor.Plugins;
  
  try {
    // Hide splash screen after app is ready
    setTimeout(async () => {
      await SplashScreen.hide();
      console.log('✓ Splash screen hidden');
    }, 2000);
  } catch (error) {
    console.warn('Splash screen setup failed:', error);
  }
}

// Keyboard
async function initializeKeyboard() {
  if (!window.Capacitor?.Plugins?.Keyboard) return;
  
  const { Keyboard } = window.Capacitor.Plugins;
  
  try {
    await Keyboard.setResizeMode({ mode: 'body' });
    await Keyboard.setStyle({ style: 'DARK' });
    console.log('✓ Keyboard configured');
  } catch (error) {
    console.warn('Keyboard setup failed:', error);
  }
}

// App Listeners (pause, resume, back button)
function initializeAppListeners() {
  if (!window.Capacitor?.Plugins?.App) return;
  
  const { App } = window.Capacitor.Plugins;
  
  // App state change
  App.addListener('appStateChange', ({ isActive }) => {
    console.log('App state changed:', isActive ? 'active' : 'background');

    if (isActive) {
      // App came to foreground - refresh status and re-acquire wake lock
      console.log('App resumed - refreshing status');
      keepScreenAwake(true);
      window.dispatchEvent(new CustomEvent('app-resumed'));
    } else {
      // App went to background - release wake lock
      keepScreenAwake(false);
    }
  });
  
  // Back button (Android)
  App.addListener('backButton', ({ canGoBack }) => {
    console.log('Back button pressed');
    
    // Check if any modal is open
    const modals = document.querySelectorAll('.modal.open');
    if (modals.length > 0) {
      // Close the topmost modal
      modals[modals.length - 1].classList.remove('open');
      return;
    }
    
    // If no modals and can't go back, minimize app
    if (!canGoBack) {
      App.minimizeApp();
    }
  });
  
  console.log('✓ App listeners configured');
}

// Haptic Feedback
export async function hapticFeedback(style = 'MEDIUM') {
  if (!isNativeApp()) return;
  if (!window.Capacitor?.Plugins?.Haptics) return;
  
  const { Haptics, ImpactStyle } = window.Capacitor.Plugins;
  
  try {
    await Haptics.impact({ style: ImpactStyle[style] });
  } catch (error) {
    console.warn('Haptic feedback failed:', error);
  }
}

// Haptic notification
export async function hapticNotification(type = 'SUCCESS') {
  if (!isNativeApp()) return;
  if (!window.Capacitor?.Plugins?.Haptics) return;
  
  const { Haptics, NotificationType } = window.Capacitor.Plugins;
  
  try {
    await Haptics.notification({ type: NotificationType[type] });
  } catch (error) {
    console.warn('Haptic notification failed:', error);
  }
}

// Haptic selection (light feedback)
export async function hapticSelection() {
  if (!isNativeApp()) return;
  if (!window.Capacitor?.Plugins?.Haptics) return;
  
  const { Haptics } = window.Capacitor.Plugins;
  
  try {
    await Haptics.selectionStart();
    setTimeout(() => Haptics.selectionEnd(), 50);
  } catch (error) {
    console.warn('Haptic selection failed:', error);
  }
}

// Keep screen awake (prevent sleep during streaming)
export async function keepScreenAwake(enable = true) {
  if (!isNativeApp()) return;

  try {
    if (enable) {
      await KeepAwake.keepAwake();
      console.log('✓ Screen will stay awake');
    } else {
      await KeepAwake.allowSleep();
      console.log('✓ Screen can sleep');
    }
  } catch (error) {
    console.warn('Keep awake failed:', error);
  }
}

// Get device info
export async function getDeviceInfo() {
  if (!isNativeApp()) {
    return {
      platform: 'web',
      model: 'Browser',
      manufacturer: 'Unknown',
      osVersion: navigator.userAgent
    };
  }
  
  if (!window.Capacitor?.Plugins?.Device) return null;
  
  const { Device } = window.Capacitor.Plugins;
  
  try {
    const info = await Device.getInfo();
    console.log('Device info:', info);
    return info;
  } catch (error) {
    console.error('Failed to get device info:', error);
    return null;
  }
}

// Share functionality
export async function shareContent(title, text, url) {
  if (!isNativeApp()) {
    // Fallback to Web Share API
    if (navigator.share) {
      try {
        await navigator.share({ title, text, url });
        return true;
      } catch (error) {
        console.warn('Web share failed:', error);
        return false;
      }
    }
    return false;
  }
  
  if (!window.Capacitor?.Plugins?.Share) return false;
  
  const { Share } = window.Capacitor.Plugins;
  
  try {
    await Share.share({ title, text, url });
    return true;
  } catch (error) {
    console.warn('Share failed:', error);
    return false;
  }
}

// Open URL in system browser
export async function openURL(url) {
  if (!isNativeApp()) {
    window.open(url, '_blank');
    return;
  }
  
  if (!window.Capacitor?.Plugins?.Browser) {
    window.open(url, '_blank');
    return;
  }
  
  const { Browser } = window.Capacitor.Plugins;
  
  try {
    await Browser.open({ url });
  } catch (error) {
    console.warn('Browser open failed:', error);
    window.open(url, '_blank');
  }
}

// Show native toast (Android) or alert (iOS)
export async function showToast(message, duration = 'SHORT') {
  if (!isNativeApp()) {
    // Fallback to banner
    const banner = document.getElementById('connection-banner');
    const bannerMessage = document.getElementById('banner-message');
    if (banner && bannerMessage) {
      bannerMessage.textContent = message;
      banner.className = 'banner show';
      const timeout = duration === 'LONG' ? 3500 : 2000;
      setTimeout(() => banner.classList.remove('show'), timeout);
    }
    return;
  }
  
  if (!window.Capacitor?.Plugins?.Toast) {
    console.log('Toast plugin not available');
    return;
  }
  
  const { Toast } = window.Capacitor.Plugins;
  
  try {
    await Toast.show({ text: message, duration });
  } catch (error) {
    console.warn('Toast failed:', error);
  }
}

// Request permissions (camera, etc - for future features)
export async function requestPermissions(permissions = []) {
  if (!isNativeApp()) return {};
  
  const { Capacitor } = window;
  const results = {};
  
  for (const permission of permissions) {
    try {
      if (Capacitor.Plugins[permission]) {
        const result = await Capacitor.Plugins[permission].requestPermissions();
        results[permission] = result;
      }
    } catch (error) {
      console.warn(`Permission request failed for ${permission}:`, error);
      results[permission] = { error };
    }
  }
  
  return results;
}

// Get app info
export async function getAppInfo() {
  if (!isNativeApp()) {
    return {
      name: 'RoboStream',
      version: typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : '0.0.0',
      build: 'web'
    };
  }
  
  if (!window.Capacitor?.Plugins?.App) return null;
  
  const { App } = window.Capacitor.Plugins;
  
  try {
    const info = await App.getInfo();
    console.log('App info:', info);
    return info;
  } catch (error) {
    console.error('Failed to get app info:', error);
    return null;
  }
}

// Export all functions
export default {
  isNativeApp,
  getPlatform,
  initializeNativeFeatures,
  hapticFeedback,
  hapticNotification,
  hapticSelection,
  keepScreenAwake,
  getDeviceInfo,
  shareContent,
  openURL,
  showToast,
  requestPermissions,
  getAppInfo
};
