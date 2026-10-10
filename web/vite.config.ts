import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// The dev server proxies to a running stack (docker compose publishes the web
// proxy on 127.0.0.1:8080 by default) so cookies and paths match production.
const apiTarget = process.env.VPSENTINEL_API_URL ?? 'http://127.0.0.1:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': apiTarget,
      '/grafana': { target: apiTarget, ws: true },
      '^/(healthz|readyz|version)$': apiTarget,
    },
  },
})
