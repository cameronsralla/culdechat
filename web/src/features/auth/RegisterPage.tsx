import { useState, type FormEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router';
import { useQuery } from '@tanstack/react-query';
import { authApi } from '@/api/endpoints';
import { useAuth } from '@/auth/AuthContext';
import { Button, ErrorBanner, Stack, Text, TextField, errorMessage } from '@/components/ui';
import { AuthShell } from '@/layout/AuthShell';

/** Invite completion: /register?token=... */
export function RegisterPage() {
  const [params] = useSearchParams();
  const token = params.get('token') ?? '';
  const nav = useNavigate();
  const { acceptSession } = useAuth();

  const peek = useQuery({ queryKey: ['invite', token], queryFn: () => authApi.peekInvite(token), enabled: !!token, retry: false });

  const [displayName, setDisplayName] = useState('');
  const [passcode, setPasscode] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (password !== confirm) {
      setError('Passwords do not match.');
      return;
    }
    setError(null);
    setBusy(true);
    try {
      acceptSession(await authApi.completeInvite({ token, passcode, password, display_name: displayName || peek.data?.display_name || '' }));
      nav('/', { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  if (!token || peek.isError) {
    return (
      <AuthShell title="Invitation not found" subtitle="This link is invalid or has expired. Ask your community admin for a new one.">
        <Button variant="secondary" full onClick={() => nav('/login')}>
          Back to sign in
        </Button>
      </AuthShell>
    );
  }

  return (
    <AuthShell title="Join your community" subtitle={peek.data ? `Setting up ${peek.data.email} · Unit ${peek.data.unit_number}` : undefined}>
      <form onSubmit={(e) => void submit(e)}>
        <Stack gap={4}>
          <ErrorBanner message={error} />
          <TextField label="Your name" autoComplete="name" required value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder={peek.data?.display_name || 'How neighbors will see you'} />
          <TextField label="Passcode" inputMode="numeric" autoComplete="one-time-code" required hint="The 6-digit code your admin gave you." value={passcode} onChange={(e) => setPasscode(e.target.value)} />
          <TextField label="Password" type="password" autoComplete="new-password" required minLength={10} hint="At least 10 characters." value={password} onChange={(e) => setPassword(e.target.value)} />
          <TextField label="Confirm password" type="password" autoComplete="new-password" required value={confirm} onChange={(e) => setConfirm(e.target.value)} />
          <Button type="submit" full loading={busy || peek.isLoading}>
            Create account
          </Button>
          <Text variant="caption" tone="muted" className="text-center">
            Already set up? <a href="/login" className="text-brand underline">Sign in</a>
          </Text>
        </Stack>
      </form>
    </AuthShell>
  );
}
