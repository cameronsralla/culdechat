import { Pressable } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from './AppText';

const stylesFor = (t: Theme) => ({
  btn: {
    position: 'absolute' as const,
    right: t.space.xl,
    bottom: t.space.xl,
    width: 56,
    height: 56,
    borderRadius: t.radius.pill,
    backgroundColor: t.colors.brand,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
  },
  plus: { color: t.colors.white, fontSize: 28, lineHeight: 32, marginTop: -2 },
});

export function Fab({ onPress, label = '+' }: { onPress: () => void; label?: string }) {
  const styles = useStyles(stylesFor);
  return (
    <Pressable accessibilityRole="button" accessibilityLabel="Create" onPress={onPress} style={styles.btn}>
      <AppText style={styles.plus}>{label}</AppText>
    </Pressable>
  );
}
