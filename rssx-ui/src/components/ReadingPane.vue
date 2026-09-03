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
          <span> &middot; </span>
          <a
            data-test="original"
            :href="open.article.url"
            target="_blank"
            rel="noopener"
          >
            打开原文
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
            :disabled="open.nextId === ''"
            @click="$emit('next')"
          >
            下一篇
          </button>
        </div>
      </div>
    </template>

    <p v-else data-test="empty" class="reading-pane__empty">选择一篇文章</p>
  </article>
</template>

<style scoped>
.reading-pane {
  padding: 0 24px 24px;
}
.reading-pane__title {
  font-size: 1.5rem;
  margin: 16px 0 8px;
}
.reading-pane__meta {
  color: rgba(0, 0, 0, 0.6);
  margin-bottom: 16px;
}
.reading-pane__footer {
  margin-top: 24px;
}
.reading-pane__empty {
  color: rgba(0, 0, 0, 0.4);
  margin-top: 24px;
}
</style>
