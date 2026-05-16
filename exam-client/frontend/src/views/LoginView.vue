<script setup lang="ts">
import {
  NCard,
  NTabs,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  useMessage,
} from 'naive-ui'
import { reactive, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { loginApi, registerApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'

const message = useMessage()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const settings = useSettingsStore()

const loading = ref(false)

const discovering = computed(() => settings.isDiscovering)

const loginForm = reactive({
  username: '',
  password: '',
})

const regForm = reactive({
  username: '',
  password: '',
  real_name: '',
})

async function onLogin() {
  loading.value = true
  try {
    const res = await loginApi({
      username: loginForm.username.trim(),
      password: loginForm.password,
    })
    auth.setSession(res.token, res.user)
    message.success(`欢迎，${res.user.real_name || res.user.username}`)
    const redirect = (route.query.redirect as string) || ''
    router.replace(redirect || '/app')
  } catch (e) {
    message.error(e instanceof Error ? e.message : '登录失败')
  } finally {
    loading.value = false
  }
}

async function onRegister() {
  loading.value = true
  try {
    await registerApi({
      username: regForm.username.trim(),
      password: regForm.password,
      real_name: regForm.real_name.trim() || undefined,
    })
    message.success('注册成功，请使用新账号登录（默认为学生角色）')
    loginForm.username = regForm.username.trim()
    loginForm.password = regForm.password
  } catch (e) {
    message.error(e instanceof Error ? e.message : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <n-card class="login-card" title="在线考试系统" size="large">
      <p class="hint">
        <template v-if="discovering">正在确认考试服务连接…</template>
        <template v-else>
          客户端默认连接管理员配置的考试服务。
          <span class="sub-hint">
            若无法登录，请
            <router-link class="link" to="/setup">检查网络设置</router-link>
            。
          </span>
        </template>
      </p>
      <n-tabs type="line" animated>
        <n-tab-pane name="login" tab="登录">
          <n-form label-placement="top">
            <n-form-item label="用户名" required>
              <n-input v-model:value="loginForm.username" placeholder="用户名" @keyup.enter="onLogin" />
            </n-form-item>
            <n-form-item label="密码" required>
              <n-input
                v-model:value="loginForm.password"
                type="password"
                show-password-on="click"
                placeholder="密码"
                @keyup.enter="onLogin"
              />
            </n-form-item>
            <n-button
              type="primary"
              block
              size="large"
              :loading="loading"
              @click="onLogin"
            >
              登录
            </n-button>
          </n-form>
        </n-tab-pane>
        <n-tab-pane name="register" tab="注册">
          <n-form label-placement="top">
            <n-form-item label="用户名" required>
              <n-input v-model:value="regForm.username" placeholder="至少 3 个字符" />
            </n-form-item>
            <n-form-item label="密码" required>
              <n-input
                v-model:value="regForm.password"
                type="password"
                show-password-on="click"
                placeholder="至少 6 位"
              />
            </n-form-item>
            <n-form-item label="姓名（可选）">
              <n-input v-model:value="regForm.real_name" placeholder="真实姓名" />
            </n-form-item>
            <n-button type="primary" block size="large" :loading="loading" @click="onRegister">
              注册
            </n-button>
          </n-form>
        </n-tab-pane>
      </n-tabs>
      <template #footer>
        <n-space justify="center">
          <span class="muted">当前服务：{{ settings.effectiveBaseUrl }}</span>
        </n-space>
      </template>
    </n-card>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: linear-gradient(145deg, #f0f4ff 0%, #e8f5f0 100%);
}
.login-card {
  width: 100%;
  max-width: 420px;
  border-radius: 12px;
}
.hint {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 16px;
}
.sub-hint {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: #94a3b8;
}
.link {
  color: #18a058;
}
.muted {
  font-size: 12px;
  color: #94a3b8;
}
</style>
