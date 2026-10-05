import { Text as RNText, type TextProps, type TextStyle } from 'react-native';
import { useTheme, type TypeVariant } from '../../theme';

type Tone = 'ink' | 'muted' | 'brand' | 'danger' | 'white';

type Props = TextProps & {
  variant?: TypeVariant;
  tone?: Tone;
};

export function AppText({ variant = 'body', tone = 'ink', style, ...rest }: Props) {
  const theme = useTheme();
  const toneColor: Record<Tone, string> = {
    ink: theme.colors.ink,
    muted: theme.colors.muted,
    brand: theme.colors.brandDark,
    danger: theme.colors.danger,
    white: theme.colors.white,
  };
  const base: TextStyle = { ...theme.type[variant], color: toneColor[tone] };
  return <RNText style={[base, style]} {...rest} />;
}
