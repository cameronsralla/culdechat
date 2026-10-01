import type { TextStyle, ViewStyle } from 'react-native';

/**
 * Visual source of truth. Screens should use the UI kit and `useStyles`,
 * not hard-coded colors or font sizes.
 *
 * Direction: lagoon teal town square on warm stone paper — cozy, clear,
 * neighborly. Connection over engagement; calm when idle.
 */
export const theme = {
  colors: {
    brand: '#0C7C86',
    brandDark: '#085A62',
    brandSoft: '#DDF0F2',
    brandLine: '#A9D4DA',
    brandWash: '#E8F5F6',
    paper: '#F0EBE3',
    paperDeep: '#E2D9CC',
    surface: '#FFFDF9',
    surfaceMuted: '#F7F2EA',
    ink: '#1C2529',
    muted: '#5A656C',
    line: '#D9D0C2',
    lineSoft: '#EBE3D7',
    danger: '#B42318',
    dangerSoft: '#F9E8E6',
    pin: '#F3E6C4',
    pinLine: '#E0CC8F',
    white: '#FFFFFF',
    overlay: 'rgba(28, 37, 41, 0.04)',
    inkFaint: 'rgba(28, 37, 41, 0.06)',
  },
  space: {
    xxs: 2,
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    xxl: 32,
    xxxl: 48,
  },
  radius: {
    sm: 10,
    md: 16,
    lg: 24,
    pill: 999,
  },
  type: {
    display: {
      fontFamily: 'Nunito_800ExtraBold',
      fontSize: 36,
      lineHeight: 42,
      letterSpacing: -0.8,
    },
    title: {
      fontFamily: 'Nunito_700Bold',
      fontSize: 26,
      lineHeight: 32,
      letterSpacing: -0.4,
    },
    subtitle: {
      fontFamily: 'Nunito_600SemiBold',
      fontSize: 17,
      lineHeight: 22,
      letterSpacing: -0.15,
    },
    body: {
      fontFamily: 'Nunito_400Regular',
      fontSize: 16,
      lineHeight: 24,
    },
    label: {
      fontFamily: 'Nunito_600SemiBold',
      fontSize: 12,
      lineHeight: 16,
      letterSpacing: 0.4,
    },
    caption: {
      fontFamily: 'Nunito_400Regular',
      fontSize: 13,
      lineHeight: 18,
    },
  } satisfies Record<string, TextStyle>,
  shadow: {
    card: {
      shadowColor: '#1C2529',
      shadowOpacity: 0.06,
      shadowRadius: 16,
      shadowOffset: { width: 0, height: 6 },
      elevation: 2,
      boxShadow: '0 6px 20px rgba(28, 37, 41, 0.06)',
    } as ViewStyle,
    raised: {
      shadowColor: '#085A62',
      shadowOpacity: 0.28,
      shadowRadius: 18,
      shadowOffset: { width: 0, height: 10 },
      elevation: 5,
      boxShadow: '0 10px 28px rgba(8, 90, 98, 0.28)',
    } as ViewStyle,
    soft: {
      shadowColor: '#1C2529',
      shadowOpacity: 0.035,
      shadowRadius: 8,
      shadowOffset: { width: 0, height: 2 },
      elevation: 1,
      boxShadow: '0 2px 10px rgba(28, 37, 41, 0.04)',
    } as ViewStyle,
  },
  layout: {
    compactBreakpoint: 800,
    sidebarWidth: 264,
    contentMax: 720,
    loginMax: 440,
    controlMinHeight: 52,
    controlMinHeightSm: 40,
    inputPadY: 14,
    headerLogo: 34,
    sidebarLogo: 42,
    logoLoading: 72,
    logoAuth: 104,
    logoAuthCompact: 120,
    navIcon: 22,
    avatar: 48,
    avatarLg: 88,
    avatarSm: 40,
    border: 1,
    borderStrong: 1.5,
    heroMinHeight: 112,
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
