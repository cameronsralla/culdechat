import type { HTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Props = HTMLAttributes<HTMLDivElement> & { tone?: 'surface' | 'wash' | 'pin' | 'raised'; padded?: boolean };

const TONE = {
  surface: 'bg-surface border-line',
  raised: 'bg-raised border-line',
  wash: 'bg-raised border-line border-l-[3px] border-l-brand',
  pin: 'bg-pin border-pin-line',
};

/** Panel container. Hairline border + surface step — no default drop shadow. */
export function Card({ tone = 'surface', padded = true, className, ...rest }: Props) {
  return <div className={cn('rounded-md border', TONE[tone], padded && 'p-4 md:p-4', className)} {...rest} />;
}
