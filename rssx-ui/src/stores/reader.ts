/**
 * Pinia wrapper around the framework-agnostic reader store factory. The factory
 * (src/reader/readerStore.ts) holds all the orchestration logic and is unit
 * tested on its own; this file only makes its state reactive and exposes it to
 * components.
 */

import { defineStore } from 'pinia'
import { reactive } from 'vue'
import { createReaderApi } from '@/api/reader'
import { http } from '@/api/http'
import { createInitialState, createReaderStore } from '@/reader/readerStore'

export const useReaderStore = defineStore('reader', () => {
  const store = createReaderStore(
    createReaderApi(http),
    reactive(createInitialState())
  )

  return {
    state: store.state,
    loadFeeds: store.loadFeeds,
    openFeed: store.openFeed,
    openArticle: store.openArticle,
    openNext: store.openNext,
    loadMore: store.loadMore,
    markPageRead: store.markPageRead,
    deselectArticle: store.deselectArticle,
    deselectFeed: store.deselectFeed,
    addFeed: store.addFeed,
    removeFeed: store.removeFeed,
    syncFeed: store.syncFeed
  }
})
