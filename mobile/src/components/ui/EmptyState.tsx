import { AppText } from './AppText';
import { Button } from './Button';
import { Stack } from './Stack';

type Props = {
  title: string;
  actionLabel?: string;
  onAction?: () => void;
};

export function EmptyState({ title, actionLabel, onAction }: Props) {
  return (
    <Stack gap="md">
      <AppText tone="muted">{title}</AppText>
      {actionLabel && onAction ? <Button label={actionLabel} variant="ghost" onPress={onAction} /> : null}
    </Stack>
  );
}
