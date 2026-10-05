import { Slot } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { View } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import {
  Nunito_400Regular,
  Nunito_600SemiBold,
  Nunito_700Bold,
  Nunito_800ExtraBold,
  useFonts,
} from '@expo-google-fonts/nunito';
import { AuthProvider } from '../src/auth/AuthContext';
import { LoadingScreen } from '../src/components/layout/LoadingScreen';
import { ThemeProvider, useStyles, type Theme } from '../src/theme';

const stylesFor = (t: Theme) => ({
  root: { flex: 1, backgroundColor: t.colors.paper },
});

function ThemedTree() {
  const [fontsLoaded] = useFonts({
    Nunito_400Regular,
    Nunito_600SemiBold,
    Nunito_700Bold,
    Nunito_800ExtraBold,
  });
  const styles = useStyles(stylesFor);

  if (!fontsLoaded) {
    return <LoadingScreen />;
  }

  return (
    <View style={styles.root}>
      <StatusBar style="dark" />
      <AuthProvider>
        <Slot />
      </AuthProvider>
    </View>
  );
}

export default function RootLayout() {
  return (
    <SafeAreaProvider>
      <ThemeProvider>
        <ThemedTree />
      </ThemeProvider>
    </SafeAreaProvider>
  );
}
