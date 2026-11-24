import { defineConfig } from "vitest/config";
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  root: ".",
  plugins: [tsconfigPaths()],
  esbuild: {
    tsconfigRaw: "{}",
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./vitest.setup.ts",
    exclude: ["node_modules/**"],
  },
});
