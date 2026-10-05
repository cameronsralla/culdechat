import { Logo } from '@/components/ui/Logo';

export function LoadingScreen() {
  return (
    <div className="flex min-h-dvh items-center justify-center bg-paper">
      <Logo size={72} className="animate-pulse" />
    </div>
  );
}
