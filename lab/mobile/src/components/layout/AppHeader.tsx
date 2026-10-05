import { Image, View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from '../ui/AppText';

const logo = require('../../../assets/logo.jpg');

const stylesFor = (t: Theme) => ({
  bar: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.md,
    paddingHorizontal: t.space.xl,
    paddingVertical: t.space.md,
    borderBottomWidth: t.layout.border,
    borderBottomColor: t.colors.lineSoft,
    backgroundColor: t.colors.surface,
  },
  logoWrap: {
    width: t.layout.headerLogo + 8,
    height: t.layout.headerLogo + 8,
    borderRadius: t.radius.sm,
    backgroundColor: t.colors.brandWash,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    overflow: 'hidden' as const,
  },
  logo: { width: t.layout.headerLogo, height: t.layout.headerLogo },
  titles: { flex: 1, gap: 1 },
});

export function AppHeader({ title }: { title: string }) {
  const styles = useStyles(stylesFor);
  return (
    <View style={styles.bar}>
      <View style={styles.logoWrap}>
        <Image source={logo} style={styles.logo} resizeMode="contain" />
      </View>
      <View style={styles.titles}>
        <AppText variant="subtitle">{title}</AppText>
        <AppText variant="caption" tone="muted">
          Neighbors only
        </AppText>
      </View>
    </View>
  );
}
