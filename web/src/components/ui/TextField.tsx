import { forwardRef, useId, type InputHTMLAttributes, type TextareaHTMLAttributes } from 'react';
import { cn } from '@/lib/cn';
import { Text } from './Text';

type Base = {
  label?: string;
  hint?: string;
  error?: string;
};

type InputProps = Base & InputHTMLAttributes<HTMLInputElement> & { multiline?: false };
type AreaProps = Base & TextareaHTMLAttributes<HTMLTextAreaElement> & { multiline: true; rows?: number };
export type TextFieldProps = InputProps | AreaProps;

const FIELD =
  'w-full rounded-sm border bg-raised px-3 text-body text-ink placeholder:text-muted/70 ' +
  'border-line focus:border-brand focus:outline-none focus:ring-2 focus:ring-brand/20 ' +
  'disabled:bg-surface-muted disabled:text-muted';

export const TextField = forwardRef<HTMLInputElement | HTMLTextAreaElement, TextFieldProps>(function TextField(
  props,
  ref,
) {
  const id = useId();
  const { label, hint, error, className, ...rest } = props;
  const describedBy = error ? `${id}-err` : hint ? `${id}-hint` : undefined;

  return (
    <div className={cn('flex flex-col gap-1', className)}>
      {label && (
        <label htmlFor={id}>
          <Text variant="label" tone="muted">
            {label}
          </Text>
        </label>
      )}
      {'multiline' in rest && rest.multiline ? (
        <textarea
          id={id}
          ref={ref as React.Ref<HTMLTextAreaElement>}
          aria-invalid={!!error}
          aria-describedby={describedBy}
          className={cn(FIELD, 'min-h-28 resize-y py-2.5', error && 'border-danger')}
          {...(omit(rest as AreaProps, 'multiline') as TextareaHTMLAttributes<HTMLTextAreaElement>)}
        />
      ) : (
        <input
          id={id}
          ref={ref as React.Ref<HTMLInputElement>}
          aria-invalid={!!error}
          aria-describedby={describedBy}
          className={cn(FIELD, 'min-h-control md:min-h-control-dense', error && 'border-danger')}
          {...(omit(rest as InputProps, 'multiline') as InputHTMLAttributes<HTMLInputElement>)}
        />
      )}
      {error ? (
        <Text id={`${id}-err`} variant="caption" tone="danger" role="alert">
          {error}
        </Text>
      ) : hint ? (
        <Text id={`${id}-hint`} variant="caption" tone="muted">
          {hint}
        </Text>
      ) : null}
    </div>
  );
});

function omit<T extends object, K extends keyof T>(obj: T, key: K): Omit<T, K> {
  const { [key]: _drop, ...rest } = obj;
  void _drop;
  return rest;
}
