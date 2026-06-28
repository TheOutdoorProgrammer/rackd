import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// In dev, run `npm run dev` (Vite on :5173) alongside the Go server (:8080).
// API calls are proxied to the backend so there is no CORS to manage.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
