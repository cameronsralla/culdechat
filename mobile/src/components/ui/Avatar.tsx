import { useEffect, useMemo, useState } from 'react';
import { Image, View } from 'react-native';
import { getAccessToken, mediaUrl } from '../../api/client';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppText } from './AppText';

type Size = 'sm' | 'md' | 'lg';

type Props = {
  path?: string | null;
  label?: string;
  size?: Size;
};

const palette = ['#0F8A96', '#2A7B6F', '#3D6B8A', '#6B7A4E', '#8A6B4E', '#5E6A91'] as const;

function colorFor(label?: string): string {
  const raw = (label || '?').trim().toLowerCase();
  let hash = 0;
  for (let i = 0; i < raw.length; i += 1) {
    hash = (hash * 31 + raw.charCodeAt(i)) >>> 0;
  }
  return palette[hash % palette.length];
}

const stylesFor = (t: Theme) => ({
  sm: {
    width: t.layout.avatarSm,
    height: t.layout.avatarSm,
    borderRadius: t.layout.avatarSm / 2,
  },
  md: {
    width: t.layout.avatar,
    height: t.layout.avatar,
    borderRadius: t.layout.avatar / 2,
  },
  lg: {
    width: t.layout.avatarLg,
    height: t.layout.avatarLg,
    borderRadius: t.layout.avatarLg / 2,
  },
  base: {
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    overflow: 'hidden' as const,
    borderWidth: t.layout.border,
    borderColor: t.colors.white,
  },
  image: { width: '100%' as const, height: '100%' as const },
});

export function Avatar({ path, label, size = 'md' }: Props) {
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  const [uri, setUri] = useState<string | null>(null);
  const initial = (label?.trim()?.[0] || '?').toUpperCase();
  const bg = useMemo(() => colorFor(label), [label]);
  const box = size === 'lg' ? styles.lg : size === 'sm' ? styles.sm : styles.md;

  useEffect(() => {
    let objectUrl: string | null = null;
    let cancelled = false;
    if (!path) {
      setUri(null);
      return;
    }
    const url = mediaUrl(path);
    const token = getAccessToken();
    fetch(url, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
      .then((res) => {
        if (!res.ok) {
          throw new Error('photo');
        }
        return res.blob();
      })
      .then((blob) => {
        objectUrl = URL.createObjectURL(blob);
        if (!cancelled) {
          setUri(objectUrl);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setUri(null);
        }
      });
    return () => {
      cancelled = true;
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
      }
    };
  }, [path]);

  return (
    <View style={[styles.base, box, { backgroundColor: uri ? theme.colors.brandSoft : bg }]}>
      {uri ? (
        <Image source={{ uri }} style={styles.image} />
      ) : (
        <AppText
          variant={size === 'lg' ? 'title' : size === 'sm' ? 'caption' : 'label'}
          tone="white"
        >
          {initial}
        </AppText>
      )}
    </View>
  );
}
