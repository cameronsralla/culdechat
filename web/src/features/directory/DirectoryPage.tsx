import { useState } from 'react';
import { useNavigate } from 'react-router';
import { useQuery } from '@tanstack/react-query';
import { usersApi } from '@/api/endpoints';
import type { DirectoryEntry } from '@/api/types';
import { useAuth } from '@/auth/AuthContext';
import {
  Avatar,
  Button,
  DataTable,
  ErrorBanner,
  PageHeader,
  errorMessage,
  type DataTableColumn,
  type DataTableQuery,
} from '@/components/ui';
import { Screen } from '@/layout/Screen';

function rowKey(p: DirectoryEntry) {
  return p.kind === 'person' && p.id ? p.id : `unit:${p.unit_number}`;
}

export function DirectoryPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [query, setQuery] = useState<DataTableQuery>({
    q: '',
    sortKey: 'unit_number',
    sortDir: 'asc',
    page: 1,
    pageSize: 10,
    filters: {},
  });
  const people = useQuery({
    queryKey: ['directory', query],
    queryFn: () =>
      usersApi.directory({
        q: query.q,
        sort: query.sortKey,
        dir: query.sortDir,
        page: query.page,
        pageSize: query.pageSize,
        name: query.filters.display_name,
        email: query.filters.email,
        unit: query.filters.unit_number,
      }),
  });

  const columns: DataTableColumn<DirectoryEntry>[] = [
    {
      key: 'display_name',
      header: 'Name',
      sortable: true,
      filter: 'text',
      render: (p) => (
        <span className="inline-flex items-center gap-2">
          <Avatar name={p.kind === 'unit' ? `Unit ${p.unit_number}` : p.display_name || p.email || 'Neighbor'} size="sm" />
          {p.kind === 'unit' ? 'Not listed' : p.display_name || '—'}
        </span>
      ),
    },
    {
      key: 'email',
      header: 'Email',
      sortable: true,
      filter: 'text',
      render: (p) => p.email || '—',
    },
    {
      key: 'unit_number',
      header: 'Unit',
      sortable: true,
      filter: 'text',
      className: 'font-mono',
      render: (p) => p.unit_number,
    },
    {
      key: 'message',
      header: '',
      render: (p) =>
        p.self || (p.id && p.id === user?.id) ? null : (
          <Button
            size="sm"
            variant="secondary"
            onClick={() => navigate('/chat', { state: { start: p } })}
          >
            Message
          </Button>
        ),
    },
  ];

  const filtered = query.q !== '' || Object.keys(query.filters).length > 0;

  return (
    <Screen title="People">
      <PageHeader title="People" description="Listed neighbors, and units where someone lives but isn't listed." />
      {people.isError && <ErrorBanner message={errorMessage(people.error)} />}
      <DataTable
        columns={columns}
        rows={people.data?.items ?? []}
        total={people.data?.total ?? 0}
        query={query}
        onQueryChange={setQuery}
        rowId={rowKey}
        selectable={false}
        loading={people.isFetching}
        empty={filtered ? 'No matching people or units.' : 'No one is active yet.'}
      />
    </Screen>
  );
}
