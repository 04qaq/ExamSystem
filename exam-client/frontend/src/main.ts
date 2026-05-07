import { createApp } from 'vue'
import { createPinia } from 'pinia'
import naive from 'naive-ui'
import App from './App.vue'
import router from './router'
import { useSettingsStore } from '@/stores/settings'
import './style.css'

function bootstrap() {
  const pinia = createPinia()
  const app = createApp(App)
  app.use(pinia)
  app.use(router)
  app.use(naive)
  app.mount('#app')
  // 自动发现可能扫描较多 IP，若在 mount 之前 await，他人电脑上会像「长时间白屏」甚至误以为程序坏了
  const settings = useSettingsStore()
  void settings.ensureDiscoveredOnStartup()
}

bootstrap()
