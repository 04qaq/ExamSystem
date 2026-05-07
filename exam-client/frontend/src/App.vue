<script setup lang="ts">
import {
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  zhCN,
  dateZhCN,
  darkTheme,
} from 'naive-ui'
import { computed, provide, ref } from 'vue'

const prefersDark = ref(
  typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-color-scheme: dark)').matches
)

const theme = computed(() => (prefersDark.value ? darkTheme : null))

function toggleTheme() {
  prefersDark.value = !prefersDark.value
}

provide('toggleTheme', toggleTheme)
provide('isDark', prefersDark)
</script>

<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN" :theme="theme">
    <n-message-provider>
      <n-dialog-provider>
        <router-view />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>
