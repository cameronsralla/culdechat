import { View, type StyleProp, type ViewStyle } from 'react-native';
import { type ReactNode } from 'react';
import { useTheme, type Space } from '../../theme';

type Props = {
  gap?: Space;
  children: ReactNode;
  style?: StyleProp<ViewStyle>;
};

/** Vertical stack using theme spacing. Default gap is `md`. */
export function Stack({ gap = 'md', style, children }: Props) {
  const theme = useTheme();
  return <View style={[{ gap: theme.space[gap] }, style]}>{children}</View>;
}
