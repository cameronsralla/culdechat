import { clsx, type ClassValue } from 'clsx';
import { extendTailwindMerge } from 'tailwind-merge';

/**
 * Tailwind-aware class merging. Our type scale uses custom names
 * (text-title, text-body, ...) which tailwind-merge would otherwise treat as
 * colors and drop when combined with text-ink etc. Teach it our scale.
 */
const twMerge = extendTailwindMerge({
  extend: {
    theme: {
      text: ['display', 'title', 'subtitle', 'body', 'label', 'caption', 'mono'],
    },
  },
});

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
