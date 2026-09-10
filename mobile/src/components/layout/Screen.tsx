import { type ReactNode, useRef } from 'react';
import { RefreshControl, ScrollView, View, type NativeScrollEvent, type NativeSyntheticEvent } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useStyles, useTheme, type Theme } from '../../theme';
import { useCompactLayout } from './useCompactLayout';

type Props = {
  scroll?: boolean;
  padded?: boolean;
  inShell?: boolean;
  onEndReached?: () => void;
  refreshing?: boolean;
  onRefresh?: () => void;
  children: ReactNode;
};

const stylesFor = (t: Theme) => ({
  fill: { flex: 1, backgroundColor: t.colors.paper },
  grow: { flexGrow: 1 },
});

export function Screen({
  scroll,
  padded = true,
  inShell = false,
  onEndReached,
  refreshing = false,
  onRefresh,
  children,
}: Props) {
  const insets = useSafeAreaInsets();
  const compact = useCompactLayout();
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  const locked = useRef(false);
  const pad = {
    paddingBottom: theme.space.lg,
    paddingHorizontal: padded ? (compact ? theme.space.xl : theme.space.xxxl) : 0,
    paddingTop: inShell ? theme.space.lg : Math.max(insets.top, theme.space.lg),
    maxWidth: inShell && !compact ? theme.layout.contentMax : undefined,
    width: '100%' as const,
    alignSelf: 'center' as const,
  };

  function maybeEnd(e: NativeSyntheticEvent<NativeScrollEvent>) {
    if (!onEndReached) {
      return;
    }
    const { layoutMeasurement, contentOffset, contentSize } = e.nativeEvent;
    if (layoutMeasurement.height + contentOffset.y >= contentSize.height - 160) {
      if (!locked.current) {
        locked.current = true;
        onEndReached();
        setTimeout(() => {
          locked.current = false;
        }, 800);
      }
    }
  }

  if (scroll) {
    return (
      <ScrollView
        style={styles.fill}
        contentContainerStyle={[styles.grow, pad]}
        keyboardShouldPersistTaps="handled"
        onScroll={maybeEnd}
        scrollEventThrottle={80}
        refreshControl={
          onRefresh ? (
            <RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={theme.colors.brand} />
          ) : undefined
        }
      >
        {children}
      </ScrollView>
    );
  }
  return <View style={[styles.fill, pad]}>{children}</View>;
}
