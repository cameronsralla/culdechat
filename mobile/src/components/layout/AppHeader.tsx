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
    borderBottomColor: t.colors.line,
    backgroundColor: t.colors.surface,
  },
  logo: { width: t.layout.headerLogo, height: t.layout.headerLogo },
  titles: { flex: 1 },
});

export function AppHeader({ title, subtitle }: { title: string; subtitle?: string }) {
  const styles = useStyles(stylesFor);
  return (
    <View style={styles.bar}>
      <Image source={logo} style={styles.logo} resizeMode="contain" />
      <View style={styles.titles}>
        <AppText variant="title">{title}</AppText>
        {subtitle ? (
          <AppText variant="caption" tone="muted">
            {subtitle}
          </AppText>
        ) : null}
      </View>
    </View>
  );
}
