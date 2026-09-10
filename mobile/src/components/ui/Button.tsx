import { ActivityIndicator, Pressable, View, type StyleProp, type ViewStyle } from 'react-native';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppText } from './AppText';

type Variant = 'primary' | 'ghost' | 'danger';

type Props = {
  label: string;
  variant?: Variant;
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
  row: { flexDirection: 'row' as const, alignItems: 'center' as const, gap: t.space.sm },
  primary: { backgroundColor: t.colors.brand },
  ghost: { backgroundColor: 'transparent', borderWidth: t.layout.borderStrong, borderColor: t.colors.brand },
  danger: { backgroundColor: t.colors.danger },
  pressed: { opacity: t.opacity.pressed },
  disabled: { opacity: t.opacity.disabled },
});

export function Button({ label, variant = 'primary', loading, disabled, style, onPress }: Props) {
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  const isDisabled = disabled || loading;

  function variantStyle(v: Variant) {
    switch (v) {
      case 'primary':
        return styles.primary;
      case 'ghost':
        return styles.ghost;
      case 'danger':
        return styles.danger;
      default: {
        const _never: never = v;
        return _never;
      }
    }
  }

  return (
    <Pressable
      accessibilityRole="button"
      disabled={isDisabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.base,
        variantStyle(variant),
        pressed && !isDisabled ? styles.pressed : null,
        isDisabled ? styles.disabled : null,
        style,
      ]}
    >
      <View style={styles.row}>
        {loading ? (
          <ActivityIndicator color={variant === 'ghost' ? theme.colors.brandDark : theme.colors.white} />
        ) : null}
        <AppText variant="subtitle" tone={variant === 'ghost' ? 'brand' : 'white'}>
          {label}
        </AppText>
      </View>
    </Pressable>
  );
}
