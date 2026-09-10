import { Platform } from 'react-native';
import * as SecureStore from 'expo-secure-store';

const accessKey = 'culdechat.access';
const refreshKey = 'culdechat.refresh';

async function getItem(key: string): Promise<string | null> {
  if (Platform.OS === 'web') {
    try {
      return window.localStorage.getItem(key);
    } catch {
      return null;
    }
  }
  return SecureStore.getItemAsync(key);
}

async function setItem(key: string, value: string): Promise<void> {
  if (Platform.OS === 'web') {
    window.localStorage.setItem(key, value);
    return;
  }
  await SecureStore.setItemAsync(key, value);
}

async function removeItem(key: string): Promise<void> {
  if (Platform.OS === 'web') {
    window.localStorage.removeItem(key);
    return;
  }
  await SecureStore.deleteItemAsync(key);
}

export async function loadTokens(): Promise<{ access: string | null; refresh: string | null }> {
  const [access, refresh] = await Promise.all([getItem(accessKey), getItem(refreshKey)]);
  return { access, refresh };
}

export async function saveTokens(access: string, refresh: string): Promise<void> {
  await Promise.all([setItem(accessKey, access), setItem(refreshKey, refresh)]);
}

export async function clearTokens(): Promise<void> {
  await Promise.all([removeItem(accessKey), removeItem(refreshKey)]);
}
