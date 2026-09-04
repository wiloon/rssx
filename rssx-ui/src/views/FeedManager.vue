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
    notify(errorText(err, 'Failed to load feeds'))
  } finally {
    loading.value = false
  }
}

// --- Add ---------------------------------------------------------------------
const draft = reactive({ title: '', url: '' })
const adding = ref(false)

async function add (): Promise<void> {
  if (draft.title.trim() === '' || draft.url.trim() === '') {
    notify('Title and URL are both required')
    return
  }
  adding.value = true
  try {
    await api.create(draft.title.trim(), draft.url.trim())
    draft.title = ''
    draft.url = ''
    notify('Feed added')
    await refresh()
  } catch (err) {
    notify(errorText(err, 'Failed to add feed'))
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
    notify('Title and URL are both required')
    return
  }
  saving.value = true
  try {
    await api.update(editing.id, editing.title.trim(), editing.url.trim())
    editDialog.value = false
    notify('Changes saved')
    await refresh()
  } catch (err) {
    notify(errorText(err, 'Failed to save changes'))
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
    notify('Feed deleted')
    await refresh()
  } catch (err) {
    notify(errorText(err, 'Failed to delete feed'))
  } finally {
    deleting.value = false
  }
}

// --- Sync ---------------------------------------------------------------
async function syncNow (feed: AdminFeed): Promise<void> {
  try {
    await api.sync(feed.id)
    notify(`Sync triggered: ${feed.title}`)
  } catch (err) {
    notify(errorText(err, 'Failed to sync feed'))
  }
}

onMounted(refresh)
</script>

<template>
  <v-container class="feed-manager" style="max-width: 900px">
    <div class="feed-manager__header">
      <h2>Feed Manager</h2>
      <v-btn variant="text" data-test="back-to-reader" @click="router.push({ name: 'Reader' })">
        Back to reader
      </v-btn>
    </div>

    <v-sheet class="pa-4 mb-4" border rounded>
      <div class="feed-manager__add">
        <v-text-field
          v-model="draft.title"
          data-test="add-title"
          label="Title"
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
          Add
        </v-btn>
      </div>
    </v-sheet>

    <v-table data-test="feed-table">
      <thead>
        <tr>
          <th>Title</th>
          <th>URL</th>
          <th class="text-right">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="!loading && feeds.length === 0">
          <td colspan="3" class="text-medium-emphasis">No feeds yet</td>
        </tr>
        <tr v-for="feed in feeds" :key="feed.id" data-test="feed-row">
          <td data-test="feed-title">{{ feed.title }}</td>
          <td class="feed-manager__url">
            <a :href="feed.url" target="_blank" rel="noopener">{{ feed.url }}</a>
          </td>
          <td class="text-right">
            <v-btn size="small" variant="text" data-test="sync" @click="syncNow(feed)">
              Sync
            </v-btn>
            <v-btn size="small" variant="text" data-test="edit" @click="openEdit(feed)">
              Edit
            </v-btn>
            <v-btn
              size="small"
              variant="text"
              color="error"
              data-test="delete"
              @click="openDelete(feed)"
            >
              Delete
            </v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>

    <v-dialog v-model="editDialog" max-width="500">
      <v-card>
        <v-card-title>Edit feed</v-card-title>
        <v-card-text>
          <v-text-field v-model="editing.title" data-test="edit-title" label="Title" />
          <v-text-field v-model="editing.url" data-test="edit-url" label="RSS URL" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="editDialog = false">Cancel</v-btn>
          <v-btn
            color="primary"
            data-test="edit-save"
            :loading="saving"
            @click="saveEdit"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="deleteDialog" max-width="460">
      <v-card>
        <v-card-title>Permanently delete feed</v-card-title>
        <v-card-text>
          This will delete "{{ target?.title }}" along with all of its fetched articles
          and read history. This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn
            color="error"
            data-test="delete-confirm"
            :loading="deleting"
            @click="confirmDelete"
          >
            Delete
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
