import { useWindowDimensions } from 'react-native';
import { useTheme } from '../../theme';

export function useCompactLayout(): boolean {
  const { layout } = useTheme();
  const { width } = useWindowDimensions();
  return width < layout.compactBreakpoint;
}
