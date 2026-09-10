import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { StyleSheet, type ImageStyle, type TextStyle, type ViewStyle } from 'react-native';
import { theme as defaultTheme, type Theme } from './theme';

const ThemeContext = createContext<Theme>(defaultTheme);

export function ThemeProvider({
  theme = defaultTheme,
  children,
}: {
  theme?: Theme;
  children: ReactNode;
}) {
  return <ThemeContext.Provider value={theme}>{children}</ThemeContext.Provider>;
}

export function useTheme(): Theme {
  return useContext(ThemeContext);
}

type NamedStyles<T> = { [P in keyof T]: ViewStyle | TextStyle | ImageStyle };

/** Build styles from the active theme. Pass a module-level factory, not an inline function. */
export function useStyles<T extends NamedStyles<T>>(factory: (theme: Theme) => T): T {
  const current = useTheme();
  return useMemo(() => StyleSheet.create(factory(current)), [current, factory]);
}
