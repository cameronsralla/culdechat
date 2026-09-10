import { useEffect, useState } from 'react';
import { Image, View } from 'react-native';
import { getAccessToken, mediaUrl } from '../../api/client';
import { useStyles, type Theme } from '../../theme';
import { AppText } from './AppText';

type Size = 'sm' | 'lg';

type Props = {
  path?: string | null;
  label?: string;
  size?: Size;
};

const stylesFor = (t: Theme) => ({
  sm: {
    width: t.layout.avatar,
    height: t.layout.avatar,
    borderRadius: t.layout.avatar / 2,
    backgroundColor: t.colors.brandSoft,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    overflow: 'hidden' as const,
  },
  lg: {
    width: t.layout.avatarLg,
    height: t.layout.avatarLg,
    borderRadius: t.layout.avatarLg / 2,
    backgroundColor: t.colors.brandSoft,
    borderWidth: t.layout.border,
    borderColor: t.colors.brandLine,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
    overflow: 'hidden' as const,
  },
  image: { width: '100%' as const, height: '100%' as const },
});

export function Avatar({ path, label, size = 'sm' }: Props) {
  const styles = useStyles(stylesFor);
  const [uri, setUri] = useState<string | null>(null);
  const initial = (label?.trim()?.[0] || '?').toUpperCase();

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
    <View style={size === 'lg' ? styles.lg : styles.sm}>
      {uri ? (
        <Image source={{ uri }} style={styles.image} />
      ) : (
        <AppText variant={size === 'lg' ? 'title' : 'label'} tone="brand">
          {initial}
        </AppText>
      )}
    </View>
  );
}
