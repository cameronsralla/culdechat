import { useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/api/endpoints';
import type { InviteResult, User } from '@/api/types';
import { useAuth } from '@/auth/AuthContext';
import { Avatar, Badge, Button, Card, ErrorBanner, ListGroup, ListRow, PageHeader, Stack, Text, TextField, errorMessage, useToast } from '@/components/ui';
import { Screen } from '@/layout/Screen';

export function AdminPage() {
  const qc = useQueryClient();
  const toast = useToast();
  const { user: me } = useAuth();
  const users = useQuery({ queryKey: ['admin', 'users'], queryFn: adminApi.users });

  const [email, setEmail] = useState('');
  const [unit, setUnit] = useState('');
  const [name, setName] = useState('');
  const [last, setLast] = useState<InviteResult | null>(null);

  const invite = useMutation({
    mutationFn: adminApi.invite,
    onSuccess: (res) => {
      setLast(res);
      setEmail('');
      setUnit('');
      setName('');
      void qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
  const reinvite = useMutation({ mutationFn: adminApi.reinvite, onSuccess: setLast });
  const setStatus = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) => adminApi.setStatus(id, active),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['admin', 'users'] }),
    onError: (e) => toast({ title: 'Could not update', body: errorMessage(e), tone: 'danger' }),
  });
  const setAdmin = useMutation({
    mutationFn: ({ id, is_admin }: { id: string; is_admin: boolean }) => adminApi.setAdmin(id, is_admin),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['admin', 'users'] }),
    onError: (e) => toast({ title: 'Could not update', body: errorMessage(e), tone: 'danger' }),
  });

  function submit(e: FormEvent) {
    e.preventDefault();
    invite.mutate({ email, unit_number: unit, display_name: name || undefined });
  }

  return (
    <Screen title="Admin" wide>
      <PageHeader eyebrow="Community" title="Admin" description="Invite residents and manage the roster." />

      <Card>
        <form onSubmit={submit}>
          <Stack gap={4}>
            <Text variant="subtitle">Invite a resident</Text>
            <ErrorBanner message={invite.isError ? errorMessage(invite.error) : null} />
            <div className="grid gap-4 md:grid-cols-3">
              <TextField label="Email" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
              <TextField label="Unit" required maxLength={32} value={unit} onChange={(e) => setUnit(e.target.value)} />
              <TextField label="Name (optional)" maxLength={80} value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <Button type="submit" loading={invite.isPending} className="self-start">
              Send invite
            </Button>
          </Stack>
        </form>
      </Card>

      {last && (
        <Card tone="pin">
          <Stack gap={2}>
            <Text variant="subtitle">Passcode for {last.user.email}</Text>
            <Text variant="display" as="p" className="tracking-widest">
              {last.passcode}
            </Text>
            <Text variant="caption" tone="muted">
              Give this to them in person or by text. The email link alone won't work without it. Expires {new Date(last.expires_at).toLocaleDateString()}.
            </Text>
            {last.invite_url && (
              <Text variant="caption" tone="muted" className="break-all">
                Dev link: {last.invite_url}
              </Text>
            )}
          </Stack>
        </Card>
      )}

      <Stack gap={3}>
        <Text variant="subtitle">Residents</Text>
        {users.isError && <ErrorBanner message={errorMessage(users.error)} />}
        {users.data && (
          <ListGroup>
            {users.data.map((u) => (
              <UserRow
                key={u.id}
                u={u}
                isSelf={u.id === me?.id}
                onReinvite={() => reinvite.mutate(u.id)}
                onToggleActive={() => setStatus.mutate({ id: u.id, active: u.status !== 'active' })}
                onToggleAdmin={() => setAdmin.mutate({ id: u.id, is_admin: !u.is_admin })}
              />
            ))}
          </ListGroup>
        )}
      </Stack>
    </Screen>
  );
}

function UserRow({ u, isSelf, onReinvite, onToggleActive, onToggleAdmin }: { u: User; isSelf: boolean; onReinvite: () => void; onToggleActive: () => void; onToggleAdmin: () => void }) {
  return (
    <ListRow
      leading={<Avatar name={u.display_name || u.email} size="sm" />}
      title={
        <span className="flex items-center gap-2">
          {u.display_name || u.email}
          {u.is_admin && <Badge tone="pin">Admin</Badge>}
          {u.status === 'invited' && <Badge tone="neutral">Invited</Badge>}
          {u.status === 'inactive' && <Badge tone="danger">Inactive</Badge>}
        </span>
      }
      subtitle={`${u.email} · Unit ${u.unit_number}`}
      trailing={
        <div className="flex shrink-0 flex-wrap justify-end gap-1.5">
          {u.status === 'invited' && (
            <Button size="sm" variant="ghost" onClick={onReinvite}>
              Resend
            </Button>
          )}
          {!isSelf && u.status !== 'invited' && (
            <>
              <Button size="sm" variant="ghost" onClick={onToggleAdmin}>
                {u.is_admin ? 'Remove admin' : 'Make admin'}
              </Button>
              <Button size="sm" variant={u.status === 'active' ? 'danger' : 'secondary'} onClick={onToggleActive}>
                {u.status === 'active' ? 'Deactivate' : 'Reactivate'}
              </Button>
            </>
          )}
        </div>
      }
    />
  );
}
