<script setup lang="ts">
import {
  NCard,
  NDataTable,
  NPagination,
  NSpace,
  NButton,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { h, onMounted, reactive, ref } from 'vue'
import type { UserRow } from '@/api/types'
import { adminCreateUser, adminUpdateUser, adminUsers } from '@/api/admin'

const message = useMessage()

const loading = ref(false)
const items = ref<UserRow[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const showCreate = ref(false)
const showEdit = ref(false)
const creating = ref(false)
const updating = ref(false)
const createForm = reactive({
  username: '',
  password: '',
  role: 3,
  real_name: '',
})
const editForm = reactive({
  id: 0,
  username: '',
  password: '',
  role: 3,
  real_name: '',
  status: 1,
})

const roleOpts = [
  { label: '管理员', value: 1 },
  { label: '教师', value: 2 },
  { label: '学生', value: 3 },
]

const roleLabel = (r: number) => roleOpts.find((x) => x.value === r)?.label || String(r)

const columns: DataTableColumns<UserRow> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '用户名', key: 'username', width: 140 },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render(row) {
      return roleLabel(row.role)
    },
  },
  { title: '姓名', key: 'real_name', width: 120 },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render(row) {
      return row.status === 1 ? '正常' : '禁用'
    },
  },
  {
    title: '操作',
    key: 'act',
    width: 160,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, { default: () => '编辑' }),
          h(
            NButton,
            {
              size: 'tiny',
              onClick: () => toggleStatus(row),
            },
            { default: () => (row.status === 1 ? '禁用' : '启用') }
          ),
        ],
      })
    },
  },
]

async function load() {
  loading.value = true
  try {
    const res = await adminUsers(page.value, pageSize.value)
    items.value = res.items
    total.value = res.total
  } catch (e) {
    message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function openEdit(row: UserRow) {
  editForm.id = row.id
  editForm.username = row.username
  editForm.password = ''
  editForm.role = row.role
  editForm.real_name = row.real_name || ''
  editForm.status = row.status
  showEdit.value = true
}

async function toggleStatus(row: UserRow) {
  try {
    await adminUpdateUser(row.id, { status: row.status === 1 ? 0 : 1 })
    message.success('已更新')
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '操作失败')
  }
}

async function updateUser() {
  updating.value = true
  try {
    await adminUpdateUser(editForm.id, {
      real_name: editForm.real_name.trim(),
      role: editForm.role,
      status: editForm.status,
      password: editForm.password || undefined,
    })
    message.success('用户已更新')
    showEdit.value = false
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '更新失败')
  } finally {
    updating.value = false
  }
}

async function createUser() {
  creating.value = true
  try {
    await adminCreateUser({
      username: createForm.username.trim(),
      password: createForm.password,
      role: createForm.role,
      real_name: createForm.real_name.trim() || undefined,
    })
    message.success('用户已创建')
    showCreate.value = false
    createForm.username = ''
    createForm.password = ''
    createForm.real_name = ''
    createForm.role = 3
    await load()
  } catch (e) {
    message.error(e instanceof Error ? e.message : '创建失败')
  } finally {
    creating.value = false
  }
}

onMounted(load)
</script>

<template>
  <n-card title="用户管理">
    <n-space style="margin-bottom: 16px">
      <n-button type="primary" @click="showCreate = true">新建用户</n-button>
    </n-space>
    <n-data-table :columns="columns" :data="items" :loading="loading" />
    <div style="margin-top: 16px; display: flex; justify-content: flex-end">
      <n-pagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="total"
        show-size-picker
        :page-sizes="[10, 20, 50]"
        @update:page="load"
        @update:page-size="load"
      />
    </div>
  </n-card>

  <n-modal v-model:show="showCreate" preset="card" title="新建用户" style="width: 420px">
    <n-form label-placement="top">
      <n-form-item label="用户名" required>
        <n-input v-model:value="createForm.username" />
      </n-form-item>
      <n-form-item label="密码" required>
        <n-input v-model:value="createForm.password" type="password" show-password-on="click" />
      </n-form-item>
      <n-form-item label="角色" required>
        <n-select v-model:value="createForm.role" :options="roleOpts" />
      </n-form-item>
      <n-form-item label="姓名">
        <n-input v-model:value="createForm.real_name" />
      </n-form-item>
      <n-button type="primary" block :loading="creating" @click="createUser">创建</n-button>
    </n-form>
  </n-modal>

  <n-modal v-model:show="showEdit" preset="card" :title="`编辑用户：${editForm.username}`" style="width: 420px">
    <n-form label-placement="top">
      <n-form-item label="用户名">
        <n-input :value="editForm.username" disabled />
      </n-form-item>
      <n-form-item label="姓名">
        <n-input v-model:value="editForm.real_name" placeholder="可留空" />
      </n-form-item>
      <n-form-item label="角色" required>
        <n-select v-model:value="editForm.role" :options="roleOpts" />
      </n-form-item>
      <n-form-item label="状态" required>
        <n-select
          v-model:value="editForm.status"
          :options="[
            { label: '正常', value: 1 },
            { label: '禁用', value: 0 },
          ]"
        />
      </n-form-item>
      <n-form-item label="重置密码（可选）">
        <n-input
          v-model:value="editForm.password"
          type="password"
          show-password-on="click"
          placeholder="不填写则不修改密码"
        />
      </n-form-item>
      <n-button type="primary" block :loading="updating" @click="updateUser">保存修改</n-button>
    </n-form>
  </n-modal>
</template>
