import type { TextStyle } from 'react-native';

/**
 * Visual source of truth. Screens should use the UI kit and `useStyles`,
 * not hard-coded colors or font sizes.
 */
export const theme = {
  colors: {
    brand: '#1A9BA8',
    brandDark: '#147A85',
    brandSoft: '#E7F6F7',
    brandLine: '#C5E6EA',
    paper: '#F6F1E8',
    surface: '#FFFCF7',
    ink: '#1C2328',
    muted: '#667078',
    line: '#E4DCD0',
    danger: '#B42318',
    dangerSoft: '#F9E8E6',
    pin: '#F4E6C4',
    pinLine: '#E6D09A',
    white: '#FFFFFF',
  },
  space: {
    xxs: 2,
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    xxl: 32,
    xxxl: 40,
  },
  radius: {
    sm: 10,
    md: 16,
    lg: 24,
    pill: 999,
  },
  type: {
    display: { fontFamily: 'Nunito_800ExtraBold', fontSize: 28, lineHeight: 34 },
    title: { fontFamily: 'Nunito_700Bold', fontSize: 22, lineHeight: 28 },
    subtitle: { fontFamily: 'Nunito_600SemiBold', fontSize: 17, lineHeight: 22 },
    body: { fontFamily: 'Nunito_400Regular', fontSize: 16, lineHeight: 22 },
    label: { fontFamily: 'Nunito_600SemiBold', fontSize: 14, lineHeight: 18 },
    caption: { fontFamily: 'Nunito_400Regular', fontSize: 13, lineHeight: 18 },
  } satisfies Record<string, TextStyle>,
  layout: {
    compactBreakpoint: 800,
    sidebarWidth: 240,
    contentMax: 840,
    loginMax: 420,
    controlMinHeight: 52,
    inputPadY: 14,
    headerLogo: 40,
    sidebarLogo: 44,
    logoLoading: 72,
    logoAuth: 88,
    logoAuthCompact: 108,
    navIcon: 20,
    avatar: 48,
    avatarLg: 88,
    border: 1,
    borderStrong: 1.5,
  },
  opacity: {
    pressed: 0.88,
    disabled: 0.5,
  },
} as const;

export type Theme = typeof theme;
export type Space = keyof Theme['space'];
export type TypeVariant = keyof Theme['type'];
export type ColorName = keyof Theme['colors'];
