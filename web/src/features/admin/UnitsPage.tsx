import { useState, type FormEvent } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/api/endpoints';
import type { Unit } from '@/api/types';
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

export function UnitsPage() {
  const qc = useQueryClient();
  const toast = useToast();
  const [number, setNumber] = useState('');
  const [query, setQuery] = useState<DataTableQuery>({
    q: '',
    sortKey: 'number',
    sortDir: 'asc',
    page: 1,
    pageSize: 10,
    filters: {},
  });

  const units = useQuery({
    queryKey: ['admin', 'units', query],
    queryFn: () =>
      adminApi.units({
        q: query.filters.number || query.q,
        dir: query.sortDir,
        page: query.page,
        pageSize: query.pageSize,
      }),
  });

  const createUnit = useMutation({
    mutationFn: adminApi.createUnit,
    onSuccess: () => {
      setNumber('');
      void qc.invalidateQueries({ queryKey: ['admin', 'units'] });
    },
  });
  const deleteUnit = useMutation({
    mutationFn: adminApi.deleteUnit,
    onSuccess: () => {
      toast({ title: 'Unit deleted', tone: 'success' });
      void qc.invalidateQueries({ queryKey: ['admin', 'units'] });
    },
    onError: (err) => toast({ title: 'Could not delete', body: errorMessage(err), tone: 'danger' }),
  });

  function addUnit(e: FormEvent) {
    e.preventDefault();
    createUnit.mutate(number);
  }

  const columns: DataTableColumn<Unit>[] = [
    {
      key: 'number',
      header: 'Unit',
      sortable: true,
      filter: 'text',
      className: 'font-mono',
      render: (u) => u.number,
    },
    {
      key: 'primary',
      header: 'Primary',
      render: (u) => {
        const primary = u.residents.find((r) => r.is_primary);
        return primary?.display_name || primary?.email || 'Vacant';
      },
    },
    {
      key: 'email',
      header: 'Email',
      render: (u) => u.residents.find((r) => r.is_primary)?.email || '—',
    },
    {
      key: 'status',
      header: 'Status',
      render: (u) => {
        const primary = u.residents.find((r) => r.is_primary);
        if (!primary) return 'Vacant';
        if (primary.status === 'active') return <Badge tone="success">Active</Badge>;
        if (primary.status === 'inactive') return <Badge tone="danger">Inactive</Badge>;
        return <Badge tone="neutral">Invited</Badge>;
      },
    },
    {
      key: 'delete',
      header: '',
      render: (u) => (
        <Button
          size="sm"
          variant="ghost"
          loading={deleteUnit.isPending && deleteUnit.variables === u.id}
          onClick={() => {
            if (window.confirm(`Delete unit ${u.number}?`)) deleteUnit.mutate(u.id);
          }}
        >
          Delete
        </Button>
      ),
    },
  ];

  return (
    <Screen title="Units" wide>
      <PageHeader title="Units" description="Define units, remove empty ones, and see who is the primary resident." />
      <Card>
        <form onSubmit={addUnit}>
          <Stack gap={4}>
            <Text variant="subtitle">Add a unit</Text>
            <ErrorBanner message={createUnit.isError ? errorMessage(createUnit.error) : null} />
            <div className="flex flex-wrap items-end gap-3">
              <TextField label="Unit number" required maxLength={32} value={number} onChange={(e) => setNumber(e.target.value)} className="w-48" />
              <Button type="submit" loading={createUnit.isPending}>
                Add unit
              </Button>
            </div>
          </Stack>
        </form>
      </Card>
      {units.isError && <ErrorBanner message={errorMessage(units.error)} />}
      <DataTable
        columns={columns}
        rows={units.data?.items ?? []}
        total={units.data?.total ?? 0}
        query={query}
        onQueryChange={setQuery}
        rowId={(u) => u.id}
        selectable={false}
        loading={units.isFetching}
        empty="No units yet."
      />
    </Screen>
  );
}
