import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { RoleAdmin, RoleTeacher, RoleStudent } from '@/api/types'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/setup',
      name: 'setup',
      component: () => import('@/views/SettingsView.vue'),
    },
    {
      path: '/app',
      component: () => import('@/layouts/AppShell.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/views/HomeRedirect.vue'),
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/SettingsView.vue'),
        },
        {
          path: 'student/papers',
          name: 'student-papers',
          component: () => import('@/views/student/StudentPapersView.vue'),
          meta: { roles: [RoleStudent] },
        },
        {
          path: 'student/exam/:paperId',
          name: 'student-exam',
          component: () => import('@/views/student/StudentExamView.vue'),
          meta: { roles: [RoleStudent] },
        },
        {
          path: 'student/records',
          name: 'student-records',
          component: () => import('@/views/student/StudentRecordsView.vue'),
          meta: { roles: [RoleStudent] },
        },
        {
          path: 'student/records/:id',
          name: 'student-record-detail',
          component: () => import('@/views/student/StudentRecordDetailView.vue'),
          meta: { roles: [RoleStudent] },
        },
        {
          path: 'teacher/papers',
          name: 'teacher-papers',
          component: () => import('@/views/teacher/TeacherPapersView.vue'),
          meta: { roles: [RoleTeacher] },
        },
        {
          path: 'teacher/questions',
          name: 'teacher-questions',
          component: () => import('@/views/teacher/TeacherQuestionsView.vue'),
          meta: { roles: [RoleTeacher] },
        },
        {
          path: 'teacher/grading',
          name: 'teacher-grading',
          component: () => import('@/views/teacher/TeacherGradingView.vue'),
          meta: { roles: [RoleTeacher] },
        },
        {
          path: 'teacher/statistics/:paperId',
          name: 'teacher-statistics',
          component: () => import('@/views/teacher/TeacherStatisticsView.vue'),
          meta: { roles: [RoleTeacher] },
        },
        {
          path: 'admin/users',
          name: 'admin-users',
          component: () => import('@/views/admin/AdminUsersView.vue'),
          meta: { roles: [RoleAdmin] },
        },
        {
          path: 'admin/logs',
          name: 'admin-logs',
          component: () => import('@/views/admin/AdminLogsView.vue'),
          meta: { roles: [RoleAdmin] },
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.guest && auth.isLoggedIn) {
    next({ name: 'home' })
    return
  }
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }
  const roles = to.meta.roles as number[] | undefined
  if (roles?.length && auth.role != null && !roles.includes(auth.role)) {
    next({ name: 'home' })
    return
  }
  next()
})

export default router
