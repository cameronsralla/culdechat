import { useEffect, useId, useRef, useState, type ReactNode } from 'react';
import { cn } from '@/lib/cn';
import { Button } from './Button';
import { Icon } from './Icon';
import { Text } from './Text';
import { TextField } from './TextField';

export type SortDir = 'asc' | 'desc';

export type DataTableQuery = {
  q: string;
  sortKey: string;
  sortDir: SortDir;
  page: number;
  pageSize: number;
  filters: Record<string, string>;
};

export type DataTableColumn<T> = {
  key: string;
  header: string;
  sortable?: boolean;
  filter?: 'text' | 'select';
  options?: { value: string; label: string }[];
  render: (row: T) => ReactNode;
  className?: string;
};

type Props<T> = {
  columns: DataTableColumn<T>[];
  rows: T[];
  total: number;
  query: DataTableQuery;
  onQueryChange: (next: DataTableQuery) => void;
  rowId: (row: T) => string;
  /** When false, the checkbox column is omitted. Defaults to true. */
  selectable?: boolean;
  selected?: string[];
  onSelectedChange?: (ids: string[]) => void;
  loading?: boolean;
  /** Shown above the table when the parent has a selection. */
  actions?: ReactNode;
  empty?: string;
};

const PAGE_SIZES = [10, 25, 50];

/**
 * Server-backed table. Search, sort, column filters, and paging are reported
 * upward; this component does not filter the current page locally.
 */
export function DataTable<T>({
  columns,
  rows,
  total,
  query,
  onQueryChange,
  rowId,
  selectable = true,
  selected = [],
  onSelectedChange,
  loading,
  actions,
  empty = 'No matching rows.',
}: Props<T>) {
  const [draftQ, setDraftQ] = useState(query.q);
  const [menu, setMenu] = useState<{ key: string; x: number; y: number } | null>(null);

  useEffect(() => {
    setDraftQ(query.q);
  }, [query.q]);

  const queryRef = useRef(query);
  queryRef.current = query;

  useEffect(() => {
    const handle = window.setTimeout(() => {
      const current = queryRef.current;
      if (draftQ !== current.q) onQueryChange({ ...current, q: draftQ, page: 1 });
    }, 250);
    return () => window.clearTimeout(handle);
  }, [draftQ, onQueryChange]);

  const visibleIds = rows.map(rowId);
  const allVisible = visibleIds.length > 0 && visibleIds.every((id) => selected.includes(id));
  const from = total === 0 ? 0 : (query.page - 1) * query.pageSize + 1;
  const to = Math.min(query.page * query.pageSize, total);
  const pages = Math.max(1, Math.ceil(total / query.pageSize));

  function toggleSort(key: string) {
    if (query.sortKey === key) {
      onQueryChange({ ...query, sortDir: query.sortDir === 'asc' ? 'desc' : 'asc', page: 1 });
    } else {
      onQueryChange({ ...query, sortKey: key, sortDir: 'asc', page: 1 });
    }
  }

  function setFilter(key: string, value: string) {
    const filters = { ...query.filters };
    if (value) filters[key] = value;
    else delete filters[key];
    onQueryChange({ ...query, filters, page: 1 });
    setMenu(null);
  }

  function toggleRow(id: string) {
    if (!onSelectedChange) return;
    onSelectedChange(selected.includes(id) ? selected.filter((x) => x !== id) : [...selected, id]);
  }

  function togglePage() {
    if (!onSelectedChange) return;
    if (allVisible) onSelectedChange(selected.filter((id) => !visibleIds.includes(id)));
    else onSelectedChange([...new Set([...selected, ...visibleIds])]);
  }

  return (
    <div className="flex flex-col gap-3">
      <TextField
        label="Search"
        placeholder="Search all results"
        value={draftQ}
        onChange={(e) => setDraftQ(e.target.value)}
      />

      {actions}

      <div className="overflow-x-auto rounded-md border border-line bg-surface">
        <table className="w-full border-collapse text-left">
          <thead>
            <tr className="border-b border-line bg-surface-muted">
              {selectable && (
                <th className="w-10 px-3 py-2">
                  <input
                    type="checkbox"
                    aria-label="Select all on this page"
                    checked={allVisible}
                    onChange={togglePage}
                    className="size-4 accent-brand"
                  />
                </th>
              )}
              {columns.map((col) => (
                <th key={col.key} className={cn('px-3 py-2 font-medium', col.className)}>
                  <div className="group flex items-center gap-1">
                    {col.sortable ? (
                      <button
                        type="button"
                        className="inline-flex items-center gap-1 text-caption text-muted hover:text-ink"
                        onClick={() => toggleSort(col.key)}
                        onContextMenu={(e) => {
                          if (!col.filter) return;
                          e.preventDefault();
                          setMenu({ key: col.key, x: e.clientX, y: e.clientY });
                        }}
                      >
                        {col.header}
                        {query.sortKey === col.key && (
                          <span aria-hidden className="text-brand">
                            {query.sortDir === 'asc' ? '↑' : '↓'}
                          </span>
                        )}
                      </button>
                    ) : (
                      <span className="text-caption text-muted">{col.header}</span>
                    )}
                    {col.filter && (
                      <button
                        type="button"
                        aria-label={`Filter ${col.header}`}
                        className={cn(
                          'rounded-sm p-0.5 text-muted hover:bg-raised hover:text-ink',
                          query.filters[col.key] ? 'opacity-100 text-brand' : 'opacity-0 group-hover:opacity-100',
                        )}
                        onClick={(e) => {
                          const rect = e.currentTarget.getBoundingClientRect();
                          setMenu({ key: col.key, x: rect.left, y: rect.bottom + 4 });
                        }}
                      >
                        <Icon name="filter" size={14} />
                      </button>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className={cn(loading && 'opacity-60')}>
            {rows.length === 0 && (
              <tr>
                <td colSpan={columns.length + (selectable ? 1 : 0)} className="px-3 py-8 text-center">
                  <Text variant="caption" tone="muted">
                    {empty}
                  </Text>
                </td>
              </tr>
            )}
            {rows.map((row) => {
              const id = rowId(row);
              return (
                <tr key={id} className="border-b border-line-soft last:border-0">
                  {selectable && (
                    <td className="px-3 py-2.5">
                      <input
                        type="checkbox"
                        aria-label={`Select row ${id}`}
                        checked={selected.includes(id)}
                        onChange={() => toggleRow(id)}
                        className="size-4 accent-brand"
                      />
                    </td>
                  )}
                  {columns.map((col) => (
                    <td key={col.key} className={cn('px-3 py-2.5 align-middle', col.className)}>
                      {col.render(row)}
                    </td>
                  ))}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <Text variant="caption" tone="muted">
          {total === 0 ? '0 results' : `${from}–${to} of ${total}`}
        </Text>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-2">
            <Text variant="caption" tone="muted">
              Rows
            </Text>
            <select
              value={query.pageSize}
              onChange={(e) => onQueryChange({ ...query, pageSize: Number(e.target.value), page: 1 })}
              className="min-h-control-sm rounded-sm border border-line bg-raised px-2 text-caption md:min-h-control-dense-sm"
            >
              {PAGE_SIZES.map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </label>
          <Button size="sm" variant="secondary" disabled={query.page <= 1} onClick={() => onQueryChange({ ...query, page: query.page - 1 })}>
            Previous
          </Button>
          <Text variant="caption" tone="muted">
            {query.page} / {pages}
          </Text>
          <Button size="sm" variant="secondary" disabled={query.page >= pages} onClick={() => onQueryChange({ ...query, page: query.page + 1 })}>
            Next
          </Button>
        </div>
      </div>

      {menu && (
        <FilterMenu
          header={columns.find((c) => c.key === menu.key)?.header ?? 'Column'}
          filter={columns.find((c) => c.key === menu.key)?.filter}
          options={columns.find((c) => c.key === menu.key)?.options}
          value={query.filters[menu.key] ?? ''}
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          onApply={(value) => setFilter(menu.key, value)}
        />
      )}
    </div>
  );
}

function FilterMenu({
  header,
  filter,
  options,
  value,
  x,
  y,
  onClose,
  onApply,
}: {
  header: string;
  filter?: 'text' | 'select';
  options?: { value: string; label: string }[];
  value: string;
  x: number;
  y: number;
  onClose: () => void;
  onApply: (value: string) => void;
}) {
  const id = useId();
  const [draft, setDraft] = useState(value);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function onDoc(e: MouseEvent) {
      if (!ref.current?.contains(e.target as Node)) onClose();
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    document.addEventListener('mousedown', onDoc);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDoc);
      document.removeEventListener('keydown', onKey);
    };
  }, [onClose]);

  return (
    <div
      ref={ref}
      role="dialog"
      aria-labelledby={id}
      className="fixed z-50 w-64 rounded-md border border-line bg-raised p-3 shadow-raised"
      style={{ left: Math.min(x, window.innerWidth - 280), top: Math.min(y, window.innerHeight - 220) }}
    >
      <Text id={id} variant="label" tone="muted">
        Filter {header}
      </Text>
      {filter === 'select' ? (
        <div className="mt-2 flex flex-col gap-1">
          <FilterChoice active={value === ''} label="Any" onClick={() => onApply('')} />
          {options?.map((opt) => (
            <FilterChoice key={opt.value} active={value === opt.value} label={opt.label} onClick={() => onApply(opt.value)} />
          ))}
        </div>
      ) : (
        <form
          className="mt-2 flex flex-col gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            onApply(draft.trim());
          }}
        >
          <TextField label="Contains" value={draft} onChange={(e) => setDraft(e.target.value)} autoFocus />
          <div className="flex gap-2">
            <Button type="submit" size="sm">
              Apply
            </Button>
            <Button type="button" size="sm" variant="ghost" onClick={() => onApply('')}>
              Clear
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}

function FilterChoice({ active, label, onClick }: { active: boolean; label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'rounded-sm px-2 py-1.5 text-left text-caption hover:bg-surface-muted',
        active && 'bg-brand-wash text-ink',
      )}
    >
      {label}
    </button>
  );
}
