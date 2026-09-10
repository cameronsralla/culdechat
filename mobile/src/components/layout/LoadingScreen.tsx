import { ActivityIndicator, View } from 'react-native';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppText } from '../ui/AppText';
import { Logo } from '../ui/Logo';

const stylesFor = (t: Theme) => ({
  wrap: {
    flex: 1,
    backgroundColor: t.colors.paper,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    gap: t.space.md,
  },
});

export function LoadingScreen({ label = 'Loading…' }: { label?: string }) {
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  return (
    <View style={styles.wrap}>
      <Logo size={theme.layout.logoLoading} />
      <ActivityIndicator color={theme.colors.brand} />
      <AppText variant="caption" tone="muted">
        {label}
      </AppText>
    </View>
  );
}
