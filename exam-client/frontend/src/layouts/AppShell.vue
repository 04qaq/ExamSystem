<script setup lang="ts">
import {
  NLayout,
  NLayoutHeader,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NButton,
  NSpace,
  NBreadcrumb,
  NBreadcrumbItem,
  useMessage,
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { computed, h, inject, ref, watch, type Ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { RoleAdmin, RoleTeacher, RoleStudent } from '@/api/types'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const message = useMessage()

const collapsed = ref(false)
const toggleTheme = inject<() => void>('toggleTheme')
const isDark = inject<Ref<boolean>>('isDark')

function renderIcon(label: string) {
  return () => h('span', { class: 'nav-ico' }, label)
}

const menuOptions = computed<MenuOption[]>(() => {
  const r = auth.role
  const opts: MenuOption[] = [
    {
      label: () =>
        h(
          RouterLink,
          { to: { name: 'settings' } },
          { default: () => '网络设置' }
        ),
      key: 'settings',
      icon: renderIcon('⚙'),
    },
  ]
  if (r === RoleStudent) {
    opts.unshift(
      {
        label: () =>
          h(RouterLink, { to: { name: 'student-papers' } }, { default: () => '可考试卷' }),
        key: 'student-papers',
        icon: renderIcon('📝'),
      },
      {
        label: () =>
          h(RouterLink, { to: { name: 'student-records' } }, { default: () => '我的成绩' }),
        key: 'student-records',
        icon: renderIcon('📊'),
      }
    )
  }
  if (r === RoleTeacher) {
    opts.unshift(
      {
        label: () =>
          h(RouterLink, { to: { name: 'teacher-papers' } }, { default: () => '试卷管理' }),
        key: 'teacher-papers',
        icon: renderIcon('📄'),
      },
      {
        label: () =>
          h(RouterLink, { to: { name: 'teacher-questions' } }, { default: () => '题库' }),
        key: 'teacher-questions',
        icon: renderIcon('❓'),
      },
      {
        label: () =>
          h(RouterLink, { to: { name: 'teacher-grading' } }, { default: () => '批阅' }),
        key: 'teacher-grading',
        icon: renderIcon('✏'),
      }
    )
  }
  if (r === RoleAdmin) {
    opts.unshift(
      {
        label: () =>
          h(RouterLink, { to: { name: 'admin-users' } }, { default: () => '用户管理' }),
        key: 'admin-users',
        icon: renderIcon('👥'),
      },
      {
        label: () =>
          h(RouterLink, { to: { name: 'admin-logs' } }, { default: () => '操作日志' }),
        key: 'admin-logs',
        icon: renderIcon('📜'),
      }
    )
  }
  return opts
})

const routeMenuKey: Record<string, string> = {
  'student-papers': 'student-papers',
  'student-exam': 'student-papers',
  'student-records': 'student-records',
  'student-record-detail': 'student-records',
  'teacher-papers': 'teacher-papers',
  'teacher-questions': 'teacher-questions',
  'teacher-grading': 'teacher-grading',
  'teacher-statistics': 'teacher-papers',
  'admin-users': 'admin-users',
  'admin-logs': 'admin-logs',
  settings: 'settings',
}

const activeKey = computed(() => routeMenuKey[route.name?.toString() || ''] || '')

watch(
  () => route.fullPath,
  () => {
    if (route.meta.roles && auth.role != null) {
      const roles = route.meta.roles as number[]
      if (!roles.includes(auth.role)) {
        message.warning('当前账号无权访问该页面')
        router.replace({ name: 'home' })
      }
    }
  },
  { immediate: true }
)

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}

const titleMap: Record<string, string> = {
  home: '首页',
  settings: '网络设置',
  'student-papers': '可考试卷',
  'student-exam': '考试中',
  'student-records': '我的成绩',
  'student-record-detail': '成绩详情',
  'teacher-papers': '试卷管理',
  'teacher-questions': '题库',
  'teacher-grading': '批阅',
  'teacher-statistics': '试卷统计',
  'admin-users': '用户管理',
  'admin-logs': '操作日志',
}

const breadcrumbLabel = computed(() => titleMap[route.name?.toString() || ''] || '')
</script>

<template>
  <n-layout position="absolute" style="height: 100vh">
    <n-layout-header bordered style="height: 56px; padding: 0 20px; display: flex; align-items: center; justify-content: space-between">
      <div style="font-weight: 600; font-size: 16px">在线考试系统</div>
      <n-space align="center">
        <span style="opacity: 0.85">{{ auth.user?.real_name || auth.user?.username }}</span>
        <n-button quaternary size="small" @click="toggleTheme?.()">
          {{ isDark?.value ? '浅色' : '深色' }}
        </n-button>
        <n-button secondary size="small" @click="logout">退出登录</n-button>
      </n-space>
    </n-layout-header>
    <n-layout has-sider position="absolute" style="top: 56px; bottom: 0">
      <n-layout-sider
        bordered
        collapse-mode="width"
        :collapsed-width="64"
        :width="220"
        show-trigger
        v-model:collapsed="collapsed"
        content-style="padding: 12px 8px"
      >
        <n-menu
          :collapsed="collapsed"
          :collapsed-width="64"
          :options="menuOptions"
          :value="activeKey"
        />
      </n-layout-sider>
      <n-layout content-style="padding: 20px 24px">
        <n-breadcrumb style="margin-bottom: 16px">
          <n-breadcrumb-item>主页</n-breadcrumb-item>
          <n-breadcrumb-item v-if="breadcrumbLabel">{{ breadcrumbLabel }}</n-breadcrumb-item>
        </n-breadcrumb>
        <n-layout-content>
          <router-view />
        </n-layout-content>
      </n-layout>
    </n-layout>
  </n-layout>
</template>

<style scoped>
.nav-ico {
  display: inline-flex;
  width: 1.25rem;
  justify-content: center;
}
</style>
