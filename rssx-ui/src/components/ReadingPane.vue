<script setup lang="ts">
import type { OpenArticle } from '@/api/reader'

defineProps<{
  open: OpenArticle | null
}>()

defineEmits<{
  (e: 'next'): void
  (e: 'back'): void
}>()
</script>

<template>
  <article class="reading-pane">
    <header class="reading-pane__header">
      <button data-test="back" class="reading-pane__back" @click="$emit('back')">
        ‹
      </button>
    </header>

    <template v-if="open">
      <div data-test="article" class="reading-pane__article">
        <h1 data-test="title" class="reading-pane__title">
          {{ open.article.title }}
        </h1>
        <p class="reading-pane__meta">
          <span>{{ open.article.pubDate }}</span>
          <span class="reading-pane__meta-sep">&middot;</span>
          <a
            data-test="original"
            class="reading-pane__link"
            :href="open.article.url"
            target="_blank"
            rel="noopener"
          >
            Open original
          </a>
        </p>
        <!-- Feed-provided HTML; sanitising is a follow-up (docs/adr/0001). -->
        <!-- eslint-disable-next-line vue/no-v-html -->
        <div
          data-test="body"
          class="reading-pane__body"
          v-html="open.article.content"
        />
        <div class="reading-pane__footer">
          <button
            data-test="next"
            class="reading-pane__next"
            :disabled="open.nextId === ''"
            @click="$emit('next')"
          >
            Next article →
          </button>
        </div>
      </div>
    </template>

    <div v-else class="reading-pane__empty-wrap">
      <p data-test="empty" class="reading-pane__empty">Select an article to read</p>
    </div>
  </article>
</template>

<style scoped>
.reading-pane {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.reading-pane__header {
  padding: 6px 8px;
}
.reading-pane__back {
  border: none;
  background: none;
  font-size: 20px;
  line-height: 1;
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  color: rgba(0, 0, 0, 0.6);
}
.reading-pane__back:hover {
  background: rgba(0, 0, 0, 0.06);
}
.reading-pane__article {
  max-width: 760px;
  margin: 0 auto;
  padding: 0 32px 48px;
  width: 100%;
  box-sizing: border-box;
}
.reading-pane__title {
  font-size: 1.75rem;
  font-weight: 700;
  line-height: 1.3;
  margin: 8px 0 12px;
}
.reading-pane__meta {
  color: rgba(0, 0, 0, 0.55);
  font-size: 13px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}
.reading-pane__meta-sep {
  margin: 0 6px;
}
.reading-pane__link {
  color: #1976d2;
  text-decoration: none;
}
.reading-pane__link:hover {
  text-decoration: underline;
}
.reading-pane__body {
  font-size: 15px;
  line-height: 1.7;
  color: rgba(0, 0, 0, 0.85);
}
.reading-pane__body :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 6px;
}
.reading-pane__footer {
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}
.reading-pane__next {
  padding: 10px 18px;
  border-radius: 6px;
  border: none;
  background: #1976d2;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.reading-pane__next:hover:not(:disabled) {
  background: #1565c0;
}
.reading-pane__next:disabled {
  background: rgba(0, 0, 0, 0.12);
  color: rgba(0, 0, 0, 0.38);
  cursor: default;
}
.reading-pane__empty-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.reading-pane__empty {
  color: rgba(0, 0, 0, 0.4);
  font-size: 15px;
}
</style>
