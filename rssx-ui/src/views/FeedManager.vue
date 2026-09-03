<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { http } from '@/api/http'
import { createFeedAdminApi, type AdminFeed } from '@/api/feedAdmin'

const router = useRouter()
const api = createFeedAdminApi(http)

const feeds = ref<AdminFeed[]>([])
const loading = ref(false)

const snackbar = reactive({ show: false, text: '' })
function notify (text: string): void {
  snackbar.text = text
  snackbar.show = true
}

/** Pull the backend's `{ error }` message out of an axios failure. */
function errorText (err: unknown, fallback: string): string {
  const res = (err as { response?: { data?: { error?: string } } }).response
  return res?.data?.error ?? fallback
}

async function refresh (): Promise<void> {
  loading.value = true
  try {
    feeds.value = await api.list()
  } catch (err) {
    notify(errorText(err, '加载订阅源失败'))
  } finally {
    loading.value = false
  }
}

// --- Add ---------------------------------------------------------------------
const draft = reactive({ title: '', url: '' })
const adding = ref(false)

async function add (): Promise<void> {
  if (draft.title.trim() === '' || draft.url.trim() === '') {
    notify('标题和 URL 都不能为空')
    return
  }
  adding.value = true
  try {
    await api.create(draft.title.trim(), draft.url.trim())
    draft.title = ''
    draft.url = ''
    notify('已添加')
    await refresh()
  } catch (err) {
    notify(errorText(err, '添加失败'))
  } finally {
    adding.value = false
  }
}

// --- Edit ------------------------------------------------------------------
const editDialog = ref(false)
const editing = reactive({ id: 0, title: '', url: '' })
const saving = ref(false)

function openEdit (feed: AdminFeed): void {
  editing.id = feed.id
  editing.title = feed.title
  editing.url = feed.url
  editDialog.value = true
}

async function saveEdit (): Promise<void> {
  if (editing.title.trim() === '' || editing.url.trim() === '') {
    notify('标题和 URL 都不能为空')
    return
  }
  saving.value = true
  try {
    await api.update(editing.id, editing.title.trim(), editing.url.trim())
    editDialog.value = false
    notify('已保存')
    await refresh()
  } catch (err) {
    notify(errorText(err, '保存失败'))
  } finally {
    saving.value = false
  }
}

// --- Delete ---------------------------------------------------------------
const deleteDialog = ref(false)
const deleting = ref(false)
const target = ref<AdminFeed | null>(null)

function openDelete (feed: AdminFeed): void {
  target.value = feed
  deleteDialog.value = true
}

async function confirmDelete (): Promise<void> {
  if (target.value === null) return
  deleting.value = true
  try {
    await api.remove(target.value.id)
    deleteDialog.value = false
    notify('已删除')
    await refresh()
  } catch (err) {
    notify(errorText(err, '删除失败'))
  } finally {
    deleting.value = false
  }
}

// --- Sync ---------------------------------------------------------------
async function syncNow (feed: AdminFeed): Promise<void> {
  try {
    await api.sync(feed.id)
    notify(`已触发同步：${feed.title}`)
  } catch (err) {
    notify(errorText(err, '同步失败'))
  }
}

onMounted(refresh)
</script>

<template>
  <v-container class="feed-manager" style="max-width: 900px">
    <div class="feed-manager__header">
      <h2>订阅源管理</h2>
      <v-btn variant="text" data-test="back-to-reader" @click="router.push({ name: 'Reader' })">
        返回阅读
      </v-btn>
    </div>

    <v-sheet class="pa-4 mb-4" border rounded>
      <div class="feed-manager__add">
        <v-text-field
          v-model="draft.title"
          data-test="add-title"
          label="标题"
          density="compact"
          hide-details
        />
        <v-text-field
          v-model="draft.url"
          data-test="add-url"
          label="RSS URL"
          density="compact"
          hide-details
        />
        <v-btn
          color="primary"
          data-test="add-submit"
          :loading="adding"
          @click="add"
        >
          添加
        </v-btn>
      </div>
    </v-sheet>

    <v-table data-test="feed-table">
      <thead>
        <tr>
          <th>标题</th>
          <th>URL</th>
          <th class="text-right">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="!loading && feeds.length === 0">
          <td colspan="3" class="text-medium-emphasis">还没有订阅源</td>
        </tr>
        <tr v-for="feed in feeds" :key="feed.id" data-test="feed-row">
          <td data-test="feed-title">{{ feed.title }}</td>
          <td class="feed-manager__url">
            <a :href="feed.url" target="_blank" rel="noopener">{{ feed.url }}</a>
          </td>
          <td class="text-right">
            <v-btn size="small" variant="text" data-test="sync" @click="syncNow(feed)">
              同步
            </v-btn>
            <v-btn size="small" variant="text" data-test="edit" @click="openEdit(feed)">
              编辑
            </v-btn>
            <v-btn
              size="small"
              variant="text"
              color="error"
              data-test="delete"
              @click="openDelete(feed)"
            >
              删除
            </v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>

    <v-dialog v-model="editDialog" max-width="500">
      <v-card>
        <v-card-title>编辑订阅源</v-card-title>
        <v-card-text>
          <v-text-field v-model="editing.title" data-test="edit-title" label="标题" />
          <v-text-field v-model="editing.url" data-test="edit-url" label="RSS URL" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="editDialog = false">取消</v-btn>
          <v-btn
            color="primary"
            data-test="edit-save"
            :loading="saving"
            @click="saveEdit"
          >
            保存
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="deleteDialog" max-width="460">
      <v-card>
        <v-card-title>彻底删除订阅源</v-card-title>
        <v-card-text>
          将删除「{{ target?.title }}」及其全部已抓取文章和已读记录，此操作不可恢复。
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">取消</v-btn>
          <v-btn
            color="error"
            data-test="delete-confirm"
            :loading="deleting"
            @click="confirmDelete"
          >
            删除
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-snackbar v-model="snackbar.show" :timeout="3000">{{ snackbar.text }}</v-snackbar>
  </v-container>
</template>

<style scoped>
.feed-manager__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.feed-manager__add {
  display: flex;
  gap: 12px;
  align-items: center;
}
.feed-manager__url {
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
