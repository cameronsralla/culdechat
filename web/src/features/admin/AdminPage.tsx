import { useState, type FormEvent } from 'react';
import { Link } from 'react-router';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/api/endpoints';
import type { InviteResult, User } from '@/api/types';
import { useAuth } from '@/auth/AuthContext';
import {
  Badge,
  Button,
  Card,
  DataTable,
  ErrorBanner,
  PageHeader,
  Stack,
  Text,
  TextField,
  errorMessage,
  useToast,
  type DataTableColumn,
  type DataTableQuery,
} from '@/components/ui';
import { Screen } from '@/layout/Screen';

type TempSecret =
  | { kind: 'invite'; result: InviteResult }
  | { kind: 'reset'; rows: { email: string; temporary_password: string }[] };

const columns: DataTableColumn<User>[] = [
  {
    key: 'display_name',
    header: 'Name',
    sortable: true,
    filter: 'text',
    render: (u) => u.display_name || '—',
  },
  {
    key: 'email',
    header: 'Email',
    sortable: true,
    filter: 'text',
    render: (u) => u.email,
  },
  {
    key: 'unit_number',
    header: 'Unit',
    sortable: true,
    filter: 'text',
    className: 'font-mono',
    render: (u) => u.unit_number,
  },
  {
    key: 'is_primary',
    header: 'Primary',
    render: (u) => (u.is_primary ? 'Yes' : 'No'),
  },
  {
    key: 'status',
    header: 'Status',
    sortable: true,
    filter: 'select',
    options: [
      { value: 'invited', label: 'Invited' },
      { value: 'active', label: 'Active' },
      { value: 'inactive', label: 'Inactive' },
    ],
    render: (u) => <StatusBadge status={u.status} />,
  },
  {
    key: 'is_admin',
    header: 'Admin',
    sortable: true,
    filter: 'select',
    options: [
      { value: 'true', label: 'Admin' },
      { value: 'false', label: 'Not admin' },
    ],
    render: (u) => (u.is_admin ? 'Yes' : 'No'),
  },
  {
    key: 'directory_opt_in',
    header: 'Directory',
    sortable: true,
    filter: 'select',
    options: [
      { value: 'true', label: 'Listed' },
      { value: 'false', label: 'Hidden' },
    ],
    render: (u) => (u.directory_opt_in ? 'Listed' : 'Hidden'),
  },
  {
    key: 'created_at',
    header: 'Joined',
    sortable: true,
    className: 'font-mono',
    render: (u) => new Date(u.created_at).toLocaleDateString(),
  },
];

export function AdminPage() {
  const qc = useQueryClient();
  const toast = useToast();
  const { user: me } = useAuth();
  const [query, setQuery] = useState<DataTableQuery>({
    q: '',
    sortKey: 'unit_number',
    sortDir: 'asc',
    page: 1,
    pageSize: 10,
    filters: {},
  });
  const [picked, setPicked] = useState<Map<string, User>>(new Map());
  const [email, setEmail] = useState('');
  const [unitID, setUnitID] = useState('');
  const [name, setName] = useState('');
  const [secret, setSecret] = useState<TempSecret | null>(null);

  const vacantUnits = useQuery({
    queryKey: ['admin', 'units', 'vacant'],
    queryFn: () => adminApi.units({ vacant: true, pageSize: 100 }),
  });

  const users = useQuery({
    queryKey: ['admin', 'users', query],
    queryFn: () =>
      adminApi.users({
        q: query.q,
        sort: query.sortKey,
        dir: query.sortDir,
        page: query.page,
        pageSize: query.pageSize,
        name: query.filters.display_name,
        email: query.filters.email,
        unit: query.filters.unit_number,
        status: query.filters.status,
        admin: query.filters.is_admin,
        directory: query.filters.directory_opt_in,
      }),
  });

  const invite = useMutation({
    mutationFn: adminApi.invite,
    onSuccess: (res) => {
      setSecret({ kind: 'invite', result: res });
      setEmail('');
      setUnitID('');
      setName('');
      void qc.invalidateQueries({ queryKey: ['admin', 'users'] });
      void qc.invalidateQueries({ queryKey: ['admin', 'units'] });
    },
  });
  const reinvite = useMutation({ mutationFn: adminApi.reinvite });
  const resetPassword = useMutation({ mutationFn: adminApi.resetPassword });
  const setStatus = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) => adminApi.setStatus(id, active),
  });
  const setAdmin = useMutation({
    mutationFn: ({ id, is_admin }: { id: string; is_admin: boolean }) => adminApi.setAdmin(id, is_admin),
  });

  const busy = reinvite.isPending || resetPassword.isPending || setStatus.isPending || setAdmin.isPending;
  const chosen = [...picked.values()];
  const rows = users.data?.items ?? [];

  function select(ids: string[]) {
    const byId = new Map(rows.map((u) => [u.id, u]));
    setPicked((prev) => {
      const next = new Map<string, User>();
      for (const id of ids) {
        const row = byId.get(id) ?? prev.get(id);
        if (row) next.set(id, row);
      }
      return next;
    });
  }

  async function refresh() {
    setPicked(new Map());
    await qc.invalidateQueries({ queryKey: ['admin', 'users'] });
  }

  async function run(label: string, targets: User[], fn: (u: User) => Promise<unknown>) {
    if (targets.length === 0) {
      toast({ title: 'Nothing to do', body: `None of the selected residents can be ${label}.`, tone: 'info' });
      return;
    }
    const failures: string[] = [];
    for (const u of targets) {
      try {
        await fn(u);
      } catch (err) {
        failures.push(`${u.email}: ${errorMessage(err)}`);
      }
    }
    await refresh();
    if (failures.length) {
      toast({ title: 'Some updates failed', body: failures.slice(0, 3).join(' '), tone: 'danger' });
    } else {
      toast({ title: 'Updated', body: `${targets.length} resident${targets.length === 1 ? '' : 's'}.`, tone: 'success' });
    }
  }

  async function resendSelected() {
    const targets = chosen.filter((u) => u.status === 'invited');
    if (targets.length === 0) {
      toast({ title: 'Nothing to resend', body: 'Select residents who are still invited.', tone: 'info' });
      return;
    }
    const last: InviteResult[] = [];
    const failures: string[] = [];
    for (const u of targets) {
      try {
        last.push(await reinvite.mutateAsync(u.id));
      } catch (err) {
        failures.push(`${u.email}: ${errorMessage(err)}`);
      }
    }
    if (last.length === 1) setSecret({ kind: 'invite', result: last[0] });
    else if (last.length > 1) {
      toast({
        title: 'Invites resent',
        body: `${last.length} passcodes were emailed. Open each row individually if you still need the code — the latest is shown.`,
        tone: 'success',
      });
      setSecret({ kind: 'invite', result: last[last.length - 1] });
    }
    await refresh();
    if (failures.length) toast({ title: 'Some invites failed', body: failures.slice(0, 3).join(' '), tone: 'danger' });
  }

  async function resetSelected() {
    const targets = chosen.filter((u) => u.id !== me?.id && u.status !== 'invited');
    if (targets.length === 0) {
      toast({ title: 'Nothing to reset', body: 'Select registered residents other than yourself.', tone: 'info' });
      return;
    }
    if (!window.confirm(`Reset passwords for ${targets.length} resident${targets.length === 1 ? '' : 's'}? Their sessions will be signed out.`)) return;
    const done: { email: string; temporary_password: string }[] = [];
    const failures: string[] = [];
    for (const u of targets) {
      try {
        const res = await resetPassword.mutateAsync(u.id);
        done.push({ email: res.user.email, temporary_password: res.temporary_password });
      } catch (err) {
        failures.push(`${u.email}: ${errorMessage(err)}`);
      }
    }
    if (done.length) setSecret({ kind: 'reset', rows: done });
    await refresh();
    if (failures.length) toast({ title: 'Some resets failed', body: failures.slice(0, 3).join(' '), tone: 'danger' });
  }

  function submit(e: FormEvent) {
    e.preventDefault();
    invite.mutate({ email, unit_id: unitID, display_name: name || undefined });
  }

  return (
    <Screen title="Admin" wide>
      <PageHeader title="Admin" description="Invite residents and manage who can sign in." />

      <Card>
        <form onSubmit={submit}>
          <Stack gap={4}>
            <Text variant="subtitle">Invite a resident</Text>
            <ErrorBanner message={invite.isError ? errorMessage(invite.error) : null} />
            <div className="grid gap-4 md:grid-cols-3">
              <TextField label="Email" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
              <label className="flex flex-col gap-1">
                <Text variant="label" tone="muted">
                  Unit
                </Text>
                <select
                  required
                  value={unitID}
                  onChange={(e) => setUnitID(e.target.value)}
                  className="min-h-control rounded-sm border border-line bg-raised px-3 text-body md:min-h-control-dense"
                >
                  <option value="">Select a vacant unit</option>
                  {(vacantUnits.data?.items ?? []).map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.number}
                    </option>
                  ))}
                </select>
              </label>
              <TextField label="Name (optional)" maxLength={80} value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            {(vacantUnits.data?.total ?? 0) === 0 && !vacantUnits.isLoading && (
              <Text variant="caption" tone="muted">
                No vacant units. <Link to="/units" className="text-brand">Add one on Units</Link> before inviting someone.
              </Text>
            )}
            <Button type="submit" loading={invite.isPending} className="self-start">
              Send invite
            </Button>
          </Stack>
        </form>
      </Card>

      {secret?.kind === 'invite' && (
        <Card tone="pin">
          <Stack gap={2}>
            <Text variant="subtitle">Passcode for {secret.result.user.email}</Text>
            <Text variant="display" as="p" className="font-mono tracking-widest">
              {secret.result.passcode}
            </Text>
            <Text variant="caption" tone="muted">
              Give this to them in person or by text. The email link alone won't work without it. Expires{' '}
              {new Date(secret.result.expires_at).toLocaleDateString()}.
            </Text>
            {secret.result.invite_url && (
              <Text variant="caption" tone="muted" className="break-all">
                Dev link: {secret.result.invite_url}
              </Text>
            )}
          </Stack>
        </Card>
      )}

      {secret?.kind === 'reset' && (
        <Card tone="pin">
          <Stack gap={3}>
            <Text variant="subtitle">Temporary passwords</Text>
            <Text variant="caption" tone="muted">
              Give these out of band. Those sessions were signed out.
            </Text>
            {secret.rows.map((row) => (
              <div key={row.email}>
                <Text variant="caption" tone="muted">
                  {row.email}
                </Text>
                <Text variant="subtitle" as="p" className="font-mono">
                  {row.temporary_password}
                </Text>
              </div>
            ))}
          </Stack>
        </Card>
      )}

      <Stack gap={3}>
        <Text variant="subtitle">Residents</Text>
        {users.isError && <ErrorBanner message={errorMessage(users.error)} />}
        <DataTable
          columns={columns}
          rows={rows}
          total={users.data?.total ?? 0}
          query={query}
          onQueryChange={setQuery}
          rowId={(u) => u.id}
          selected={[...picked.keys()]}
          onSelectedChange={select}
          loading={users.isFetching}
          empty="No residents match."
          actions={
            chosen.length > 0 ? (
              <div className="flex flex-wrap items-center gap-2">
                <Text variant="caption" tone="muted">
                  {chosen.length} selected
                </Text>
                <Button size="sm" variant="secondary" disabled={busy} onClick={() => void resetSelected()}>
                  Reset password
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  disabled={busy}
                  onClick={() => void resendSelected()}
                >
                  Resend invite
                </Button>
                <Button
                  size="sm"
                  disabled={busy}
                  onClick={() =>
                    void run(
                      'activated',
                      chosen.filter((u) => u.status === 'inactive'),
                      (u) => setStatus.mutateAsync({ id: u.id, active: true }),
                    )
                  }
                >
                  Activate
                </Button>
                <Button
                  size="sm"
                  variant="danger"
                  disabled={busy}
                  onClick={() => {
                    const targets = chosen.filter((u) => u.status === 'active' && u.id !== me?.id);
                    if (targets.length === 0) {
                      toast({ title: 'Nothing to deactivate', body: 'Select active residents other than yourself.', tone: 'info' });
                      return;
                    }
                    if (!window.confirm(`Deactivate ${targets.length} resident${targets.length === 1 ? '' : 's'}?`)) return;
                    void run('deactivated', targets, (u) => setStatus.mutateAsync({ id: u.id, active: false }));
                  }}
                >
                  Deactivate
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  disabled={busy}
                  onClick={() =>
                    void run(
                      'made admin',
                      chosen.filter((u) => !u.is_admin && u.status !== 'invited'),
                      (u) => setAdmin.mutateAsync({ id: u.id, is_admin: true }),
                    )
                  }
                >
                  Make admin
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  disabled={busy}
                  onClick={() =>
                    void run(
                      'removed as admin',
                      chosen.filter((u) => u.is_admin && u.id !== me?.id),
                      (u) => setAdmin.mutateAsync({ id: u.id, is_admin: false }),
                    )
                  }
                >
                  Remove admin
                </Button>
              </div>
            ) : null
          }
        />
      </Stack>
    </Screen>
  );
}

function StatusBadge({ status }: { status: User['status'] }) {
  if (status === 'active') return <Badge tone="success">Active</Badge>;
  if (status === 'inactive') return <Badge tone="danger">Inactive</Badge>;
  return <Badge tone="neutral">Invited</Badge>;
}
