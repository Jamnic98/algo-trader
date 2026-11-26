import { defineConfig } from "vitest/config";
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  root: ".",
  plugins: [tsconfigPaths()],
  esbuild: {
    // keep the automatic JSX runtime so component tests don't need `import React`
    tsconfigRaw: '{"compilerOptions":{"jsx":"react-jsx"}}',
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./vitest.setup.ts",
    exclude: ["node_modules/**"],
  },
});
