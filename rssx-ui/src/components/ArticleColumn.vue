<script setup lang="ts">
import type { Article } from '@/reader/articleList'

defineProps<{
  articles: Article[]
  selectedArticleId: string | null
  feedSelected: boolean
}>()

defineEmits<{
  (e: 'select', articleId: string): void
  (e: 'mark-page-read'): void
  (e: 'load-more'): void
  (e: 'back'): void
}>()
</script>

<template>
  <section class="article-column">
    <header class="article-column__header">
      <button data-test="back" class="article-column__back" @click="$emit('back')">
        ‹
      </button>
    </header>

    <ul class="article-column__list">
      <li
        v-for="a in articles"
        :key="a.id"
        data-test="article"
        class="article-column__item"
        :class="{
          'is-read': a.read,
          'is-selected': a.id === selectedArticleId
        }"
        @click="$emit('select', a.id)"
      >
        {{ a.title }}
      </li>
    </ul>

    <footer v-if="feedSelected" class="article-column__footer">
      <button data-test="mark-page-read" @click="$emit('mark-page-read')">
        Mark page as read
      </button>
      <button data-test="load-more" @click="$emit('load-more')">Load more</button>
    </footer>
  </section>
</template>

<style scoped>
.article-column__list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.article-column__item {
  padding: 8px 16px;
  cursor: pointer;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}
.article-column__item.is-read {
  color: rgba(0, 0, 0, 0.5);
}
.article-column__item.is-selected {
  background: rgba(25, 118, 210, 0.12);
}
.article-column__footer {
  padding: 8px;
  border-top: 1px solid rgba(0, 0, 0, 0.12);
}
</style>
