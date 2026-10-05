import { useState } from 'react';
import {
  TextInput,
  View,
  type NativeSyntheticEvent,
  type TextInputChangeEventData,
  type TextInputProps,
} from 'react-native';
import { useStyles, useTheme, type Theme } from '../../theme';
import { AppText } from './AppText';

type Props = TextInputProps & {
  label: string;
};

const stylesFor = (t: Theme) => ({
  wrap: { gap: t.space.xs },
  label: { marginLeft: t.space.xs },
  input: {
    ...t.type.body,
    color: t.colors.ink,
    backgroundColor: t.colors.white,
    borderWidth: t.layout.border,
    borderColor: t.colors.line,
    borderRadius: t.radius.md,
    paddingHorizontal: t.space.lg,
    paddingVertical: t.layout.inputPadY,
  },
  inputFocused: {
    borderColor: t.colors.brand,
    backgroundColor: t.colors.surface,
    ...t.shadow.soft,
  },
  multiline: {
    minHeight: 120,
    textAlignVertical: 'top' as const,
  },
});

export function TextField({ label, style, onChangeText, onChange, ...rest }: Props) {
  const theme = useTheme();
  const styles = useStyles(stylesFor);
  const [focused, setFocused] = useState(false);

  function handleChange(e: NativeSyntheticEvent<TextInputChangeEventData>) {
    onChange?.(e);
    // Bridge for RN Web when onChangeText is flaky.
    if (onChangeText) {
      const text = e.nativeEvent?.text;
      if (typeof text === 'string') {
        onChangeText(text);
      }
    }
  }

  return (
    <View style={styles.wrap}>
      <AppText variant="label" tone="muted" style={styles.label}>
        {label}
      </AppText>
      <TextInput
        {...rest}
        accessibilityLabel={label}
        onChange={handleChange}
        onFocus={(e) => {
          setFocused(true);
          rest.onFocus?.(e);
        }}
        onBlur={(e) => {
          setFocused(false);
          rest.onBlur?.(e);
        }}
        placeholderTextColor={theme.colors.muted}
        style={[
          styles.input,
          { outlineStyle: 'none' } as object,
          rest.multiline ? styles.multiline : null,
          focused ? styles.inputFocused : null,
          style,
        ]}
      />
    </View>
  );
}
