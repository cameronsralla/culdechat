import { useState } from 'react';
import { KeyboardAvoidingView, Platform, View } from 'react-native';
import { Redirect, useRouter } from 'expo-router';
import { useAuth } from '../auth/AuthContext';
import { AppText, Button, ErrorBanner, Logo, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useCompactLayout } from '../components/layout/useCompactLayout';
import { useStyles, useTheme, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  flex: { flex: 1, backgroundColor: t.colors.paper },
  wash: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    right: 0,
    height: 320,
    backgroundColor: t.colors.brandWash,
  },
  washOrb: {
    position: 'absolute' as const,
    top: -40,
    right: -60,
    width: 220,
    height: 220,
    borderRadius: 110,
    backgroundColor: t.colors.brandSoft,
    opacity: 0.9,
  },
  washCurve: {
    position: 'absolute' as const,
    top: 250,
    left: -40,
    right: -40,
    height: 120,
    borderRadius: 80,
    backgroundColor: t.colors.paper,
  },
  panel: {
    maxWidth: t.layout.loginMax,
    width: '100%' as const,
    alignSelf: 'center' as const,
    marginTop: t.space.xxl,
    backgroundColor: t.colors.surface,
    borderWidth: t.layout.border,
    borderColor: t.colors.lineSoft,
    borderRadius: t.radius.lg,
    padding: t.space.xxl,
    ...t.shadow.card,
  },
  panelCompact: {
    marginTop: t.space.xl,
    backgroundColor: 'transparent',
    borderWidth: 0,
    padding: 0,
    shadowOpacity: 0,
    elevation: 0,
    boxShadow: 'none',
  },
  hero: { alignItems: 'center' as const, gap: t.space.sm },
  tagline: { textAlign: 'center' as const, maxWidth: 320 },
  chip: {
    backgroundColor: t.colors.brandSoft,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.md,
    paddingVertical: t.space.xs,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
    marginTop: t.space.xs,
  },
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
    <KeyboardAvoidingView style={styles.flex} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <View style={styles.wash} />
      <View style={styles.washOrb} />
      <View style={styles.washCurve} />
      <Screen scroll>
        <Stack gap="xl" style={[styles.panel, compact ? styles.panelCompact : null]}>
          <Stack gap="sm" style={styles.hero}>
            <Logo size={compact ? theme.layout.logoAuthCompact : theme.layout.logoAuth} />
            <AppText variant="display">Cul-de-Chat</AppText>
            <AppText variant="body" tone="muted" style={styles.tagline}>
              A private town square for the people who share your place.
            </AppText>
            <View style={styles.chip}>
              <AppText variant="label" tone="brand">
                Invite-only · Neighbors only
              </AppText>
            </View>
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
