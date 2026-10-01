import { View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from './AppText';

type Props = {
  title: string;
  subtitle?: string;
  /** Soft place / exclusivity line (e.g. "Neighbors only"). */
  eyebrow?: string;
  /** Hide on compact layouts when the shell header already names the tab. */
  hideTitleOnCompact?: boolean;
  compact?: boolean;
};

const stylesFor = (t: Theme) => ({
  wrap: { gap: t.space.sm, marginBottom: t.space.xs },
  eyebrowRow: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: t.space.sm,
  },
  chip: {
    alignSelf: 'flex-start' as const,
    backgroundColor: t.colors.brandSoft,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.md,
    paddingVertical: t.space.xs,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
  },
  subtitle: { maxWidth: 520 },
});

export function PageHeader({ title, subtitle, eyebrow, hideTitleOnCompact, compact }: Props) {
  const styles = useStyles(stylesFor);
  const showTitle = !(hideTitleOnCompact && compact);
  return (
    <View style={styles.wrap}>
      {eyebrow ? (
        <View style={styles.eyebrowRow}>
          <View style={styles.chip}>
            <AppText variant="label" tone="brand">
              {eyebrow}
            </AppText>
          </View>
        </View>
      ) : null}
      {showTitle ? <AppText variant="title">{title}</AppText> : null}
      {subtitle ? (
        <AppText variant="body" tone="muted" style={styles.subtitle}>
          {subtitle}
        </AppText>
      ) : null}
    </View>
  );
}
