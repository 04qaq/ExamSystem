<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { RoleAdmin, RoleTeacher, RoleStudent } from '@/api/types'

const router = useRouter()
const auth = useAuthStore()

onMounted(() => {
  const r = auth.role
  if (r === RoleStudent) router.replace({ name: 'student-papers' })
  else if (r === RoleTeacher) router.replace({ name: 'teacher-papers' })
  else if (r === RoleAdmin) router.replace({ name: 'admin-users' })
  else router.replace({ name: 'settings' })
})
</script>

<template>
  <div />
</template>
