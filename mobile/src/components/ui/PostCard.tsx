import { View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from './AppText';
import { Card } from './Card';

export type PostCardProps = {
  title: string;
  author: string;
  board?: string;
  snippet: string;
  comments: number;
  reactions: number;
  badge?: 'bulletin' | 'pinned' | null;
  when?: string;
  onPress?: () => void;
};

const stylesFor = (t: Theme) => ({
  meta: { flexDirection: 'row' as const, justifyContent: 'space-between' as const, marginBottom: t.space.xs },
  title: { marginBottom: t.space.xs },
  counts: { marginTop: t.space.sm },
});

export function PostCard({
  title,
  author,
  board,
  snippet,
  comments,
  reactions,
  badge,
  when,
  onPress,
}: PostCardProps) {
  const styles = useStyles(stylesFor);
  const badgeLabel = badge === 'bulletin' ? 'Bulletin' : badge === 'pinned' ? 'Pinned' : null;
  return (
    <Card onPress={onPress} accent={badge ? 'pin' : 'none'}>
      <View style={styles.meta}>
        <AppText variant="caption" tone="muted">
          {author}
          {board ? ` · in ${board}` : ''}
          {when ? ` · ${when}` : ''}
        </AppText>
        {badgeLabel ? (
          <AppText variant="caption" tone="brand">
            {badgeLabel}
          </AppText>
        ) : null}
      </View>
      <AppText variant="subtitle" style={styles.title}>
        {title}
      </AppText>
      <AppText variant="body" tone="muted" numberOfLines={2}>
        {snippet}
      </AppText>
      <AppText variant="caption" tone="muted" style={styles.counts}>
        {comments} comments · {reactions} reactions
      </AppText>
    </Card>
  );
}
