import path from "path";

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import tsconfigPaths from 'vite-tsconfig-paths'

const target = process.env.DOCKER ?
'http://trader-core:8080' : 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
    resolve: {
    alias: {
      components: path.resolve(__dirname, './src/components'),
      context: path.resolve(__dirname, './src/context'),
      layouts: path.resolve(__dirname, './src/layouts'),
      pages: path.resolve(__dirname, './src/pages'),
    },
  },
  plugins: [react(), tsconfigPaths(), tailwindcss()],
  server: {
      host: true,
      proxy: {
          '/api': {
              target,
              ws: true,
              changeOrigin: true,
              headers: {
                  'X-Accel-Buffering': 'no',
              },
          },
      }
  }
});
