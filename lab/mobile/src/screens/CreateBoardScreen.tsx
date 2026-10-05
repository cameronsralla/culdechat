import { useState } from 'react';
import { useRouter } from 'expo-router';
import { createBoard } from '../api/community';
import { AppText, Button, ErrorBanner, Stack, TextField } from '../components/ui';
import { Screen } from '../components/layout/Screen';

export function CreateBoardScreen() {
  const router = useRouter();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit() {
    setError(null);
    setLoading(true);
    try {
      const board = await createBoard(name.trim(), description.trim());
      router.replace(`/boards/${board.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create board');
    } finally {
      setLoading(false);
    }
  }

  return (
    <Screen scroll inShell>
      <Stack gap="lg">
        <AppText variant="title">New board</AppText>
        <ErrorBanner message={error} />
        <TextField label="Name" value={name} onChangeText={setName} />
        <TextField label="Description" value={description} onChangeText={setDescription} multiline />
        <Button label="Create board" onPress={() => void onSubmit()} loading={loading} disabled={!name.trim()} />
        <Button label="Cancel" variant="ghost" onPress={() => router.back()} />
      </Stack>
    </Screen>
  );
}
