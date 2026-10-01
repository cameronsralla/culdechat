import { useCallback, useState } from 'react';
import { View } from 'react-native';
import { useFocusEffect, useRouter } from 'expo-router';
import { listFeed } from '../api/community';
import { authorLabel, postBadge, type FeedPost } from '../api/types';
import {
  AppText,
  EmptyState,
  ErrorBanner,
  Fab,
  HeroBand,
  PageHeader,
  PostCard,
  Stack,
} from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useCompactLayout } from '../components/layout/useCompactLayout';
import { relativeTime } from '../lib/time';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  wrap: { flex: 1 },
  more: { paddingVertical: t.space.md, alignItems: 'center' as const },
  sectionLabel: { marginTop: t.space.xs },
});

export function HomeScreen() {
  const router = useRouter();
  const compact = useCompactLayout();
  const styles = useStyles(stylesFor);
  const [posts, setPosts] = useState<FeedPost[]>([]);
  const [cursor, setCursor] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async (reset: boolean, nextCursor?: string | null, silent?: boolean) => {
    try {
      if (reset && !silent) {
        setLoading(true);
      } else if (!reset) {
        setLoadingMore(true);
      }
      const page = await listFeed({ cursor: reset ? undefined : nextCursor });
      setPosts((prev) => (reset ? page.posts : [...prev, ...page.posts]));
      setCursor(page.next_page_cursor ?? null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load feed');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      void load(true);
    }, [load]),
  );

  return (
    <View style={styles.wrap}>
      <Screen
        scroll
        inShell
        onEndReached={cursor ? () => void load(false, cursor) : undefined}
        refreshing={refreshing}
        onRefresh={() => {
          setRefreshing(true);
          void load(true, undefined, true).finally(() => setRefreshing(false));
        }}
      >
        <Stack gap="lg">
          <HeroBand>
            <PageHeader
              title="The square"
              subtitle="What’s going on where you live — open to neighbors here, not the open web."
              eyebrow="Neighbors only"
              hideTitleOnCompact={false}
              compact={compact}
            />
          </HeroBand>
          <ErrorBanner message={error} />
          {posts.length > 0 ? (
            <AppText variant="label" tone="muted" style={styles.sectionLabel}>
              Recent
            </AppText>
          ) : null}
          {loading && posts.length === 0 ? <AppText tone="muted">Loading the square…</AppText> : null}
          {!loading && posts.length === 0 ? (
            <EmptyState
              title={error ? 'Could not load the feed.' : 'Quiet for now.'}
              subtitle={
                error
                  ? undefined
                  : 'Nothing urgent. When you’re ready, share a note, ask, or hello with your neighbors.'
              }
              actionLabel={error ? 'Try again' : 'Write a post'}
              onAction={error ? () => void load(true) : () => router.push('/posts/new')}
              icon="newspaper-outline"
            />
          ) : null}
          {posts.map((post) => (
            <PostCard
              key={post.id}
              title={post.title}
              author={authorLabel(post.author)}
              board={post.board.name}
              snippet={post.snippet}
              comments={post.comment_count}
              reactions={post.reaction_count}
              badge={postBadge(post)}
              when={relativeTime(post.created_at)}
              onPress={() => router.push(`/posts/${post.id}`)}
            />
          ))}
          {loadingMore ? (
            <AppText tone="muted" style={styles.more}>
              Loading more…
            </AppText>
          ) : null}
        </Stack>
      </Screen>
      <Fab onPress={() => router.push('/posts/new')} />
    </View>
  );
}
