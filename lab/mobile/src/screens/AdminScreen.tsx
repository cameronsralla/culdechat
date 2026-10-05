import { useCallback, useState } from 'react';
import { View } from 'react-native';
import { Redirect, useFocusEffect } from 'expo-router';
import { inviteResident, listAdminUsers, offboardUser } from '../api/community';
import type { AdminUser } from '../api/types';
import { useAuth } from '../auth/AuthContext';
import { AppText, Button, Card, EmptyState, ErrorBanner, PageHeader, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useCompactLayout } from '../components/layout/useCompactLayout';
import { copyText } from '../lib/clipboard';
import { confirm } from '../lib/confirm';
import { useStyles, type Theme } from '../theme';

type InviteSecrets = {
  email: string;
  registration_token: string;
  passcode: string;
  email_sent: boolean;
  link: string;
};

const stylesFor = (t: Theme) => ({
  actions: { flexDirection: 'row' as const, flexWrap: 'wrap' as const, gap: t.space.sm },
});

function inviteLink(token: string): string {
  if (typeof window !== 'undefined' && window.location?.origin) {
    return `${window.location.origin}/register?token=${encodeURIComponent(token)}`;
  }
  return `/register?token=${encodeURIComponent(token)}`;
}

export function AdminScreen() {
  const { user } = useAuth();
  const compact = useCompactLayout();
  const styles = useStyles(stylesFor);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteUnit, setInviteUnit] = useState('');
  const [invite, setInvite] = useState<InviteSecrets | null>(null);
  const [copied, setCopied] = useState<string | null>(null);
  const [roster, setRoster] = useState<AdminUser[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [loaded, setLoaded] = useState(false);

  const load = useCallback(async () => {
    try {
      setRoster(await listAdminUsers());
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load residents');
    } finally {
      setLoaded(true);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      if (user?.is_admin) {
        void load();
      }
    }, [load, user?.is_admin]),
  );

  if (!user?.is_admin) {
    return <Redirect href="/" />;
  }

  async function onInvite() {
    setSaving(true);
    try {
      const out = await inviteResident(inviteEmail.trim(), inviteUnit.trim());
      setInvite({
        email: out.email,
        registration_token: out.registration_token,
        passcode: out.passcode,
        email_sent: out.email_sent,
        link: inviteLink(out.registration_token),
      });
      setInviteEmail('');
      setInviteUnit('');
      setRoster(await listAdminUsers());
      setCopied(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not invite');
    } finally {
      setSaving(false);
    }
  }

  async function copy(label: string, value: string) {
    try {
      await copyText(value);
      setCopied(`Copied ${label}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not copy');
    }
  }

  return (
    <Screen
      scroll
      inShell
      refreshing={refreshing}
      onRefresh={() => {
        setRefreshing(true);
        void load().finally(() => setRefreshing(false));
      }}
    >
      <Stack gap="xl">
        <PageHeader
          title="Admin"
          subtitle="Invite neighbors and keep the membership list tidy. Stewardship, not surveillance."
          eyebrow="Stewards"
          hideTitleOnCompact
          compact={compact}
        />
        <ErrorBanner message={error} />
        {copied ? <AppText tone="brand">{copied}</AppText> : null}

        <Card>
          <Stack gap="md">
            <AppText variant="subtitle">Invite a resident</AppText>
            <TextField label="Email" value={inviteEmail} onChangeText={setInviteEmail} autoCapitalize="none" />
            <TextField label="Unit number" value={inviteUnit} onChangeText={setInviteUnit} />
            <Button
              label="Send invite"
              onPress={() => void onInvite()}
              loading={saving}
              disabled={!inviteEmail.trim() || !inviteUnit.trim()}
            />
            {invite ? (
              <Stack gap="sm">
                <AppText>
                  {invite.email_sent ? 'Invite email sent.' : 'Email did not send. Share these by hand.'}
                </AppText>
                <AppText variant="caption" tone="muted">
                  Token: {invite.registration_token}
                </AppText>
                <AppText variant="caption" tone="muted">
                  Passcode: {invite.passcode}
                </AppText>
                <AppText variant="caption" tone="muted">
                  {invite.link}
                </AppText>
                <View style={styles.actions}>
                  <Button label="Copy token" variant="ghost" onPress={() => void copy('token', invite.registration_token)} />
                  <Button label="Copy passcode" variant="ghost" onPress={() => void copy('passcode', invite.passcode)} />
                  <Button label="Copy link" variant="ghost" onPress={() => void copy('link', invite.link)} />
                </View>
              </Stack>
            ) : null}
          </Stack>
        </Card>

        <Card>
          <Stack gap="md">
            <AppText variant="subtitle">Residents</AppText>
            {loaded && roster.length === 0 ? (
              <EmptyState title={error ? 'Could not load residents.' : 'No residents yet.'} actionLabel={error ? 'Try again' : undefined} onAction={error ? () => void load() : undefined} />
            ) : null}
            {roster.map((row) => (
              <Stack key={row.id} gap="xs">
                <AppText>
                  {row.name || row.email} · Unit {row.unit_number} · {row.status}
                  {row.is_admin ? ' · Admin' : ''}
                </AppText>
                {!row.is_admin && row.status !== 'inactive' ? (
                  <Button
                    label="Offboard"
                    variant="danger"
                    onPress={() =>
                      confirm(`Offboard ${row.email}?`, () => {
                        void offboardUser(row.id)
                          .then(load)
                          .catch((err: unknown) => {
                            setError(err instanceof Error ? err.message : 'Could not offboard');
                          });
                      })
                    }
                  />
                ) : null}
              </Stack>
            ))}
          </Stack>
        </Card>
      </Stack>
    </Screen>
  );
}
