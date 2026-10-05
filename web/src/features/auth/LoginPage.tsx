import { useState, type FormEvent } from 'react';
import { useLocation, useNavigate } from 'react-router';
import { useAuth } from '@/auth/AuthContext';
import { Button, ErrorBanner, Stack, TextField, Text, errorMessage } from '@/components/ui';
import { AuthShell } from '@/layout/AuthShell';

export function LoginPage() {
  const { signIn } = useAuth();
  const nav = useNavigate();
  const loc = useLocation() as { state?: { from?: string } };
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      await signIn(email, password);
      nav(loc.state?.from ?? '/', { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthShell title="Welcome back" subtitle="Sign in to your community square.">
      <form onSubmit={(e) => void submit(e)}>
        <Stack gap={4}>
          <ErrorBanner message={error} />
          <TextField label="Email" type="email" autoComplete="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
          <TextField label="Password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
          <Button type="submit" full loading={busy}>
            Sign in
          </Button>
          <Text variant="caption" tone="muted" className="text-center">
            New here? Use the invitation link your community admin sent you.
          </Text>
        </Stack>
      </form>
    </AuthShell>
  );
}
