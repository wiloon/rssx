/**
 * The reader store is the seam that orchestrates the three panes: it holds the
 * current selection, calls the API client, and drives the pure article-list
 * model. Views read its state and call its actions; they never talk to the API
 * directly.
 *
 * It is a plain factory so it can be unit-tested with a fake API and no
 * framework. `createReaderStore` takes its state object so the Pinia wrapper
 * (src/stores/reader.ts) can hand it a `reactive()` one and have mutations
 * drive rendering; tests let it default to a fresh plain object.
 */

import { Feed, OpenArticle, ReaderApi } from '@/api/reader'
import {
  ArticleList,
  clearSelection,
  createArticleList,
  loadOlder,
  loadWindow,
  selectArticle
} from '@/reader/articleList'

export interface ReaderState {
  feeds: Feed[]
  selectedFeedId: number | null
  list: ArticleList
  open: OpenArticle | null
}

export function createInitialState (): ReaderState {
  return {
    feeds: [],
    selectedFeedId: null,
    list: createArticleList(),
    open: null
  }
}

export function createReaderStore (
  api: ReaderApi,
  state: ReaderState = createInitialState()
) {
  /** Load the subscribed feeds shown in the left column. */
  async function loadFeeds (): Promise<void> {
    state.feeds = await api.listFeeds()
  }

  /** Open an article in the reading pane; it greys out in the middle column. */
  async function openArticle (articleId: string): Promise<void> {
    if (state.selectedFeedId === null) return
    state.open = await api.getArticle(state.selectedFeedId, articleId)
    state.list = selectArticle(state.list, articleId)
  }

  return {
    state,
    openArticle,
    loadFeeds,

    /** Select a feed and load its unread window into the middle column. */
    async openFeed (feedId: number): Promise<void> {
      state.selectedFeedId = feedId
      state.open = null
      state.list = loadWindow(state.list, await api.listUnread(feedId))
    },

    /** Pull in the article just before the oldest one currently shown. */
    async loadMore (): Promise<void> {
      const { articles } = state.list
      const oldest = articles[articles.length - 1]
      if (state.selectedFeedId === null || oldest === undefined) return
      const previous = await api.previousArticle(state.selectedFeedId, oldest.id)
      state.list = loadOlder(state.list, previous.article)
    },

    /** Open the article after the one in the reading pane, if there is one. */
    async openNext (): Promise<void> {
      if (state.open === null || state.open.nextId === '') return
      await openArticle(state.open.nextId)
    },

    /** Advance the read boundary a page and show the next unread window. */
    async markPageRead (): Promise<void> {
      if (state.selectedFeedId === null) return
      state.open = null
      state.list = loadWindow(
        state.list,
        await api.markPageRead(state.selectedFeedId)
      )
    },

    /** Close the reading pane (narrow-view back from an article). */
    deselectArticle (): void {
      state.open = null
      state.list = clearSelection(state.list)
    },

    /** Return to the feed column (narrow-view back from an article list). */
    deselectFeed (): void {
      state.selectedFeedId = null
      state.open = null
      state.list = createArticleList()
    },

    /**
     * Trigger a one-off sync of one feed, then refresh the left column (and the
     * open window, if it is the feed being synced). The backend sync runs
     * asynchronously, so fresh articles may only show up on a later refresh.
     */
    async syncFeed (feedId: number): Promise<void> {
      await api.syncOne(feedId)
      await loadFeeds()
      if (state.selectedFeedId === feedId) {
        state.list = loadWindow(state.list, await api.listUnread(feedId))
      }
    },

    /** Subscribe to a new feed, then refresh the left column. */
    async addFeed (url: string, title: string): Promise<void> {
      await api.addFeed(url, title)
      await loadFeeds()
    },

    /** Unsubscribe from a feed; clear the panes if it was the open one. */
    async removeFeed (feedId: number): Promise<void> {
      await api.removeFeed(feedId)
      if (state.selectedFeedId === feedId) {
        state.selectedFeedId = null
        state.list = createArticleList()
        state.open = null
      }
      await loadFeeds()
    }
  }
}

export type ReaderStore = ReturnType<typeof createReaderStore>
