import { View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from './AppText';
import { AppIcon } from './Icon';
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
  top: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    justifyContent: 'space-between' as const,
    gap: t.space.sm,
    marginBottom: t.space.md,
  },
  meta: { flex: 1, minWidth: 0 },
  badge: {
    backgroundColor: t.colors.white,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.sm,
    paddingVertical: t.space.xxs,
    borderWidth: t.layout.border,
    borderColor: t.colors.pinLine,
  },
  title: { marginBottom: t.space.xs },
  counts: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.lg,
    marginTop: t.space.md,
    paddingTop: t.space.md,
    borderTopWidth: t.layout.border,
    borderTopColor: t.colors.lineSoft,
  },
  count: { flexDirection: 'row' as const, alignItems: 'center' as const, gap: t.space.xs },
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
      <View style={styles.top}>
        <AppText variant="caption" tone="muted" style={styles.meta} numberOfLines={1}>
          {author}
          {board ? ` · ${board}` : ''}
          {when ? ` · ${when}` : ''}
        </AppText>
        {badgeLabel ? (
          <View style={styles.badge}>
            <AppText variant="label" tone="brand">
              {badgeLabel}
            </AppText>
          </View>
        ) : null}
      </View>
      <AppText variant="subtitle" style={styles.title}>
        {title}
      </AppText>
      <AppText variant="body" tone="muted" numberOfLines={3}>
        {snippet}
      </AppText>
      <View style={styles.counts}>
        <View style={styles.count}>
          <AppIcon name="chatbubble-outline" size={15} />
          <AppText variant="caption" tone="muted">
            {comments}
          </AppText>
        </View>
        <View style={styles.count}>
          <AppIcon name="heart-outline" size={15} />
          <AppText variant="caption" tone="muted">
            {reactions}
          </AppText>
        </View>
      </View>
    </Card>
  );
}
