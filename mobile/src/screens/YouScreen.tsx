import { createElement, useCallback, useState } from 'react';
import { Platform, View } from 'react-native';
import { useFocusEffect } from 'expo-router';
import { changePassword, getProfile, updateProfile, uploadProfilePhoto } from '../api/community';
import type { Profile } from '../api/types';
import { useAuth } from '../auth/AuthContext';
import { AppText, Avatar, Button, Card, ErrorBanner, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  identity: { flexDirection: 'row' as const, alignItems: 'center' as const, gap: t.space.lg },
  grow: { flex: 1 },
});

function PhotoPicker({ onFile }: { onFile: (file: Blob) => void }) {
  if (Platform.OS !== 'web') {
    return <AppText tone="muted">Photo upload is available in the web app for now.</AppText>;
  }
  return createElement('input', {
    type: 'file',
    accept: 'image/jpeg,image/png,image/webp',
    onChange: (event: { currentTarget: { files: FileList | null; value: string } }) => {
      const file = event.currentTarget.files?.[0];
      if (file) {
        onFile(file);
      }
      event.currentTarget.value = '';
    },
  });
}

export function YouScreen() {
  const { user, logout, refreshUser } = useAuth();
  const styles = useStyles(stylesFor);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [name, setName] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async () => {
    try {
      const next = await getProfile();
      setProfile(next);
      setName(next.name);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load profile');
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      void load();
    }, [load]),
  );

  async function saveProfile() {
    setSaving(true);
    try {
      const next = await updateProfile({ name: name.trim() });
      setProfile(next);
      await refreshUser();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save profile');
    } finally {
      setSaving(false);
    }
  }

  async function onPhoto(file: Blob) {
    setSaving(true);
    try {
      const next = await uploadProfilePhoto(file);
      setProfile(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not upload photo');
    } finally {
      setSaving(false);
    }
  }

  async function onOptIn() {
    if (!profile) {
      return;
    }
    try {
      const next = await updateProfile({ directory_opt_in: !profile.directory_opt_in });
      setProfile(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not update directory setting');
    }
  }

  async function onPassword() {
    setSaving(true);
    try {
      await changePassword(currentPassword, newPassword);
      setCurrentPassword('');
      setNewPassword('');
      setNotice('Password changed. Other sessions were signed out.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not change password');
    } finally {
      setSaving(false);
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
        <ErrorBanner message={error} />
        {notice ? <AppText tone="brand">{notice}</AppText> : null}

        <Card>
          <Stack gap="md">
            <View style={styles.identity}>
              <Avatar path={profile?.profile_picture_url} label={profile?.name || user?.name} size="lg" />
              <View style={styles.grow}>
                <AppText variant="subtitle">{profile?.name || user?.name || 'Resident'}</AppText>
                <AppText tone="muted">{profile?.email || user?.email}</AppText>
                <AppText variant="caption" tone="muted">
                  Unit {profile?.unit_number || user?.unit_number}
                  {profile?.is_admin || user?.is_admin ? ' · Admin' : ''}
                </AppText>
              </View>
            </View>
            <PhotoPicker onFile={(file) => void onPhoto(file)} />
            <TextField label="Display name" value={name} onChangeText={setName} />
            <Button label="Save name" onPress={() => void saveProfile()} loading={saving} disabled={!name.trim()} />
            <Button
              label={profile?.directory_opt_in ? 'Listed in directory' : 'Hidden from directory'}
              variant="ghost"
              onPress={() => void onOptIn()}
            />
            <Button label="Log out" variant="ghost" onPress={() => void logout()} />
          </Stack>
        </Card>

        <Card>
          <Stack gap="md">
            <AppText variant="subtitle">Change password</AppText>
            <TextField label="Current password" value={currentPassword} onChangeText={setCurrentPassword} secureTextEntry />
            <TextField label="New password" value={newPassword} onChangeText={setNewPassword} secureTextEntry />
            <Button
              label="Update password"
              onPress={() => void onPassword()}
              loading={saving}
              disabled={!currentPassword || newPassword.length < 8}
            />
          </Stack>
        </Card>
      </Stack>
    </Screen>
  );
}
