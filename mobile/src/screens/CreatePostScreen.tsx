import { useCallback, useState } from 'react';
import { Pressable, View } from 'react-native';
import { useFocusEffect, useLocalSearchParams, useRouter } from 'expo-router';
import { createPost, listBoards } from '../api/community';
import type { Board } from '../api/types';
import { useAuth } from '../auth/AuthContext';
import { AppText, Button, ErrorBanner, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  chips: { flexDirection: 'row' as const, flexWrap: 'wrap' as const, gap: t.space.sm },
  chip: {
    borderWidth: t.layout.borderStrong,
    borderColor: t.colors.line,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.md,
    paddingVertical: t.space.sm,
    backgroundColor: t.colors.white,
  },
  chipOn: { borderColor: t.colors.brand, backgroundColor: t.colors.brandSoft },
});

export function CreatePostScreen() {
  const router = useRouter();
  const { boardId: preset } = useLocalSearchParams<{ boardId?: string }>();
  const { user } = useAuth();
  const styles = useStyles(stylesFor);
  const [boards, setBoards] = useState<Board[]>([]);
  const [boardId, setBoardId] = useState(preset ?? '');
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [bulletin, setBulletin] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useFocusEffect(
    useCallback(() => {
      void listBoards()
        .then((all) => {
          const choices = user?.is_admin ? all : all.filter((b) => b.is_subscribed);
          setBoards(choices);
          if (!boardId && choices.length === 1) {
            setBoardId(choices[0].id);
          }
        })
        .catch((err: unknown) => {
          setError(err instanceof Error ? err.message : 'Could not load boards');
        });
    }, [boardId, user?.is_admin]),
  );

  async function onSubmit() {
    setError(null);
    setLoading(true);
    try {
      const post = await createPost(boardId, {
        title: title.trim(),
        content: content.trim(),
        post_type: bulletin ? 'bulletin' : 'standard',
      });
      router.replace(`/posts/${post.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create post');
    } finally {
      setLoading(false);
    }
  }

  return (
    <Screen scroll inShell>
      <Stack gap="lg">
        <AppText variant="title">New post</AppText>
        <ErrorBanner message={error} />
        <AppText variant="label" tone="muted">
          Board
        </AppText>
        {boards.length === 0 ? (
          <AppText tone="muted">Subscribe to a board before posting.</AppText>
        ) : (
          <View style={styles.chips}>
            {boards.map((board) => {
              const on = board.id === boardId;
              return (
                <Pressable
                  key={board.id}
                  onPress={() => setBoardId(board.id)}
                  style={[styles.chip, on ? styles.chipOn : null]}
                >
                  <AppText variant="label" tone={on ? 'brand' : 'ink'}>
                    {board.name}
                  </AppText>
                </Pressable>
              );
            })}
          </View>
        )}
        <TextField label="Title" value={title} onChangeText={setTitle} />
        <TextField label="Content" value={content} onChangeText={setContent} multiline />
        {user?.is_admin ? (
          <Button
            label={bulletin ? 'Bulletin (comments off)' : 'Standard post'}
            variant="ghost"
            onPress={() => setBulletin((v) => !v)}
          />
        ) : null}
        <Button
          label="Post"
          onPress={() => void onSubmit()}
          loading={loading}
          disabled={!boardId || !title.trim() || !content.trim()}
        />
        <Button label="Cancel" variant="ghost" onPress={() => router.back()} />
      </Stack>
    </Screen>
  );
}
