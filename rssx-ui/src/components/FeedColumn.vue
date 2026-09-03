<script setup lang="ts">
import type { Feed } from '@/api/reader'

defineProps<{
  feeds: Feed[]
  selectedFeedId: number | null
}>()

defineEmits<{
  (e: 'select', feedId: number): void
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
  padding: 8px 16px;
  cursor: pointer;
}
.feed-column__item:hover {
  background: rgba(0, 0, 0, 0.04);
}
.feed-column__item.is-selected {
  background: rgba(25, 118, 210, 0.12);
}
.feed-column__badge {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.6);
}
</style>
