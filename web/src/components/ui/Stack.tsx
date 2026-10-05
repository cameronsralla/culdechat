import type { HTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Gap = 1 | 2 | 3 | 4 | 6 | 8;
const GAP: Record<Gap, string> = { 1: 'gap-1', 2: 'gap-2', 3: 'gap-3', 4: 'gap-4', 6: 'gap-6', 8: 'gap-8' };

type Props = HTMLAttributes<HTMLDivElement> & { gap?: Gap; row?: boolean; align?: 'start' | 'center' | 'end' | 'stretch'; justify?: 'start' | 'center' | 'end' | 'between' };

const ALIGN = { start: 'items-start', center: 'items-center', end: 'items-end', stretch: 'items-stretch' };
const JUSTIFY = { start: 'justify-start', center: 'justify-center', end: 'justify-end', between: 'justify-between' };

/** Flex layout helper: vertical by default, `row` for horizontal. */
export function Stack({ gap = 3, row, align, justify, className, ...rest }: Props) {
  return (
    <div
      className={cn('flex', row ? 'flex-row' : 'flex-col', GAP[gap], align && ALIGN[align], justify && JUSTIFY[justify], className)}
      {...rest}
    />
  );
}
