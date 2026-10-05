import { useState, type FormEvent } from 'react';
import { useAuth } from '@/auth/AuthContext';
import { authApi, usersApi } from '@/api/endpoints';
import { Avatar, Badge, Button, Card, ErrorBanner, PageHeader, Stack, Text, TextField, errorMessage, useToast } from '@/components/ui';
import { Screen } from '@/layout/Screen';

export function YouPage() {
  const { user, refreshUser, signOut, acceptSession } = useAuth();
  const toast = useToast();

  const [name, setName] = useState(user?.display_name ?? '');
  const [optIn, setOptIn] = useState(user?.directory_opt_in ?? true);
  const [profileErr, setProfileErr] = useState<string | null>(null);
  const [savingProfile, setSavingProfile] = useState(false);

  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');
  const [pwErr, setPwErr] = useState<string | null>(null);
  const [savingPw, setSavingPw] = useState(false);

  async function saveProfile(e: FormEvent) {
    e.preventDefault();
    setProfileErr(null);
    setSavingProfile(true);
    try {
      await usersApi.updateMe({ display_name: name, directory_opt_in: optIn });
      await refreshUser();
      toast({ title: 'Profile saved', tone: 'success' });
    } catch (err) {
      setProfileErr(errorMessage(err));
    } finally {
      setSavingProfile(false);
    }
  }

  async function changePassword(e: FormEvent) {
    e.preventDefault();
    setPwErr(null);
    setSavingPw(true);
    try {
      acceptSession(await authApi.changePassword(current, next));
      setCurrent('');
      setNext('');
      toast({ title: 'Password changed', body: 'Other devices were signed out.', tone: 'success' });
    } catch (err) {
      setPwErr(errorMessage(err));
    } finally {
      setSavingPw(false);
    }
  }

  if (!user) return null;

  return (
    <Screen title="You">
      <PageHeader
        title={user.display_name || user.email}
        description={
          <span>
            Unit <span className="font-mono">{user.unit_number}</span>
          </span>
        }
        actions={user.is_admin && <Badge tone="pin">Admin</Badge>}
      />

      <Card>
        <Stack row gap={4} align="center">
          <Avatar name={user.display_name || user.email} size="lg" />
          <div className="min-w-0">
            <Text variant="subtitle" className="truncate">
              {user.email}
            </Text>
            <Text variant="caption" as="p" tone="muted">
              Member since {new Date(user.created_at).toLocaleDateString()}
            </Text>
          </div>
        </Stack>
      </Card>

      <Card>
        <form onSubmit={(e) => void saveProfile(e)}>
          <Stack gap={4}>
            <Text variant="subtitle">About you</Text>
            <ErrorBanner message={profileErr} />
            <TextField label="Display name" required maxLength={80} value={name} onChange={(e) => setName(e.target.value)} />
            <label className="flex items-start gap-3">
              <input type="checkbox" checked={optIn} onChange={(e) => setOptIn(e.target.checked)} className="mt-1 size-4 accent-brand md:size-3.5" />
              <span>
                <Text variant="body">Show my name in the People directory</Text>
                <Text variant="caption" as="p" tone="muted">
                  Your unit stays visible while you're active, so a neighbor can reach you. You choose whether to accept.
                </Text>
              </span>
            </label>
            <Button type="submit" loading={savingProfile} className="self-start">
              Save
            </Button>
          </Stack>
        </form>
      </Card>

      <Card>
        <form onSubmit={(e) => void changePassword(e)}>
          <Stack gap={4}>
            <Text variant="subtitle">Password</Text>
            <ErrorBanner message={pwErr} />
            <TextField label="Current password" type="password" autoComplete="current-password" required value={current} onChange={(e) => setCurrent(e.target.value)} />
            <TextField label="New password" type="password" autoComplete="new-password" required minLength={10} hint="At least 10 characters." value={next} onChange={(e) => setNext(e.target.value)} />
            <Button type="submit" variant="secondary" loading={savingPw} className="self-start">
              Change password
            </Button>
          </Stack>
        </form>
      </Card>

      <Button variant="danger" onClick={() => void signOut()} className="self-start">
        Sign out
      </Button>
    </Screen>
  );
}
