import { Image, StyleSheet, View } from 'react-native';

const logo = require('../../../assets/logo.jpg');

export function Logo({ size = 96 }: { size?: number }) {
  return (
    <View style={[styles.wrap, { width: size, height: size }]}>
      <Image source={logo} style={{ width: size, height: size }} resizeMode="contain" />
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { alignItems: 'center', justifyContent: 'center' },
});
