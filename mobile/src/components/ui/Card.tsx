import { Pressable, View, type ViewProps } from 'react-native';
import { useStyles, type Theme } from '../../theme';

type Props = ViewProps & {
  onPress?: () => void;
  padded?: boolean;
  accent?: 'none' | 'pin' | 'brand';
};

const stylesFor = (t: Theme) => ({
  card: {
    backgroundColor: t.colors.surface,
    borderRadius: t.radius.md,
    borderWidth: t.layout.border,
    borderColor: t.colors.line,
  },
  padded: { padding: t.space.lg },
  pin: { backgroundColor: t.colors.pin, borderColor: t.colors.pinLine },
  brand: { backgroundColor: t.colors.brandSoft, borderColor: t.colors.brandLine },
  pressed: { opacity: 0.92 },
});

export function Card({ onPress, padded = true, accent = 'none', style, children, ...rest }: Props) {
  const styles = useStyles(stylesFor);
  const body = (
    <View
      style={[
        styles.card,
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
