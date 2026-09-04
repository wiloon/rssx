<script setup lang="ts">
import type { Feed } from '@/api/reader'

defineProps<{
  feeds: Feed[]
  selectedFeedId: number | null
  syncingFeedId?: number | null
}>()

defineEmits<{
  (e: 'select', feedId: number): void
  (e: 'sync', feedId: number): void
}>()
</script>

<template>
  <nav class="feed-column">
    <ul class="feed-column__list">
      <li
        v-for="feed in feeds"
        :key="feed.id"
        data-test="feed"
        class="feed-column__item"
        :class="{ 'is-selected': feed.id === selectedFeedId }"
        @click="$emit('select', feed.id)"
      >
        <span data-test="feed-title" class="feed-column__title">
          {{ feed.title }}
        </span>
        <button
          v-if="feed.id !== -1"
          type="button"
          data-test="feed-sync"
          class="feed-column__sync"
          :class="{ 'is-syncing': feed.id === syncingFeedId }"
          :disabled="feed.id === syncingFeedId"
          title="Sync this feed"
          @click.stop="$emit('sync', feed.id)"
        >
          ↻
        </button>
        <span
          v-if="feed.unread > 0"
          data-test="feed-badge"
          class="feed-column__badge"
        >
          {{ feed.unread }}
        </span>
      </li>
    </ul>
  </nav>
</template>

<style scoped>
.feed-column__list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.feed-column__item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  cursor: pointer;
}
.feed-column__item:hover {
  background: rgba(0, 0, 0, 0.04);
}
.feed-column__item.is-selected {
  background: rgba(25, 118, 210, 0.12);
}
.feed-column__title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.feed-column__badge {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.6);
}
.feed-column__sync {
  border: none;
  background: none;
  padding: 0 4px;
  font-size: 14px;
  line-height: 1;
  color: rgba(0, 0, 0, 0.45);
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.1s;
}
.feed-column__item:hover .feed-column__sync,
.feed-column__sync.is-syncing {
  opacity: 1;
}
.feed-column__sync:hover {
  color: rgba(25, 118, 210, 0.9);
}
.feed-column__sync.is-syncing {
  animation: feed-column-spin 0.8s linear infinite;
  cursor: default;
}
@keyframes feed-column-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
