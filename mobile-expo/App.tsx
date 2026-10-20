import { StatusBar } from "expo-status-bar";
import { useState } from "react";
import { Button, StyleSheet, Text, View } from "react-native";

export default function App() {
  const [flexibility, setFlexibility] = useState(true);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Sticker Mule Mobile — Expo</Text>
      <Text>
        React Native Expo TestFlight OTA push — single codebase web+mobile
        FLEXIBILITY {String(flexibility)} — same repo as frontend-react
      </Text>
      <Text>9+ years Tommy Shamamba — 7M+ users — 5M/day — 99.99%</Text>
      <Button
        title={`FLEXIBILITY ${flexibility ? "ON" : "OFF"}`}
        onPress={() => setFlexibility((value) => !value)}
      />
      <StatusBar style="auto" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    alignItems: "center",
    justifyContent: "center",
    padding: 24,
    gap: 12,
  },
  title: { fontSize: 22, fontWeight: "700" },
});
