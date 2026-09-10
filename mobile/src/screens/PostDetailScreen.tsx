import { useCallback, useState } from 'react';
import { Pressable, View } from 'react-native';
import { useFocusEffect, useLocalSearchParams, useRouter } from 'expo-router';
import {
  addComment,
  clearReaction,
  deleteComment,
  deletePost,
  getPost,
  setReaction,
  updateComment,
  updatePost,
} from '../api/community';
import { authorLabel, postBadge, reactionEmoji, reactionTypes, type PostDetail } from '../api/types';
import { useAuth } from '../auth/AuthContext';
import { AppText, Button, Card, ErrorBanner, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';
import { confirm } from '../lib/confirm';
import { relativeTime } from '../lib/time';
import { useStyles, type Theme } from '../theme';

const stylesFor = (t: Theme) => ({
  reactions: { flexDirection: 'row' as const, flexWrap: 'wrap' as const, gap: t.space.sm },
  chip: {
    borderWidth: t.layout.border,
    borderColor: t.colors.line,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.md,
    paddingVertical: t.space.sm,
    backgroundColor: t.colors.white,
  },
  chipOn: { borderColor: t.colors.brand, backgroundColor: t.colors.brandSoft },
  composer: { flexDirection: 'row' as const, gap: t.space.sm, alignItems: 'flex-end' as const },
  grow: { flex: 1 },
  actions: { flexDirection: 'row' as const, flexWrap: 'wrap' as const, gap: t.space.sm },
});

export function PostDetailScreen() {
  const { postId } = useLocalSearchParams<{ postId: string }>();
  const router = useRouter();
  const { user } = useAuth();
  const styles = useStyles(stylesFor);
  const [post, setPost] = useState<PostDetail | null>(null);
  const [draft, setDraft] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);
  const [editing, setEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editBody, setEditBody] = useState('');
  const [editingCommentId, setEditingCommentId] = useState<string | null>(null);
  const [editComment, setEditComment] = useState('');

  const load = useCallback(async () => {
    if (!postId) {
      return;
    }
    try {
      const next = await getPost(postId);
      setPost(next);
      setEditTitle(next.title);
      setEditBody(next.content);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load post');
    }
  }, [postId]);

  useFocusEffect(
    useCallback(() => {
      void load();
    }, [load]),
  );

  const canManagePost = Boolean(post && user && (user.id === post.author.id || user.is_admin));
  const badge = post ? postBadge(post) : null;
  const badgeLabel = badge === 'bulletin' ? 'Bulletin' : badge === 'pinned' ? 'Pinned' : null;

  async function onReact(type: string) {
    if (!post) {
      return;
    }
    try {
      if (post.my_reaction === type) {
        await clearReaction(post.id);
      } else {
        await setReaction(post.id, type);
      }
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not react');
    }
  }

  async function onComment() {
    if (!post || !draft.trim()) {
      return;
    }
    setSending(true);
    try {
      await addComment(post.id, draft.trim());
      setDraft('');
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not comment');
    } finally {
      setSending(false);
    }
  }

  async function onSavePost() {
    if (!post) {
      return;
    }
    setSending(true);
    try {
      await updatePost(post.id, { title: editTitle.trim(), content: editBody.trim() });
      setEditing(false);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save post');
    } finally {
      setSending(false);
    }
  }

  async function onTogglePin() {
    if (!post) {
      return;
    }
    try {
      await updatePost(post.id, { is_pinned: !post.is_pinned });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not update pin');
    }
  }

  function onDeletePost() {
    if (!post) {
      return;
    }
    confirm('Delete this post?', () => {
      void deletePost(post.id)
        .then(() => router.replace('/'))
        .catch((err: unknown) => {
          setError(err instanceof Error ? err.message : 'Could not delete post');
        });
    });
  }

  async function onSaveComment(commentId: string) {
    if (!post || !editComment.trim()) {
      return;
    }
    try {
      await updateComment(post.id, commentId, editComment.trim());
      setEditingCommentId(null);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save comment');
    }
  }

  function onDeleteComment(commentId: string) {
    if (!post) {
      return;
    }
    confirm('Delete this comment?', () => {
      void deleteComment(post.id, commentId)
        .then(load)
        .catch((err: unknown) => {
          setError(err instanceof Error ? err.message : 'Could not delete comment');
        });
    });
  }

  const counts = new Map((post?.reactions ?? []).map((r) => [r.type, r.count]));

  return (
    <Screen scroll inShell>
      <Stack gap="lg">
        <ErrorBanner message={error} />
        {post ? (
          <>
            <AppText variant="caption" tone="muted">
              {authorLabel(post.author)} · in {post.board.name} · {relativeTime(post.created_at)}
              {badgeLabel ? ` · ${badgeLabel}` : ''}
            </AppText>
            {editing ? (
              <Stack gap="md">
                <TextField label="Title" value={editTitle} onChangeText={setEditTitle} />
                <TextField label="Post" value={editBody} onChangeText={setEditBody} multiline />
                <View style={styles.actions}>
                  <Button label="Save" onPress={() => void onSavePost()} loading={sending} disabled={!editTitle.trim() || !editBody.trim()} />
                  <Button label="Cancel" variant="ghost" onPress={() => setEditing(false)} />
                </View>
              </Stack>
            ) : (
              <>
                <AppText variant="title">{post.title}</AppText>
                <AppText>{post.content}</AppText>
              </>
            )}
            {canManagePost && !editing ? (
              <View style={styles.actions}>
                <Button label="Edit" variant="ghost" onPress={() => setEditing(true)} />
                {user?.is_admin && post.post_type !== 'bulletin' ? (
                  <Button label={post.is_pinned ? 'Unpin' : 'Pin'} variant="ghost" onPress={() => void onTogglePin()} />
                ) : null}
                <Button label="Delete" variant="danger" onPress={onDeletePost} />
              </View>
            ) : null}
            <View style={styles.reactions}>
              {reactionTypes.map((type) => {
                const on = post.my_reaction === type;
                return (
                  <Pressable
                    key={type}
                    onPress={() => void onReact(type)}
                    style={[styles.chip, on ? styles.chipOn : null]}
                  >
                    <AppText variant="label">
                      {reactionEmoji[type]} {counts.get(type) ?? 0}
                    </AppText>
                  </Pressable>
                );
              })}
            </View>
            {post.comments_disabled ? (
              <AppText tone="muted">Comments are off for this bulletin.</AppText>
            ) : (
              <>
                <AppText variant="subtitle">Comments</AppText>
                {post.comments.length === 0 ? <AppText tone="muted">No comments yet.</AppText> : null}
                {post.comments.map((comment) => {
                  const mine = user?.id === comment.author.id;
                  const canDelete = Boolean(mine || user?.is_admin);
                  return (
                    <Card key={comment.id}>
                      <Stack gap="xs">
                        <AppText variant="caption" tone="muted">
                          {authorLabel(comment.author)}
                          {comment.created_at ? ` · ${relativeTime(comment.created_at)}` : ''}
                        </AppText>
                        {editingCommentId === comment.id ? (
                          <>
                            <TextField label="Comment" value={editComment} onChangeText={setEditComment} multiline />
                            <View style={styles.actions}>
                              <Button label="Save" onPress={() => void onSaveComment(comment.id)} disabled={!editComment.trim()} />
                              <Button label="Cancel" variant="ghost" onPress={() => setEditingCommentId(null)} />
                            </View>
                          </>
                        ) : (
                          <AppText>{comment.content}</AppText>
                        )}
                        {editingCommentId !== comment.id && (mine || canDelete) ? (
                          <View style={styles.actions}>
                            {mine ? (
                              <Button
                                label="Edit"
                                variant="ghost"
                                onPress={() => {
                                  setEditingCommentId(comment.id);
                                  setEditComment(comment.content);
                                }}
                              />
                            ) : null}
                            {canDelete ? (
                              <Button label="Delete" variant="danger" onPress={() => onDeleteComment(comment.id)} />
                            ) : null}
                          </View>
                        ) : null}
                      </Stack>
                    </Card>
                  );
                })}
                <View style={styles.composer}>
                  <View style={styles.grow}>
                    <TextField label="Write a comment" value={draft} onChangeText={setDraft} />
                  </View>
                  <Button label="Send" onPress={() => void onComment()} loading={sending} disabled={!draft.trim()} />
                </View>
              </>
            )}
          </>
        ) : (
          <AppText tone="muted">Loading…</AppText>
        )}
        <Button label="Back" variant="ghost" onPress={() => router.back()} />
      </Stack>
    </Screen>
  );
}
