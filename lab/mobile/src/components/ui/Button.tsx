import { ActivityIndicator, Pressable, View, type StyleProp, type ViewStyle } from 'react-native';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppText } from './AppText';

type Variant = 'primary' | 'ghost' | 'soft' | 'danger' | 'link';
type Size = 'md' | 'sm';

type Props = {
  label: string;
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  disabled?: boolean;
  onPress?: () => void;
  style?: StyleProp<ViewStyle>;
};

const stylesFor = (t: Theme) => ({
  base: {
    minHeight: t.layout.controlMinHeight,
    borderRadius: t.radius.md,
    paddingHorizontal: t.space.xl,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
  },
  sm: {
    minHeight: t.layout.controlMinHeightSm,
    borderRadius: t.radius.pill,
    paddingHorizontal: t.space.lg,
  },
  row: { flexDirection: 'row' as const, alignItems: 'center' as const, gap: t.space.sm },
  primary: { backgroundColor: t.colors.brand, ...t.shadow.soft },
  ghost: {
    backgroundColor: t.colors.white,
    borderWidth: t.layout.borderStrong,
    borderColor: t.colors.brandLine,
  },
  soft: { backgroundColor: t.colors.brandSoft },
  danger: { backgroundColor: t.colors.danger },
  link: {
    backgroundColor: 'transparent',
    minHeight: undefined,
    paddingHorizontal: t.space.xs,
    paddingVertical: t.space.xs,
  },
  pressed: { opacity: t.opacity.pressed },
  disabled: { opacity: t.opacity.disabled },
});

export function Button({
  label,
  variant = 'primary',
  size = 'md',
  loading,
  disabled,
  style,
  onPress,
}: Props) {
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  const isDisabled = disabled || loading;

  function variantStyle(v: Variant) {
    switch (v) {
      case 'primary':
        return styles.primary;
      case 'ghost':
        return styles.ghost;
      case 'soft':
        return styles.soft;
      case 'danger':
        return styles.danger;
      case 'link':
        return styles.link;
      default: {
        const _never: never = v;
        return _never;
      }
    }
  }

  const textTone = variant === 'ghost' || variant === 'soft' || variant === 'link' ? 'brand' : 'white';

  return (
    <Pressable
      accessibilityRole="button"
      disabled={isDisabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.base,
        size === 'sm' ? styles.sm : null,
        variantStyle(variant),
        pressed && !isDisabled ? styles.pressed : null,
        isDisabled ? styles.disabled : null,
        style,
      ]}
    >
      <View style={styles.row}>
        {loading ? (
          <ActivityIndicator
            color={textTone === 'brand' ? theme.colors.brandDark : theme.colors.white}
          />
        ) : null}
        <AppText variant={size === 'sm' || variant === 'link' ? 'label' : 'subtitle'} tone={textTone}>
          {label}
        </AppText>
      </View>
    </Pressable>
  );
}
