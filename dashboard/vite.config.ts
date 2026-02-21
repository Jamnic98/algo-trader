import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tsconfigPaths from 'vite-tsconfig-paths'

import tailwindcss from "@tailwindcss/vite";

const target = process.env.DOCKER ?
'http://trader-core:8080' : 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tsconfigPaths(), tailwindcss()],
  server: {
    host: true,
    allowedHosts: ['ubuntu-server.local', 'http://localhost'],
    proxy: {
      '/api': {
        target,
        changeOrigin: true,
        secure: false
      },
    },
  }
});
