<script setup lang="ts">
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NAlert,
  useMessage,
} from 'naive-ui'
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { useSettingsStore } from '@/stores/settings'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const settings = useSettingsStore()

const draft = ref(settings.apiBaseUrl || settings.effectiveBaseUrl)
const testing = ref(false)

async function testConnection() {
  testing.value = true
  const base = draft.value.trim().replace(/\/+$/, '') || settings.effectiveBaseUrl
  try {
    await axios.post(
      `${base}/api/auth/login`,
      {},
      {
        headers: { 'Content-Type': 'application/json' },
        timeout: 8000,
        validateStatus: () => true,
      }
    )
    message.success('已连通服务端（返回了 API 响应，说明地址正确）')
  } catch {
    message.error('无法连接，请检查地址、端口与防火墙')
  } finally {
    testing.value = false
  }
}

function save() {
  settings.setApiBaseUrl(draft.value)
  message.success('已保存，后续请求将使用新地址')
}
</script>

<template>
  <n-space vertical size="large">
    <div v-if="route.name === 'setup'">
      <n-button quaternary size="small" @click="router.push('/login')">← 返回登录</n-button>
    </div>
    <n-alert title="前后端分离" type="info">
      本客户端仅通过 HTTP 调用 exam-server，可部署在不同机器。默认回落地址来自环境变量
      <code>VITE_API_BASE_URL</code>，未设置时为 <code>http://127.0.0.1:8080</code>。
    </n-alert>
    <n-card title="API 根地址">
      <n-form label-placement="top">
        <n-form-item label="例如 http://192.168.1.10:8080（不要以 / 结尾）">
          <n-input v-model:value="draft" placeholder="http://主机:端口" />
        </n-form-item>
        <n-space>
          <n-button @click="testConnection" :loading="testing">测试连接</n-button>
          <n-button type="primary" @click="save">保存</n-button>
        </n-space>
      </n-form>
      <p class="muted">当前生效：<strong>{{ settings.effectiveBaseUrl }}</strong></p>
    </n-card>
  </n-space>
</template>

<style scoped>
.muted {
  margin-top: 16px;
  font-size: 13px;
  color: #64748b;
}
code {
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.06);
}
</style>
