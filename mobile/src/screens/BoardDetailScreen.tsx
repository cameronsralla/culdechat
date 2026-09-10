import { useCallback, useState } from 'react';
import { View } from 'react-native';
import { useFocusEffect, useLocalSearchParams, useRouter } from 'expo-router';
import { getBoard, listFeed, toggleSubscribe } from '../api/community';
import { authorLabel, postBadge, type Board, type FeedPost } from '../api/types';
import { AppText, Button, EmptyState, ErrorBanner, Fab, PostCard, Stack } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { relativeTime } from '../lib/time';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  wrap: { flex: 1 },
  row: { flexDirection: 'row' as const, gap: t.space.md, flexWrap: 'wrap' as const },
});

export function BoardDetailScreen() {
  const { boardId } = useLocalSearchParams<{ boardId: string }>();
  const router = useRouter();
  const styles = useStyles(stylesFor);
  const [board, setBoard] = useState<Board | null>(null);
  const [posts, setPosts] = useState<FeedPost[]>([]);
  const [cursor, setCursor] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async () => {
    if (!boardId) {
      return;
    }
    try {
      const [nextBoard, page] = await Promise.all([getBoard(boardId), listFeed({ boardId })]);
      setBoard(nextBoard);
      setPosts(page.posts);
      setCursor(page.next_page_cursor ?? null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load board');
    }
  }, [boardId]);

  useFocusEffect(
    useCallback(() => {
      void load();
    }, [load]),
  );

  async function loadMore() {
    if (!boardId || !cursor) {
      return;
    }
    const page = await listFeed({ boardId, cursor });
    setPosts((prev) => [...prev, ...page.posts]);
    setCursor(page.next_page_cursor ?? null);
  }

  async function onToggle() {
    if (!board) {
      return;
    }
    setBusy(true);
    try {
      await toggleSubscribe(board.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not update subscription');
    } finally {
      setBusy(false);
    }
  }

  return (
    <View style={styles.wrap}>
      <Screen
        scroll
        inShell
        onEndReached={cursor ? () => void loadMore() : undefined}
        refreshing={refreshing}
        onRefresh={() => {
          setRefreshing(true);
          void load().finally(() => setRefreshing(false));
        }}
      >
        <Stack gap="lg">
          <ErrorBanner message={error} />
          <AppText variant="title">{board?.name ?? 'Board'}</AppText>
          {board?.description ? <AppText tone="muted">{board.description}</AppText> : null}
          <View style={styles.row}>
            <Button
              label={board?.is_subscribed ? 'Unsubscribe' : 'Subscribe'}
              variant="ghost"
              loading={busy}
              onPress={() => void onToggle()}
            />
            <Button label="Back" variant="ghost" onPress={() => router.back()} />
          </View>
          {posts.length === 0 ? (
            <EmptyState
              title={error ? 'Could not load posts.' : 'No posts on this board yet.'}
              actionLabel={error ? 'Try again' : board?.is_subscribed ? 'Write a post' : undefined}
              onAction={error ? () => void load() : board?.is_subscribed ? () => router.push(`/posts/new?boardId=${board.id}`) : undefined}
            />
          ) : null}
          {posts.map((post) => (
            <PostCard
              key={post.id}
              title={post.title}
              author={authorLabel(post.author)}
              snippet={post.snippet}
              comments={post.comment_count}
              reactions={post.reaction_count}
              badge={postBadge(post)}
              when={relativeTime(post.created_at)}
              onPress={() => router.push(`/posts/${post.id}`)}
            />
          ))}
        </Stack>
      </Screen>
      {board?.is_subscribed ? <Fab onPress={() => router.push(`/posts/new?boardId=${board.id}`)} /> : null}
    </View>
  );
}
