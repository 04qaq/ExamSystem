import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    // 可选：开发时代理到本地 exam-server（不设 VITE_API_BASE_URL 时用前端内置配置）
    proxy: {},
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
