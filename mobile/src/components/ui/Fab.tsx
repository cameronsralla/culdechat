import { Pressable } from 'react-native';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppIcon } from './Icon';

const stylesFor = (t: Theme) => ({
  btn: {
    position: 'absolute' as const,
    right: t.space.xl,
    bottom: t.space.xl,
    width: 58,
    height: 58,
    borderRadius: t.radius.pill,
    backgroundColor: t.colors.brand,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    ...t.shadow.raised,
  },
});

export function Fab({ onPress }: { onPress: () => void; label?: string }) {
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel="Create"
      onPress={onPress}
      style={({ pressed }) => [styles.btn, pressed ? { opacity: theme.opacity.pressed } : null]}
    >
      <AppIcon name="add" size={28} color={theme.colors.white} />
    </Pressable>
  );
}
