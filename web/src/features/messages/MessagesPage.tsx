import { useEffect, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '@/api/client';
import { messagesApi, usersApi } from '@/api/endpoints';
import type { Conversation, DirectoryEntry, MessagePeer } from '@/api/types';
import { useAuth } from '@/auth/AuthContext';
import { Avatar, Button, ErrorBanner, Icon, Text, errorMessage } from '@/components/ui';
import { cn } from '@/lib/cn';

type Draft = Pick<DirectoryEntry, 'kind' | 'id' | 'unit_number' | 'display_name' | 'email'>;

type Open = { mode: 'none' } | { mode: 'draft'; target: Draft } | { mode: 'thread'; id: string };

function peerLabel(peer: MessagePeer) {
  if (peer.listed && peer.display_name) return peer.display_name;
  return `Unit ${peer.unit_number}`;
}

function targetLabel(row: Draft) {
  if (row.kind === 'unit') return `Unit ${row.unit_number}`;
  return row.display_name || row.email || `Unit ${row.unit_number}`;
}

function rowKey(row: Draft) {
  return row.kind === 'person' && row.id ? row.id : `unit:${row.unit_number}`;
}

function findPerson(items: Conversation[], target: Draft) {
  if (target.kind === 'person' && target.id) return items.find((c) => c.peer.id === target.id);
  return undefined;
}

function preview(item: Conversation) {
  if (item.status === 'pending' && item.incoming) return 'Wants to chat';
  if (item.status === 'pending' && item.requested_by_me) return 'Waiting for them to accept';
  if (item.status === 'declined' && item.requested_by_me) return 'They declined';
  if (item.status === 'declined') return 'You declined';
  return item.last_message ?? '';
}

function clock(iso: string) {
  return new Date(iso).toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
}

function listTime(iso: string) {
  const d = new Date(iso);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) return clock(iso);
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' });
}

function dayLabel(iso: string) {
  const d = new Date(iso);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) return 'Today';
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (d.toDateString() === yesterday.toDateString()) return 'Yesterday';
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' });
}

export function ChatPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const qc = useQueryClient();
  const start = (location.state as { start?: Draft } | null)?.start ?? null;
  const [open, setOpen] = useState<Open>({ mode: 'none' });
  const [composing, setComposing] = useState(false);
  const [search, setSearch] = useState('');
  const [q, setQ] = useState('');
  const [text, setText] = useState('');
  const [err, setErr] = useState<string | null>(null);
  const scroller = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handle = window.setTimeout(() => setQ(search.trim()), 250);
    return () => window.clearTimeout(handle);
  }, [search]);

  const list = useQuery({
    queryKey: ['messages'],
    queryFn: messagesApi.list,
    refetchInterval: 15000,
  });
  const threadId = open.mode === 'thread' ? open.id : null;
  const detail = useQuery({
    queryKey: ['messages', threadId],
    queryFn: () => messagesApi.get(threadId!),
    enabled: !!threadId,
    refetchInterval: 15000,
  });
  const hits = useQuery({
    queryKey: ['chat', 'targets', q],
    queryFn: () => usersApi.directory({ q, page: 1, pageSize: 12 }),
    enabled: composing && q.length > 0,
  });

  const items = list.data ?? [];

  const started = useRef<Draft | null>(null);
  useEffect(() => {
    if (!start || list.isLoading || started.current === start) return;
    started.current = start;
    setComposing(false);
    navigate('/chat', { replace: true, state: null });
    const person = findPerson(items, start);
    if (start.kind === 'person') {
      setOpen(person ? { mode: 'thread', id: person.id } : { mode: 'draft', target: start });
      return;
    }
    void messagesApi.withUnit(start.unit_number).then(
      (conv) => setOpen({ mode: 'thread', id: conv.id }),
      (e) => {
        if (e instanceof ApiError && e.status === 404) setOpen({ mode: 'draft', target: start });
        else setErr(errorMessage(e));
      },
    );
  }, [start, list.isLoading, items, navigate]);

  useEffect(() => {
    const el = scroller.current;
    if (!el) return;
    el.scrollTo({ top: el.scrollHeight });
  }, [detail.data?.messages?.length, threadId, open.mode]);

  const send = useMutation({
    mutationFn: messagesApi.send,
    onSuccess: (res) => {
      setText('');
      setErr(null);
      setOpen({ mode: 'thread', id: res.id });
      void qc.invalidateQueries({ queryKey: ['messages'] });
    },
    onError: (e) => setErr(errorMessage(e)),
  });

  const decide = useMutation({
    mutationFn: (input: { id: string; accept: boolean }) =>
      input.accept ? messagesApi.accept(input.id) : messagesApi.decline(input.id),
    onSuccess: () => {
      setErr(null);
      void qc.invalidateQueries({ queryKey: ['messages'] });
    },
    onError: (e) => setErr(errorMessage(e)),
  });

  const thread = open.mode === 'thread' ? detail.data : undefined;
  const draft = open.mode === 'draft' ? open.target : undefined;
  const title = thread ? peerLabel(thread.peer) : draft ? targetLabel(draft) : '';
  const unit = thread?.peer.unit_number ?? draft?.unit_number ?? '';
  const matches = (hits.data?.items ?? []).filter((row) => !row.self && row.id !== user?.id);
  const requests = items.filter((c) => c.status === 'pending' && c.incoming);
  const chats = items.filter((c) => !(c.status === 'pending' && c.incoming));
  const showThread = open.mode !== 'none';

  function choose(target: Draft) {
    setErr(null);
    setText('');
    setComposing(false);
    setSearch('');
    setQ('');
    const person = findPerson(items, target);
    if (target.kind === 'person') {
      setOpen(person ? { mode: 'thread', id: person.id } : { mode: 'draft', target });
      return;
    }
    void messagesApi.withUnit(target.unit_number).then(
      (conv) => setOpen({ mode: 'thread', id: conv.id }),
      (e) => {
        if (e instanceof ApiError && e.status === 404) setOpen({ mode: 'draft', target });
        else setErr(errorMessage(e));
      },
    );
  }

  function submit() {
    const content = text.trim();
    if (!content || send.isPending) return;
    if (thread?.status === 'open') {
      send.mutate({ content, conversation_id: thread.id });
      return;
    }
    if (thread?.status === 'declined' && thread.requested_by_me) {
      send.mutate({ content, conversation_id: thread.id });
      return;
    }
    if (draft) {
      send.mutate(draft.kind === 'person' && draft.id ? { content, user_id: draft.id } : { content, unit_number: draft.unit_number });
    }
  }

  const canCompose = !!draft || thread?.status === 'open' || (thread?.status === 'declined' && thread.requested_by_me);
  const hint = draft?.kind === 'unit'
    ? "They aren't listed. They can accept or decline."
    : thread?.status === 'declined' && thread.requested_by_me
      ? 'They declined. A new message asks again.'
      : null;

  return (
    <div className="flex h-dvh overflow-hidden bg-paper pb-[calc(var(--spacing-tabbar)+env(safe-area-inset-bottom))] md:pb-0">
      <aside className={cn('flex min-w-0 flex-1 flex-col border-line bg-surface md:w-80 md:flex-none md:border-r', showThread && 'hidden md:flex')}>
        {composing ? (
          <div className="flex h-14 shrink-0 items-center gap-1 border-b border-line px-2">
            <button type="button" aria-label="Cancel" onClick={() => setComposing(false)} className="flex size-10 items-center justify-center rounded-sm text-ink hover:bg-surface-muted">
              <Icon name="chevronLeft" size={20} />
            </button>
            <input
              autoFocus
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Name, email, or unit"
              className="min-w-0 flex-1 bg-transparent text-body text-ink outline-none placeholder:text-muted/70"
            />
          </div>
        ) : (
          <div className="flex h-14 shrink-0 items-center justify-between border-b border-line px-4">
            <Text variant="subtitle">Chat</Text>
            <button type="button" aria-label="New chat" onClick={() => { setComposing(true); setErr(null); }} className="flex size-10 items-center justify-center rounded-sm text-brand hover:bg-brand-wash">
              <Icon name="plus" size={20} />
            </button>
          </div>
        )}

        <div className="min-h-0 flex-1 overflow-y-auto">
          {composing ? (
            <SearchResults q={q} loading={hits.isFetching} error={hits.isError ? errorMessage(hits.error) : null} matches={matches} total={hits.data?.total ?? 0} onPick={choose} />
          ) : (
            <ConversationList
              loading={list.isLoading}
              error={list.isError ? errorMessage(list.error) : null}
              requests={requests}
              chats={chats}
              activeId={threadId}
              onOpen={(id) => {
                setErr(null);
                setText('');
                setOpen({ mode: 'thread', id });
              }}
            />
          )}
        </div>
      </aside>

      <section className={cn('min-w-0 flex-1 flex-col bg-paper', showThread ? 'flex' : 'hidden md:flex')}>
        {showThread ? (
          <>
            <header className="flex h-14 shrink-0 items-center gap-2 border-b border-line bg-surface px-2 md:px-4">
              <button type="button" aria-label="Back" onClick={() => setOpen({ mode: 'none' })} className="flex size-10 items-center justify-center rounded-sm text-ink hover:bg-surface-muted md:hidden">
                <Icon name="chevronLeft" size={20} />
              </button>
              <Avatar name={title} size="sm" />
              <div className="min-w-0">
                <Text variant="subtitle" className="truncate">{title}</Text>
                <Text variant="caption" as="p" tone="muted">Unit {unit}</Text>
              </div>
            </header>
            {(err || detail.isError) && (
              <div className="px-4 pt-3">
                <ErrorBanner message={err ?? errorMessage(detail.error)} />
              </div>
            )}
            <div ref={scroller} className="min-h-0 flex-1 overflow-y-auto px-4 py-4">
              <div className="flex min-h-full flex-col justify-end gap-2">
                {draft && <EmptyThread label={title} />}
                {thread && <Messages messages={thread.messages ?? []} />}
                {open.mode === 'thread' && detail.isLoading && (
                  <Text variant="caption" tone="muted">Loading…</Text>
                )}
              </div>
            </div>
            {thread?.incoming && thread.status === 'pending' && (
              <div className="flex items-center justify-between gap-3 border-t border-line bg-surface px-4 py-3">
                <Text variant="caption" tone="muted">Accept to reply</Text>
                <div className="flex gap-2">
                  <Button size="sm" variant="secondary" loading={decide.isPending} onClick={() => decide.mutate({ id: thread.id, accept: false })}>Decline</Button>
                  <Button size="sm" loading={decide.isPending} onClick={() => decide.mutate({ id: thread.id, accept: true })}>Accept</Button>
                </div>
              </div>
            )}
            {thread?.status === 'pending' && thread.requested_by_me && (
              <div className="border-t border-line bg-surface px-4 py-3 text-center">
                <Text variant="caption" tone="muted">Waiting for them to accept</Text>
              </div>
            )}
            {thread?.status === 'declined' && !thread.requested_by_me && (
              <div className="border-t border-line bg-surface px-4 py-3 text-center">
                <Text variant="caption" tone="muted">You declined this chat</Text>
              </div>
            )}
            {canCompose && <Composer value={text} onChange={setText} onSend={submit} disabled={send.isPending} hint={hint} />}
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center px-6">
            <Text variant="body" tone="muted">Select a conversation, or start a new one.</Text>
          </div>
        )}
      </section>
    </div>
  );
}

function SearchResults({
  q,
  loading,
  error,
  matches,
  total,
  onPick,
}: {
  q: string;
  loading: boolean;
  error: string | null;
  matches: DirectoryEntry[];
  total: number;
  onPick: (row: DirectoryEntry) => void;
}) {
  if (!q) {
    return <p className="px-4 py-6 text-caption text-muted">Search for a neighbor or a unit.</p>;
  }
  if (error) return <div className="p-3"><ErrorBanner message={error} /></div>;
  if (!loading && matches.length === 0) return <p className="px-4 py-6 text-caption text-muted">No one matches.</p>;
  return (
    <div>
      {matches.map((row) => (
        <button key={rowKey(row)} type="button" onClick={() => onPick(row)} className="flex w-full items-center gap-3 px-4 py-2.5 text-left hover:bg-surface-muted">
          <Avatar name={targetLabel(row)} size="sm" />
          <span className="min-w-0">
            <Text variant="body" className="truncate">{targetLabel(row)}</Text>
            <Text variant="caption" as="p" tone="muted" className="truncate">
              {row.kind === 'unit' ? 'Not listed' : `Unit ${row.unit_number}`}
            </Text>
          </span>
        </button>
      ))}
      {total > matches.length && <p className="px-4 py-2 text-caption text-muted">Keep typing to narrow {total} matches.</p>}
    </div>
  );
}

function ConversationList({
  loading,
  error,
  requests,
  chats,
  activeId,
  onOpen,
}: {
  loading: boolean;
  error: string | null;
  requests: Conversation[];
  chats: Conversation[];
  activeId: string | null;
  onOpen: (id: string) => void;
}) {
  if (error) return <div className="p-3"><ErrorBanner message={error} /></div>;
  if (loading) return <p className="px-4 py-6 text-caption text-muted">Loading…</p>;
  if (requests.length === 0 && chats.length === 0) {
    return <p className="px-4 py-6 text-caption text-muted">No chats yet. Start one with +.</p>;
  }
  return (
    <div className="py-1">
      {requests.length > 0 && <Group label="Requests" items={requests} activeId={activeId} onOpen={onOpen} />}
      {chats.length > 0 && <Group label={requests.length > 0 ? 'Chats' : undefined} items={chats} activeId={activeId} onOpen={onOpen} />}
    </div>
  );
}

function Group({ label, items, activeId, onOpen }: { label?: string; items: Conversation[]; activeId: string | null; onOpen: (id: string) => void }) {
  return (
    <div>
      {label && <p className="px-4 pb-1 pt-3 text-label font-semibold text-muted">{label}</p>}
      {items.map((item) => (
        <button
          key={item.id}
          type="button"
          onClick={() => onOpen(item.id)}
          className={cn('flex w-full items-center gap-3 px-3 py-2.5 text-left', item.id === activeId ? 'bg-brand-wash' : 'hover:bg-surface-muted')}
        >
          <Avatar name={peerLabel(item.peer)} size="sm" />
          <span className="min-w-0 flex-1">
            <span className="flex items-baseline justify-between gap-2">
              <Text variant="body" className="truncate">{peerLabel(item.peer)}</Text>
              <Text variant="caption" tone="muted" className="shrink-0">{listTime(item.updated_at)}</Text>
            </span>
            <Text variant="caption" as="p" tone="muted" className="truncate">{preview(item)}</Text>
          </span>
        </button>
      ))}
    </div>
  );
}

function Messages({ messages }: { messages: { id: string; mine: boolean; body: string; created_at: string }[] }) {
  let lastDay = '';
  return (
    <>
      {messages.map((m) => {
        const day = dayLabel(m.created_at);
        const showDay = day !== lastDay;
        lastDay = day;
        return (
          <div key={m.id} className="flex flex-col gap-2">
            {showDay && <p className="py-1 text-center text-caption text-muted">{day}</p>}
            <div className={cn('flex', m.mine ? 'justify-end' : 'justify-start')}>
              <div className={cn('max-w-[75%] rounded-lg px-3 py-2', m.mine ? 'rounded-br-sm bg-brand text-white' : 'rounded-bl-sm border border-line bg-surface text-ink')}>
                <p className="whitespace-pre-wrap text-body">{m.body}</p>
                <p className={cn('mt-0.5 text-right text-caption', m.mine ? 'text-white/75' : 'text-muted')}>{clock(m.created_at)}</p>
              </div>
            </div>
          </div>
        );
      })}
    </>
  );
}

function EmptyThread({ label }: { label: string }) {
  return (
    <div className="flex flex-1 items-center justify-center">
      <Text variant="body" tone="muted">Say hello to {label}</Text>
    </div>
  );
}

function Composer({
  value,
  onChange,
  onSend,
  disabled,
  hint,
}: {
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  disabled: boolean;
  hint: string | null;
}) {
  const ref = useRef<HTMLTextAreaElement>(null);

  function resize(el: HTMLTextAreaElement) {
    el.style.height = '0px';
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`;
  }

  useEffect(() => {
    ref.current?.focus();
  }, []);

  return (
    <form
      className="shrink-0 border-t border-line bg-surface px-3 py-2"
      onSubmit={(e) => {
        e.preventDefault();
        onSend();
      }}
    >
      {hint && <p className="px-1 pb-1 text-caption text-muted">{hint}</p>}
      <div className="flex items-end gap-2">
        <textarea
          ref={ref}
          rows={1}
          value={value}
          maxLength={2000}
          placeholder="Message"
          onChange={(e) => {
            onChange(e.target.value);
            resize(e.target);
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              onSend();
            }
          }}
          className="max-h-40 min-h-control flex-1 resize-none rounded-pill border border-line bg-raised px-4 py-2 text-body text-ink outline-none placeholder:text-muted/70 focus:border-brand md:min-h-control-dense"
        />
        <button
          type="submit"
          aria-label="Send"
          disabled={disabled || value.trim() === ''}
          className="flex size-10 shrink-0 items-center justify-center rounded-pill bg-brand text-white hover:bg-brand-dark disabled:opacity-40"
        >
          <Icon name="send" size={18} />
        </button>
      </div>
    </form>
  );
}
