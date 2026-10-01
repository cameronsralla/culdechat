import { useCallback, useEffect, useMemo, useState } from 'react';
import { View } from 'react-native';
import { useFocusEffect, useRouter } from 'expo-router';
import { listConversations, listDirectory, searchRecipients } from '../api/community';
import { peerLabel, type ConversationSummary, type DirectoryUser, type MessageRecipient } from '../api/types';
import {
  AppText,
  Avatar,
  EmptyState,
  ErrorBanner,
  Fab,
  ListGroup,
  ListRow,
  PageHeader,
  Stack,
  TextField,
} from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useCompactLayout } from '../components/layout/useCompactLayout';
import { relativeTime } from '../lib/time';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  wrap: { flex: 1 },
  preview: { marginTop: 2 },
  meta: {
    flexDirection: 'row' as const,
    justifyContent: 'space-between' as const,
    gap: t.space.sm,
    alignItems: 'center' as const,
  },
  section: { marginTop: t.space.xs },
});

function openCompose(
  router: ReturnType<typeof useRouter>,
  target: { userId?: string; unit: string; name?: string | null },
) {
  router.push({
    pathname: '/messages/new',
    params: {
      userId: target.userId ?? '',
      unit: target.unit,
      name: target.name ?? '',
    },
  });
}

export function MessagesScreen() {
  const router = useRouter();
  const compact = useCompactLayout();
  const styles = useStyles(stylesFor);
  const [items, setItems] = useState<ConversationSummary[]>([]);
  const [directory, setDirectory] = useState<DirectoryUser[]>([]);
  const [query, setQuery] = useState('');
  const [hits, setHits] = useState<MessageRecipient[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [searching, setSearching] = useState(false);

  const load = useCallback(async () => {
    try {
      const [conversations, people] = await Promise.all([listConversations(), listDirectory()]);
      setItems(conversations);
      setDirectory(people);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load messages');
    } finally {
      setLoaded(true);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      void load();
    }, [load]),
  );

  useEffect(() => {
    const q = query.trim();
    if (q.length < 1) {
      setHits([]);
      setSearching(false);
      return;
    }
    let cancelled = false;
    setSearching(true);
    const handle = setTimeout(() => {
      void searchRecipients(q)
        .then((next) => {
          if (!cancelled) {
            setHits(next);
            setError(null);
          }
        })
        .catch((err) => {
          if (!cancelled) {
            setError(err instanceof Error ? err.message : 'Could not search');
            setHits([]);
          }
        })
        .finally(() => {
          if (!cancelled) {
            setSearching(false);
          }
        });
    }, 200);
    return () => {
      cancelled = true;
      clearTimeout(handle);
    };
  }, [query]);

  const q = query.trim().toLowerCase();
  const filteredConversations = useMemo(() => {
    if (!q) {
      return items;
    }
    return items.filter((item) => {
      const label = peerLabel(item.peer).toLowerCase();
      const unit = item.peer.unit_number.toLowerCase();
      return label.includes(q) || unit.includes(q);
    });
  }, [items, q]);

  const existingPeerIds = useMemo(() => new Set(items.map((i) => i.peer.id)), [items]);
  const existingUnits = useMemo(() => new Set(items.map((i) => i.peer.unit_number)), [items]);

  const startHits = useMemo(() => {
    if (!q) {
      return [];
    }
    return hits.filter((hit) => {
      if (hit.kind === 'user' && hit.id) {
        return !existingPeerIds.has(hit.id);
      }
      return !existingUnits.has(hit.unit_number);
    });
  }, [hits, q, existingPeerIds, existingUnits]);

  const suggestions = useMemo(() => {
    if (q) {
      return [];
    }
    return directory.filter((person) => !existingPeerIds.has(person.id)).slice(0, 8);
  }, [directory, existingPeerIds, q]);

  return (
    <View style={styles.wrap}>
      <Screen
        scroll
        inShell
        refreshing={refreshing}
        onRefresh={() => {
          setRefreshing(true);
          void load().finally(() => setRefreshing(false));
        }}
      >
        <Stack gap="lg">
          <PageHeader
            title="Messages"
            subtitle="Private one-to-one with a neighbor. Search by name or unit."
            eyebrow="Direct"
            hideTitleOnCompact
            compact={compact}
          />
          <TextField
            label="Find someone"
            value={query}
            onChangeText={setQuery}
            autoCapitalize="none"
            autoCorrect={false}
            placeholder="Name or unit number"
            returnKeyType="search"
          />
          <ErrorBanner message={error} />
          {searching ? <AppText tone="muted">Searching…</AppText> : null}

          {q && !searching && filteredConversations.length === 0 && startHits.length === 0 ? (
            <EmptyState
              title={`No matches for “${query.trim()}”.`}
              subtitle="Try another name, or a unit number for someone who stays hidden."
              actionLabel="Clear search"
              onAction={() => setQuery('')}
              icon="chatbubbles-outline"
            />
          ) : null}

          {filteredConversations.length > 0 ? (
            <Stack gap="sm">
              {q ? (
                <AppText variant="label" tone="muted" style={styles.section}>
                  Your chats
                </AppText>
              ) : null}
              <ListGroup>
                {filteredConversations.map((item, index) => {
                  const label = peerLabel(item.peer);
                  const preview = item.last_message?.content?.trim() || 'No messages yet';
                  const unitHint =
                    item.peer.directory_opt_in || label.startsWith('Unit ')
                      ? null
                      : `Unit ${item.peer.unit_number}`;
                  return (
                    <ListRow
                      key={item.id}
                      last={index === filteredConversations.length - 1}
                      onPress={() => router.push(`/messages/${item.id}`)}
                      leading={<Avatar path={item.peer.profile_picture_url} label={label} />}
                    >
                      <View style={styles.meta}>
                        <AppText variant="subtitle" numberOfLines={1} style={{ flex: 1 }}>
                          {label}
                        </AppText>
                        <AppText variant="caption" tone="muted">
                          {relativeTime(item.last_message?.created_at || item.updated_at)}
                        </AppText>
                      </View>
                      <AppText tone="muted" numberOfLines={1} style={styles.preview}>
                        {preview}
                      </AppText>
                      {unitHint ? (
                        <AppText variant="caption" tone="muted">
                          {unitHint}
                        </AppText>
                      ) : null}
                    </ListRow>
                  );
                })}
              </ListGroup>
            </Stack>
          ) : null}

          {startHits.length > 0 ? (
            <Stack gap="sm">
              <AppText variant="label" tone="muted" style={styles.section}>
                Start a new chat
              </AppText>
              <ListGroup>
                {startHits.map((hit, index) => {
                  const label =
                    hit.kind === 'user'
                      ? peerLabel({ name: hit.name, unit_number: hit.unit_number })
                      : `Unit ${hit.unit_number}`;
                  return (
                    <ListRow
                      key={`${hit.kind}-${hit.id || hit.unit_number}`}
                      last={index === startHits.length - 1}
                      onPress={() =>
                        openCompose(router, {
                          userId: hit.id ?? undefined,
                          unit: hit.unit_number,
                          name: hit.name,
                        })
                      }
                      leading={<Avatar path={hit.profile_picture_url} label={label} />}
                    >
                      <AppText variant="subtitle">{label}</AppText>
                      <AppText variant="caption" tone="muted">
                        {hit.kind === 'unit'
                          ? 'Hidden neighbor — message by unit'
                          : `Unit ${hit.unit_number}`}
                      </AppText>
                    </ListRow>
                  );
                })}
              </ListGroup>
            </Stack>
          ) : null}

          {!q && suggestions.length > 0 ? (
            <Stack gap="sm">
              <AppText variant="label" tone="muted" style={styles.section}>
                Neighbors you can message
              </AppText>
              <ListGroup>
                {suggestions.map((person, index) => (
                  <ListRow
                    key={person.id}
                    last={index === suggestions.length - 1}
                    onPress={() =>
                      openCompose(router, {
                        userId: person.id,
                        unit: person.unit_number,
                        name: person.name,
                      })
                    }
                    leading={<Avatar path={person.profile_picture_url} label={person.name} />}
                  >
                    <AppText variant="subtitle">{person.name}</AppText>
                    <AppText variant="caption" tone="muted">
                      Unit {person.unit_number}
                    </AppText>
                  </ListRow>
                ))}
              </ListGroup>
            </Stack>
          ) : null}

          {loaded && !q && items.length === 0 && suggestions.length === 0 ? (
            <EmptyState
              title={error ? 'Could not load messages.' : 'No conversations yet.'}
              subtitle={error ? undefined : 'When you’re ready, find a neighbor above or tap +.'}
              actionLabel={error ? 'Try again' : 'New message'}
              onAction={error ? () => void load() : () => router.push('/messages/new')}
              icon="chatbubbles-outline"
            />
          ) : null}
        </Stack>
      </Screen>
      <Fab onPress={() => router.push('/messages/new')} />
    </View>
  );
}
