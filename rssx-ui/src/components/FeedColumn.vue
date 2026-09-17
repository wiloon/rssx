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
    <p class="feed-column__heading">Feeds</p>
    <ul class="feed-column__list">
      <li
        v-for="feed in feeds"
        :key="feed.id"
        data-test="feed"
        class="feed-column__item"
        :class="{ 'is-selected': feed.id === selectedFeedId }"
        @click="$emit('select', feed.id)"
      >
        <span class="feed-column__dot" aria-hidden="true" />
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
.feed-column {
  padding: 8px 0 16px;
}
.feed-column__heading {
  margin: 4px 16px 8px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.45);
}
.feed-column__list {
  list-style: none;
  margin: 0;
  padding: 0 8px;
}
.feed-column__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  margin-bottom: 2px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.12s;
}
.feed-column__item:hover {
  background: rgba(0, 0, 0, 0.05);
}
.feed-column__item.is-selected {
  background: rgba(25, 118, 210, 0.12);
  color: #1976d2;
  font-weight: 600;
}
.feed-column__dot {
  width: 6px;
  height: 6px;
  min-width: 6px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.15);
}
.feed-column__item.is-selected .feed-column__dot {
  background: #1976d2;
}
.feed-column__title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
}
.feed-column__badge {
  min-width: 20px;
  padding: 1px 6px;
  border-radius: 999px;
  background: rgba(25, 118, 210, 0.14);
  color: #1976d2;
  font-size: 11px;
  font-weight: 600;
  text-align: center;
}
.feed-column__item.is-selected .feed-column__badge {
  background: #1976d2;
  color: #fff;
}
.feed-column__sync {
  border: none;
  background: none;
  padding: 0 2px;
  font-size: 15px;
  line-height: 1;
  color: rgba(0, 0, 0, 0.4);
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.1s;
}
.feed-column__item:hover .feed-column__sync,
.feed-column__sync.is-syncing {
  opacity: 1;
}
.feed-column__sync:hover {
  color: #1976d2;
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
