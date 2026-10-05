import { useCallback, useEffect, useRef, useState } from 'react';
import { View } from 'react-native';
import { useFocusEffect, useLocalSearchParams } from 'expo-router';
import { getConversation, sendDirectMessage } from '../api/community';
import { peerLabel, type ConversationDetail, type DirectMessage } from '../api/types';
import { useAuth } from '../auth/AuthContext';
import { AppText, Button, ErrorBanner, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { relativeTime } from '../lib/time';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  wrap: { flex: 1 },
  bubbleMine: {
    alignSelf: 'flex-end' as const,
    maxWidth: '82%' as const,
    backgroundColor: t.colors.brand,
    borderRadius: t.radius.lg,
    borderBottomRightRadius: t.radius.sm,
    paddingHorizontal: t.space.lg,
    paddingVertical: t.space.md,
  },
  bubbleTheirs: {
    alignSelf: 'flex-start' as const,
    maxWidth: '82%' as const,
    backgroundColor: t.colors.surface,
    borderRadius: t.radius.lg,
    borderBottomLeftRadius: t.radius.sm,
    borderWidth: t.layout.border,
    borderColor: t.colors.lineSoft,
    paddingHorizontal: t.space.lg,
    paddingVertical: t.space.md,
  },
  timeMine: { marginTop: t.space.xs, color: 'rgba(255,255,255,0.75)' },
  composer: {
    flexDirection: 'row' as const,
    gap: t.space.sm,
    alignItems: 'flex-end' as const,
    backgroundColor: t.colors.surface,
    borderRadius: t.radius.lg,
    borderWidth: t.layout.border,
    borderColor: t.colors.lineSoft,
    padding: t.space.md,
    ...t.shadow.soft,
  },
  grow: { flex: 1 },
  headerChip: {
    alignSelf: 'flex-start' as const,
    backgroundColor: t.colors.brandSoft,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.md,
    paddingVertical: t.space.xs,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
  },
});

export function ConversationScreen() {
  const { conversationId } = useLocalSearchParams<{ conversationId: string }>();
  const { user } = useAuth();
  const styles = useStyles(stylesFor);
  const [detail, setDetail] = useState<ConversationDetail | null>(null);
  const [draft, setDraft] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const load = useCallback(async () => {
    if (!conversationId) {
      return;
    }
    try {
      setDetail(await getConversation(conversationId));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load conversation');
    }
  }, [conversationId]);

  useFocusEffect(
    useCallback(() => {
      void load();
      pollRef.current = setInterval(() => {
        void load();
      }, 5000);
      return () => {
        if (pollRef.current) {
          clearInterval(pollRef.current);
        }
      };
    }, [load]),
  );

  useEffect(() => {
    return () => {
      if (pollRef.current) {
        clearInterval(pollRef.current);
      }
    };
  }, []);

  async function onSend() {
    if (!detail || !draft.trim()) {
      return;
    }
    setSending(true);
    try {
      await sendDirectMessage({ content: draft.trim(), user_id: detail.peer.id });
      setDraft('');
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not send');
    } finally {
      setSending(false);
    }
  }

  const title = detail ? peerLabel(detail.peer) : 'Conversation';

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
          <Stack gap="sm">
            <AppText variant="title">{title}</AppText>
            {detail ? (
              <View style={styles.headerChip}>
                <AppText variant="label" tone="brand">
                  {detail.peer.directory_opt_in
                    ? `Unit ${detail.peer.unit_number} · Direct`
                    : `Unit ${detail.peer.unit_number} · Private`}
                </AppText>
              </View>
            ) : null}
          </Stack>
          <ErrorBanner message={error} />
          {!detail ? <AppText tone="muted">Loading…</AppText> : null}
          {detail?.messages.map((msg: DirectMessage) => {
            const mine = user?.id === msg.sender_id;
            return (
              <View key={msg.id} style={mine ? styles.bubbleMine : styles.bubbleTheirs}>
                <AppText tone={mine ? 'white' : 'ink'}>{msg.content}</AppText>
                <AppText variant="caption" tone={mine ? 'white' : 'muted'} style={mine ? styles.timeMine : undefined}>
                  {relativeTime(msg.created_at)}
                </AppText>
              </View>
            );
          })}
          {detail ? (
            <View style={styles.composer}>
              <View style={styles.grow}>
                <TextField
                  label="Message"
                  value={draft}
                  onChangeText={setDraft}
                  multiline
                  placeholder="Write a message…"
                />
              </View>
              <Button label="Send" size="sm" onPress={() => void onSend()} loading={sending} disabled={!draft.trim()} />
            </View>
          ) : null}
        </Stack>
      </Screen>
    </View>
  );
}
