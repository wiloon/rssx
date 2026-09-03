<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useReaderStore } from '@/stores/reader'
import { visiblePanes } from '@/reader/layout'
import { queryToSelection, selectionToQuery } from '@/reader/urlSync'
import FeedColumn from '@/components/FeedColumn.vue'
import ArticleColumn from '@/components/ArticleColumn.vue'
import ReadingPane from '@/components/ReadingPane.vue'

const reader = useReaderStore()
const route = useRoute()
const router = useRouter()

// Narrow-viewport detection for the pane collapse (docs/adr/0001).
const narrowQuery =
  typeof window.matchMedia === 'function'
    ? window.matchMedia('(max-width: 900px)')
    : null
const isNarrow = ref(narrowQuery?.matches ?? false)
const onNarrowChange = (e: MediaQueryListEvent) => {
  isNarrow.value = e.matches
}
narrowQuery?.addEventListener('change', onNarrowChange)
onUnmounted(() => narrowQuery?.removeEventListener('change', onNarrowChange))

// The left-column per-feed refresh button. The backend sync is asynchronous, so
// hold the spinner for a beat even if the trigger call returns immediately.
const syncingFeedId = ref<number | null>(null)
async function onSyncFeed (feedId: number): Promise<void> {
  if (syncingFeedId.value !== null) return
  syncingFeedId.value = feedId
  try {
    await Promise.all([
      reader.syncFeed(feedId),
      new Promise((resolve) => setTimeout(resolve, 2000))
    ])
  } finally {
    syncingFeedId.value = null
  }
}

const selection = computed(() => ({
  feedId: reader.state.selectedFeedId,
  articleId: reader.state.list.selectedId
}))
const panes = computed(() => visiblePanes(selection.value, isNarrow.value))

onMounted(async () => {
  await reader.loadFeeds()
  const restored = queryToSelection(route.query as Record<string, string>)
  if (restored.feedId !== null) {
    await reader.openFeed(restored.feedId)
    if (restored.articleId !== null) await reader.openArticle(restored.articleId)
  }
})

// Keep the URL in step with the selection so reload / sharing works.
watch(selection, (value) => {
  router.replace({ query: selectionToQuery(value) })
})
</script>

<template>
  <div class="reader">
    <FeedColumn
      v-show="panes.includes('feeds')"
      class="reader__pane reader__pane--feeds"
      :feeds="reader.state.feeds"
      :selected-feed-id="reader.state.selectedFeedId"
      :syncing-feed-id="syncingFeedId"
      @select="reader.openFeed"
      @sync="onSyncFeed"
    />
    <ArticleColumn
      v-show="panes.includes('articles')"
      class="reader__pane reader__pane--articles"
      :articles="reader.state.list.articles"
      :selected-article-id="reader.state.list.selectedId"
      :feed-selected="reader.state.selectedFeedId !== null"
      @select="reader.openArticle"
      @mark-page-read="reader.markPageRead"
      @load-more="reader.loadMore"
      @back="reader.deselectFeed"
    />
    <ReadingPane
      v-show="panes.includes('reading')"
      class="reader__pane reader__pane--reading"
      :open="reader.state.open"
      @next="reader.openNext"
      @back="reader.deselectArticle"
    />
  </div>
</template>

<style scoped>
.reader {
  display: flex;
  height: calc(100vh - 48px);
}
.reader__pane {
  overflow-y: auto;
  height: 100%;
}
.reader__pane--feeds {
  flex: 0 0 240px;
  border-right: 1px solid rgba(0, 0, 0, 0.12);
}
.reader__pane--articles {
  flex: 0 0 360px;
  border-right: 1px solid rgba(0, 0, 0, 0.12);
}
.reader__pane--reading {
  flex: 1;
}

@media (max-width: 900px) {
  .reader__pane--feeds,
  .reader__pane--articles {
    flex: 1;
    border-right: none;
  }
}
</style>
