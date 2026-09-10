import { Alert, Platform } from 'react-native';

export function confirm(message: string, onYes: () => void) {
  if (Platform.OS === 'web') {
    if (typeof window !== 'undefined' && window.confirm(message)) {
      onYes();
    }
    return;
  }
  Alert.alert('Confirm', message, [
    { text: 'Cancel', style: 'cancel' },
    { text: 'OK', onPress: onYes },
  ]);
}
