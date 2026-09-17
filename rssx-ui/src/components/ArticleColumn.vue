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
      <span class="article-column__heading">Articles</span>
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
        <span class="article-column__dot" aria-hidden="true" />
        <span class="article-column__title">{{ a.title }}</span>
      </li>
    </ul>

    <footer v-if="feedSelected" class="article-column__footer">
      <button
        data-test="mark-page-read"
        class="article-column__action"
        @click="$emit('mark-page-read')"
      >
        Mark page as read
      </button>
      <button
        data-test="load-more"
        class="article-column__action article-column__action--primary"
        @click="$emit('load-more')"
      >
        Load more
      </button>
    </footer>
  </section>
</template>

<style scoped>
.article-column {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.article-column__header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 8px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}
.article-column__back {
  border: none;
  background: none;
  font-size: 20px;
  line-height: 1;
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  color: rgba(0, 0, 0, 0.6);
}
.article-column__back:hover {
  background: rgba(0, 0, 0, 0.06);
}
.article-column__heading {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.45);
}
.article-column__list {
  list-style: none;
  margin: 0;
  padding: 4px;
  flex: 1;
  overflow-y: auto;
}
.article-column__item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  margin-bottom: 2px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.12s;
}
.article-column__item:hover {
  background: rgba(0, 0, 0, 0.04);
}
.article-column__dot {
  margin-top: 6px;
  width: 6px;
  height: 6px;
  min-width: 6px;
  border-radius: 50%;
  background: #1976d2;
}
.article-column__title {
  font-size: 14px;
  line-height: 1.4;
  font-weight: 600;
}
.article-column__item.is-read .article-column__dot {
  background: transparent;
}
.article-column__item.is-read .article-column__title {
  font-weight: 400;
  color: rgba(0, 0, 0, 0.5);
}
.article-column__item.is-selected {
  background: rgba(25, 118, 210, 0.12);
}
.article-column__footer {
  display: flex;
  gap: 8px;
  padding: 10px 8px;
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}
.article-column__action {
  flex: 1;
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid rgba(0, 0, 0, 0.15);
  background: #fff;
  font-size: 13px;
  cursor: pointer;
  transition: background-color 0.12s;
}
.article-column__action:hover {
  background: rgba(0, 0, 0, 0.04);
}
.article-column__action--primary {
  border-color: #1976d2;
  color: #1976d2;
  font-weight: 600;
}
.article-column__action--primary:hover {
  background: rgba(25, 118, 210, 0.08);
}
</style>
