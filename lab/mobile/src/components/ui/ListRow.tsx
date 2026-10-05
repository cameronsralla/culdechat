import { type ReactNode } from 'react';
import { Pressable, View } from 'react-native';
import { useStyles, type Theme } from '../../theme';

type Props = {
  onPress?: () => void;
  leading?: ReactNode;
  trailing?: ReactNode;
  children: ReactNode;
  last?: boolean;
};

const stylesFor = (t: Theme) => ({
  row: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.md,
    paddingVertical: t.space.md,
    paddingHorizontal: t.space.lg,
    backgroundColor: t.colors.surface,
  },
  pressed: { backgroundColor: t.colors.brandWash },
  grow: { flex: 1, minWidth: 0 },
  divider: {
    borderBottomWidth: t.layout.border,
    borderBottomColor: t.colors.lineSoft,
  },
  group: {
    borderRadius: t.radius.md,
    borderWidth: t.layout.border,
    borderColor: t.colors.lineSoft,
    overflow: 'hidden' as const,
    backgroundColor: t.colors.surface,
    ...t.shadow.soft,
  },
});

export function ListGroup({ children }: { children: ReactNode }) {
  const styles = useStyles(stylesFor);
  return <View style={styles.group}>{children}</View>;
}

export function ListRow({ onPress, leading, trailing, children, last }: Props) {
  const styles = useStyles(stylesFor);
  const body = (
    <View style={[styles.row, !last ? styles.divider : null]}>
      {leading}
      <View style={styles.grow}>{children}</View>
      {trailing}
    </View>
  );
  if (!onPress) {
    return body;
  }
  return (
    <Pressable onPress={onPress} style={({ pressed }) => [pressed ? styles.pressed : null]}>
      {body}
    </Pressable>
  );
}
