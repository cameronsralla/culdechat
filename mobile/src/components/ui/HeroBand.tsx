import { type ReactNode } from 'react';
import { View } from 'react-native';
import { useStyles, type Theme } from '../../theme';

/** Soft branded intro strip for primary screens (town-square atmosphere). */
export function HeroBand({ children }: { children: ReactNode }) {
  const styles = useStyles(stylesFor);
  return (
    <View style={styles.band}>
      <View style={styles.orb} />
      <View style={styles.content}>{children}</View>
    </View>
  );
}

const stylesFor = (t: Theme) => ({
  band: {
    position: 'relative' as const,
    overflow: 'hidden' as const,
    backgroundColor: t.colors.brandWash,
    borderRadius: t.radius.lg,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
    paddingHorizontal: t.space.xl,
    paddingVertical: t.space.xl,
    minHeight: t.layout.heroMinHeight,
    ...t.shadow.soft,
  },
  orb: {
    position: 'absolute' as const,
    right: -36,
    top: -48,
    width: 160,
    height: 160,
    borderRadius: 80,
    backgroundColor: t.colors.brandSoft,
    opacity: 0.85,
  },
  content: { position: 'relative' as const, zIndex: 1, gap: t.space.sm },
});
