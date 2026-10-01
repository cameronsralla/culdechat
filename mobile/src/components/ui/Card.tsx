import { Pressable, View, type ViewProps } from 'react-native';
import { useStyles, type Theme } from '../../theme';

type Props = ViewProps & {
  onPress?: () => void;
  padded?: boolean;
  accent?: 'none' | 'pin' | 'brand';
  elevated?: boolean;
};

const stylesFor = (t: Theme) => ({
  card: {
    backgroundColor: t.colors.surface,
    borderRadius: t.radius.md,
    borderWidth: t.layout.border,
    borderColor: t.colors.lineSoft,
    ...t.shadow.card,
  },
  flat: {
    ...t.shadow.soft,
  },
  padded: { padding: t.space.xl },
  pin: {
    backgroundColor: t.colors.pin,
    borderColor: t.colors.pinLine,
  },
  brand: {
    backgroundColor: t.colors.brandWash,
    borderColor: t.colors.brandLine,
  },
  pressed: { opacity: 0.94, transform: [{ scale: 0.995 }] },
});

export function Card({
  onPress,
  padded = true,
  accent = 'none',
  elevated = true,
  style,
  children,
  ...rest
}: Props) {
  const styles = useStyles(stylesFor);
  const body = (
    <View
      style={[
        styles.card,
        !elevated ? styles.flat : null,
        padded ? styles.padded : null,
        accent === 'pin' ? styles.pin : null,
        accent === 'brand' ? styles.brand : null,
        style,
      ]}
      {...rest}
    >
      {children}
    </View>
  );
  if (!onPress) {
    return body;
  }
  return (
    <Pressable onPress={onPress} style={({ pressed }) => [pressed ? styles.pressed : null]}>
      {body}
    </Pressable>
  );
}
