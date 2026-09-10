import { useCallback, useState } from 'react';
import { Pressable, View } from 'react-native';
import { useFocusEffect, useRouter } from 'expo-router';
import { listBoards, toggleSubscribe } from '../api/community';
import type { Board } from '../api/types';
import { AppText, Button, Card, EmptyState, ErrorBanner, Stack } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  row: {
    flexDirection: 'row' as const,
    justifyContent: 'space-between' as const,
    alignItems: 'center' as const,
    gap: t.space.md,
  },
});

export function BoardsScreen() {
  const router = useRouter();
  const styles = useStyles(stylesFor);
  const [boards, setBoards] = useState<Board[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [loaded, setLoaded] = useState(false);

  const load = useCallback(async () => {
    try {
      setBoards(await listBoards());
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load boards');
    } finally {
      setLoaded(true);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      void load();
    }, [load]),
  );

  async function onToggle(board: Board) {
    setBusyId(board.id);
    try {
      await toggleSubscribe(board.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not update subscription');
    } finally {
      setBusyId(null);
    }
  }

  return (
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
        <ErrorBanner message={error} />
        <Button label="Create new board" onPress={() => router.push('/boards/new')} />
        {loaded && boards.length === 0 ? (
          <EmptyState
            title={error ? 'Could not load boards.' : 'No boards yet.'}
            actionLabel={error ? 'Try again' : 'Create a board'}
            onAction={error ? () => void load() : () => router.push('/boards/new')}
          />
        ) : null}
        {boards.map((board) => (
          <Card key={board.id}>
            <Stack gap="sm">
              <Pressable onPress={() => router.push(`/boards/${board.id}`)}>
                <AppText variant="subtitle">{board.name}</AppText>
              </Pressable>
              {board.description ? (
                <AppText variant="body" tone="muted">
                  {board.description}
                </AppText>
              ) : null}
              <View style={styles.row}>
                <AppText variant="caption" tone="muted">
                  {board.subscriber_count} subscriber{board.subscriber_count === 1 ? '' : 's'}
                </AppText>
                <Button
                  label={board.is_subscribed ? 'Unsubscribe' : 'Subscribe'}
                  variant="ghost"
                  loading={busyId === board.id}
                  onPress={() => void onToggle(board)}
                />
              </View>
            </Stack>
          </Card>
        ))}
      </Stack>
    </Screen>
  );
}
