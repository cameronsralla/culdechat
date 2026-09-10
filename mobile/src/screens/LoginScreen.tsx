import { useState } from 'react';
import { KeyboardAvoidingView, Platform } from 'react-native';
import { Redirect, useRouter } from 'expo-router';
import { useAuth } from '../auth/AuthContext';
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

export function LoginScreen() {
  const { user, ready, login } = useAuth();
  const router = useRouter();
  const compact = useCompactLayout();
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  if (ready && user) {
    return <Redirect href="/" />;
  }

  async function onSubmit() {
    setError(null);
    setLoading(true);
    try {
      await login(email.trim(), password);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not sign in';
      setError(message === 'invalid credentials' ? 'Invalid email or password' : message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <KeyboardAvoidingView
      style={styles.flex}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <Screen scroll>
        <Stack gap="xl" style={!compact ? styles.panelWide : undefined}>
          <Stack gap="sm" style={styles.hero}>
            <Logo size={compact ? theme.layout.logoAuthCompact : theme.layout.logoAuth} />
            <AppText variant="display">Cul-de-Chat</AppText>
            <AppText variant="body" tone="muted" style={styles.tagline}>
              The neighborhood town square.
            </AppText>
          </Stack>
          <Stack gap="lg">
            <ErrorBanner message={error} />
            <TextField
              label="Email"
              value={email}
              onChangeText={setEmail}
              autoCapitalize="none"
              autoComplete="email"
              keyboardType="email-address"
              textContentType="emailAddress"
            />
            <TextField
              label="Password"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
              autoComplete="password"
              textContentType="password"
            />
            <Button label="Log in" onPress={onSubmit} loading={loading} disabled={!email || !password} />
            <Button label="Have an invite?" variant="ghost" onPress={() => router.push('/register')} />
          </Stack>
        </Stack>
      </Screen>
    </KeyboardAvoidingView>
  );
}
