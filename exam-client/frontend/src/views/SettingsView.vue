<script setup lang="ts">
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NAlert,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { probeExamServerBase } from '@/utils/serverDiscovery'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const settings = useSettingsStore()

/** 仅展示用户填写内容；留空表示沿用自动检测与本机默认 */
const draft = ref(settings.apiBaseUrl.trim())
const testing = ref(false)
const discovering = ref(false)
const currentBaseUrl = computed(() => settings.effectiveBaseUrl)
const defaultBaseUrl = computed(() => settings.defaultBaseUrl)
const usingManual = computed(() => !!settings.apiBaseUrl.trim())

async function testConnection() {
  testing.value = true
  const base = draft.value.trim().replace(/\/+$/, '') || settings.effectiveBaseUrl
  try {
    const ok = await probeExamServerBase(base)
    if (ok) message.success('连接成功，考试服务可用')
    else message.error('无法识别为考试服务，请核对地址与端口')
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
      message.success('已找到考试服务，地址已填入上方')
    } else {
      message.warning(settings.lastDiscoveryMessage || '未发现局域网内的考试服务')
    }
  } finally {
    discovering.value = false
  }
}

function save() {
  settings.setApiBaseUrl(draft.value)
  message.success('已保存，后续请求将使用新地址')
}

function resetToDefault() {
  settings.resetApiBaseUrl()
  draft.value = ''
  message.success('已恢复为打包默认服务地址')
}
</script>

<template>
  <n-space vertical size="large">
    <div v-if="route.name === 'setup'">
      <n-button quaternary size="small" @click="router.push('/login')">← 返回登录</n-button>
    </div>
    <n-alert title="连接说明" type="info">
      客户端默认连接管理员配置的考试服务。只有在局域网临时考试环境中，才需要使用“自动查找”。
      若无法登录，请确认下方“当前实际连接地址”可以访问，或填写管理员提供的服务地址（含协议与端口，末尾不要加 <code>/</code>）。
    </n-alert>
    <n-card title="考试服务地址">
      <n-space vertical size="small" class="current-url">
        <div>
          <span class="muted">当前实际连接地址：</span>
          <n-tag type="info">{{ currentBaseUrl }}</n-tag>
        </div>
        <div>
          <span class="muted">打包默认地址：</span>
          <code>{{ defaultBaseUrl }}</code>
        </div>
        <div v-if="usingManual" class="muted">
          当前使用的是本机保存的手动地址；若新版客户端仍连接旧服务器，请恢复默认地址。
        </div>
      </n-space>
      <n-form label-placement="top">
        <n-form-item label="服务根地址（可选）">
          <n-input v-model:value="draft" placeholder="留空时使用打包默认服务地址" />
        </n-form-item>
        <n-space>
          <n-button @click="autoDiscover" :loading="discovering || settings.isDiscovering">自动查找</n-button>
          <n-button @click="testConnection" :loading="testing">测试连接</n-button>
          <n-button @click="resetToDefault">恢复默认地址</n-button>
          <n-button type="primary" @click="save">保存</n-button>
        </n-space>
      </n-form>
      <p class="muted">保存后生效；留空时使用打包默认地址。自动查找仅适用于同一局域网内的考试服务。</p>
    </n-card>
  </n-space>
</template>

<style scoped>
.muted {
  margin-top: 16px;
  font-size: 13px;
  color: #64748b;
}
.current-url {
  margin-bottom: 16px;
}
.current-url .muted {
  margin-top: 0;
}
code {
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.06);
}
</style>
