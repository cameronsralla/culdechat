import { useCallback, useEffect, useMemo, useState } from 'react';
import { View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { listDirectory, searchRecipients, sendDirectMessage } from '../api/community';
import { peerLabel, type DirectoryUser, type MessageRecipient } from '../api/types';
import { AppText, Avatar, Button, Card, EmptyState, ErrorBanner, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  row: { flexDirection: 'row' as const, alignItems: 'center' as const, gap: t.space.md },
  grow: { flex: 1, minWidth: 0 },
  selected: {
    borderWidth: t.layout.borderStrong,
    borderColor: t.colors.brand,
    borderRadius: t.radius.md,
    padding: t.space.md,
    backgroundColor: t.colors.brandSoft,
  },
});

type Selected =
  | { kind: 'user'; id: string; name?: string | null; unit_number: string; profile_picture_url?: string | null }
  | { kind: 'unit'; unit_number: string };

function asParam(value: string | string[] | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? '';
  }
  return value ?? '';
}

export function NewMessageScreen() {
  const router = useRouter();
  const params = useLocalSearchParams<{ userId?: string; unit?: string; name?: string }>();
  const styles = useStyles(stylesFor);
  const [query, setQuery] = useState('');
  const [hits, setHits] = useState<MessageRecipient[]>([]);
  const [directory, setDirectory] = useState<DirectoryUser[]>([]);
  const [selected, setSelected] = useState<Selected | null>(null);
  const [draft, setDraft] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);
  const [searching, setSearching] = useState(false);

  useEffect(() => {
    const userId = asParam(params.userId).trim();
    const unit = asParam(params.unit).trim();
    const name = asParam(params.name).trim();
    if (userId) {
      setSelected({ kind: 'user', id: userId, name: name || null, unit_number: unit });
      return;
    }
    if (unit) {
      setSelected({ kind: 'unit', unit_number: unit });
    }
  }, [params.userId, params.unit, params.name]);

  useEffect(() => {
    void listDirectory()
      .then(setDirectory)
      .catch(() => setDirectory([]));
  }, []);

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

  const selectedLabel = useMemo(() => {
    if (!selected) {
      return '';
    }
    if (selected.kind === 'user') {
      return peerLabel({ name: selected.name, unit_number: selected.unit_number });
    }
    return `Unit ${selected.unit_number}`;
  }, [selected]);

  const pickUser = useCallback((hit: MessageRecipient) => {
    if (hit.kind === 'user' && hit.id) {
      setSelected({
        kind: 'user',
        id: hit.id,
        name: hit.name,
        unit_number: hit.unit_number,
        profile_picture_url: hit.profile_picture_url,
      });
    } else {
      setSelected({ kind: 'unit', unit_number: hit.unit_number });
    }
    setQuery('');
    setHits([]);
  }, []);

  const pickDirectory = useCallback((person: DirectoryUser) => {
    setSelected({
      kind: 'user',
      id: person.id,
      name: person.name,
      unit_number: person.unit_number,
      profile_picture_url: person.profile_picture_url,
    });
    setQuery('');
    setHits([]);
  }, []);

  async function onSend() {
    if (!selected || !draft.trim()) {
      return;
    }
    setSending(true);
    try {
      const result = await sendDirectMessage({
        content: draft.trim(),
        ...(selected.kind === 'user' ? { user_id: selected.id } : { unit_number: selected.unit_number }),
      });
      router.replace(`/messages/${result.conversation_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not send message');
    } finally {
      setSending(false);
    }
  }

  const q = query.trim();
  const showDirectory = !selected && !q;
  const showNoHits = !selected && q.length > 0 && !searching && hits.length === 0;

  return (
    <Screen scroll inShell>
      <Stack gap="lg">
        <AppText variant="title">New message</AppText>
        <AppText tone="muted">
          Search by name for people in the directory, or by unit number for neighbors who stay hidden.
        </AppText>
        <ErrorBanner message={error} />

        {selected ? (
          <View style={styles.selected}>
            <View style={styles.row}>
              <Avatar
                path={selected.kind === 'user' ? selected.profile_picture_url : null}
                label={selectedLabel}
              />
              <View style={styles.grow}>
                <AppText variant="subtitle">{selectedLabel}</AppText>
                {selected.kind === 'user' && selected.name ? (
                  <AppText variant="caption" tone="muted">
                    Unit {selected.unit_number}
                  </AppText>
                ) : null}
              </View>
              <Button label="Change" variant="soft" size="sm" onPress={() => setSelected(null)} />
            </View>
          </View>
        ) : (
          <Stack gap="md">
            <TextField
              label="Search people or units"
              value={query}
              onChangeText={setQuery}
              autoCapitalize="none"
              autoCorrect={false}
              placeholder="Jordan or 512"
              returnKeyType="search"
              autoFocus
            />
            {searching ? <AppText tone="muted">Searching…</AppText> : null}
            {showNoHits ? (
              <EmptyState
                title={`No matches for “${q}”.`}
                subtitle="Listed neighbors match by name or unit. Hidden neighbors only match their unit number."
                icon="people-outline"
              />
            ) : null}
            {hits.map((hit) => {
              const label =
                hit.kind === 'user'
                  ? peerLabel({ name: hit.name, unit_number: hit.unit_number })
                  : `Unit ${hit.unit_number}`;
              return (
                <Card key={`${hit.kind}-${hit.id || hit.unit_number}`} onPress={() => pickUser(hit)}>
                  <View style={styles.row}>
                    <Avatar path={hit.profile_picture_url} label={label} />
                    <Stack gap="xxs" style={styles.grow}>
                      <AppText variant="subtitle">{label}</AppText>
                      {hit.kind === 'unit' ? (
                        <AppText variant="caption" tone="muted">
                          Hidden neighbor — messages go to this unit
                        </AppText>
                      ) : (
                        <AppText variant="caption" tone="muted">
                          Unit {hit.unit_number}
                        </AppText>
                      )}
                    </Stack>
                  </View>
                </Card>
              );
            })}
            {showDirectory ? (
              <Stack gap="md">
                <AppText variant="label" tone="muted">
                  Or pick a neighbor
                </AppText>
                {directory.length === 0 ? (
                  <AppText tone="muted">Nobody is listed in the directory yet. Search by unit number.</AppText>
                ) : (
                  directory.map((person) => (
                    <Card key={person.id} onPress={() => pickDirectory(person)}>
                      <View style={styles.row}>
                        <Avatar path={person.profile_picture_url} label={person.name} />
                        <Stack gap="xxs" style={styles.grow}>
                          <AppText variant="subtitle">{person.name}</AppText>
                          <AppText variant="caption" tone="muted">
                            Unit {person.unit_number}
                          </AppText>
                        </Stack>
                      </View>
                    </Card>
                  ))
                )}
              </Stack>
            ) : null}
          </Stack>
        )}

        {selected ? (
          <Stack gap="md">
            <TextField
              label="Message"
              value={draft}
              onChangeText={setDraft}
              multiline
              placeholder="Write your first message…"
            />
            <Button label="Send" onPress={() => void onSend()} loading={sending} disabled={!draft.trim()} />
          </Stack>
        ) : null}
      </Stack>
    </Screen>
  );
}
