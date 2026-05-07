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
import { useSettingsStore } from '@/stores/settings'
import { probeExamServerBase } from '@/utils/serverDiscovery'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const settings = useSettingsStore()

const draft = ref(settings.apiBaseUrl || settings.effectiveBaseUrl)
const testing = ref(false)
const discovering = ref(false)

async function testConnection() {
  testing.value = true
  const base = draft.value.trim().replace(/\/+$/, '') || settings.effectiveBaseUrl
  try {
    const ok = await probeExamServerBase(base)
    if (ok) message.success('已连通 exam-server')
    else message.error('无法识别为 exam-server，请核对地址与端口')
  } catch {
    message.error('无法连接，请检查地址、端口与防火墙')
  } finally {
    testing.value = false
  }
}

async function autoDiscover() {
  discovering.value = true
  try {
    const url = await settings.discoverNow()
    if (url) {
      draft.value = url
      message.success(`已找到服务端：${url}`)
    } else {
      message.warning(settings.lastDiscoveryMessage || '未发现局域网内的 exam-server')
    }
  } finally {
    discovering.value = false
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
    <n-alert title="连接说明" type="info">
      首次启动会自动探测<strong>本机</strong>（127.0.0.1 / localhost）、<strong>局域网</strong>与
      <strong>Wi‑Fi</strong>网段内的 exam-server（通过公开接口 <code>/api/ping</code>）。若设置了环境变量
      <code>VITE_API_BASE_URL</code>，则以环境变量为准且跳过自动发现。服务端需放行 TCP
      <code>8080</code>（或您映射的端口）；同一热点/路由器下的设备通常可直接被发现。
    </n-alert>
    <n-card title="API 根地址">
      <n-form label-placement="top">
        <n-form-item label="例如 http://192.168.1.10:8080（不要以 / 结尾）">
          <n-input v-model:value="draft" placeholder="留空则按自动发现或默认本机 8080" />
        </n-form-item>
        <n-space>
          <n-button @click="autoDiscover" :loading="discovering || settings.isDiscovering">自动查找服务端</n-button>
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
