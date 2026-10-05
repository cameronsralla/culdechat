import { cn } from '@/lib/cn';

export function Logo({ size = 40, className }: { size?: number; className?: string }) {
  return (
    <img
      src="/logo.jpg"
      alt="Cul-de-Chat"
      width={size}
      height={size}
      className={cn('rounded-sm object-cover', className)}
      style={{ width: size, height: size }}
    />
  );
}
