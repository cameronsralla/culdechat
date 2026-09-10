import { useEffect, useState } from 'react';
import { KeyboardAvoidingView, Platform } from 'react-native';
import { Redirect, useLocalSearchParams, useRouter } from 'expo-router';
import { useAuth } from '../auth/AuthContext';
import { firstParam } from '../lib/params';
import { AppText, Button, ErrorBanner, Logo, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useCompactLayout } from '../components/layout/useCompactLayout';
import { useStyles, useTheme, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  flex: { flex: 1, backgroundColor: t.colors.paper },
  panelWide: {
    maxWidth: t.layout.loginMax,
    width: '100%' as const,
    alignSelf: 'center' as const,
    marginTop: t.space.xxxl,
    backgroundColor: t.colors.surface,
    borderWidth: t.layout.border,
    borderColor: t.colors.line,
    borderRadius: t.radius.md,
    padding: t.space.xxl,
  },
  hero: { alignItems: 'center' as const },
  tagline: { textAlign: 'center' as const },
});

export function RegisterScreen() {
  const { user, ready, completeRegister } = useAuth();
  const router = useRouter();
  const params = useLocalSearchParams<{ token?: string | string[] }>();
  const compact = useCompactLayout();
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  const tokenFromLink = firstParam(params.token);
  const [token, setToken] = useState(tokenFromLink);
  const [passcode, setPasscode] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (tokenFromLink) {
      setToken(tokenFromLink);
    }
  }, [tokenFromLink]);

  if (ready && user) {
    return <Redirect href="/" />;
  }

  async function onSubmit() {
    setError(null);
    setLoading(true);
    try {
      await completeRegister({ token: token.trim(), passcode: passcode.trim(), password, name: name.trim() });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not complete registration');
    } finally {
      setLoading(false);
    }
  }

  return (
    <KeyboardAvoidingView style={styles.flex} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <Screen scroll>
        <Stack gap="xl" style={!compact ? styles.panelWide : undefined}>
          <Stack gap="sm" style={styles.hero}>
            <Logo size={compact ? theme.layout.logoAuthCompact : theme.layout.logoAuth} />
            <AppText variant="display">Join Cul-de-Chat</AppText>
            <AppText variant="body" tone="muted" style={styles.tagline}>
              Use the token and passcode from your invite email.
            </AppText>
          </Stack>
          <Stack gap="lg">
            <ErrorBanner message={error} />
            <TextField label="Invite token" value={token} onChangeText={setToken} autoCapitalize="none" />
            <TextField label="Passcode" value={passcode} onChangeText={setPasscode} autoCapitalize="none" />
            <TextField label="Display name" value={name} onChangeText={setName} />
            <TextField label="Password" value={password} onChangeText={setPassword} secureTextEntry />
            <Button
              label="Finish registration"
              onPress={() => void onSubmit()}
              loading={loading}
              disabled={!token.trim() || !passcode.trim() || !name.trim() || password.length < 8}
            />
            <Button label="Back to login" variant="ghost" onPress={() => router.replace('/login')} />
          </Stack>
        </Stack>
      </Screen>
    </KeyboardAvoidingView>
  );
}
