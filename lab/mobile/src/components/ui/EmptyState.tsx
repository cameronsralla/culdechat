import { View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from './AppText';
import { Button } from './Button';
import { AppIcon } from './Icon';
import { Stack } from './Stack';

type Props = {
  title: string;
  subtitle?: string;
  actionLabel?: string;
  onAction?: () => void;
  icon?: 'chatbubbles-outline' | 'people-outline' | 'newspaper-outline' | 'grid-outline';
};

const stylesFor = (t: Theme) => ({
  wrap: {
    alignItems: 'center' as const,
    paddingVertical: t.space.xxxl,
    paddingHorizontal: t.space.xl,
    backgroundColor: t.colors.surface,
    borderRadius: t.radius.lg,
    borderWidth: t.layout.border,
    borderColor: t.colors.lineSoft,
  },
  iconWrap: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: t.colors.brandWash,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    marginBottom: t.space.sm,
  },
  title: { textAlign: 'center' as const },
  subtitle: { textAlign: 'center' as const, maxWidth: 340 },
});

export function EmptyState({ title, subtitle, actionLabel, onAction, icon = 'newspaper-outline' }: Props) {
  const styles = useStyles(stylesFor);
  return (
    <View style={styles.wrap}>
      <Stack gap="md" style={{ alignItems: 'center' }}>
        <View style={styles.iconWrap}>
          <AppIcon name={icon} size={28} color="#085A62" />
        </View>
        <AppText variant="subtitle" style={styles.title}>
          {title}
        </AppText>
        {subtitle ? (
          <AppText tone="muted" style={styles.subtitle}>
            {subtitle}
          </AppText>
        ) : null}
        {actionLabel && onAction ? (
          <Button label={actionLabel} onPress={onAction} size="sm" variant="soft" />
        ) : null}
      </Stack>
    </View>
  );
}
