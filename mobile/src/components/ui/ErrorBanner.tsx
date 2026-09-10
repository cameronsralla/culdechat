import { View } from 'react-native';
import { useStyles, type Theme } from '../../theme';
import { AppText } from './AppText';

const stylesFor = (t: Theme) => ({
  box: {
    backgroundColor: t.colors.dangerSoft,
    borderRadius: t.radius.sm,
    padding: t.space.md,
  },
});

export function ErrorBanner({ message }: { message: string | null }) {
  const styles = useStyles(stylesFor);
  if (!message) {
    return null;
  }
  return (
    <View style={styles.box}>
      <AppText variant="label" tone="danger">
        {message}
      </AppText>
    </View>
  );
}
