import path from "node:path";
import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/lib/__tests__/setup.ts"],
    coverage: {
      provider: "v8",
      thresholds: {
        lines: 70,
        functions: 70,
        branches: 60,
      },
      exclude: ["src/mocks/**", "**/*.test.*", "src/app/**"],
      reporter: ["text", "json-summary"],
    },
  },
});
