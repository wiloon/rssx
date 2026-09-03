import { vi } from 'vitest'
import { createReaderStore } from '@/reader/readerStore'
import { ReaderApi } from '@/api/reader'

const article = (id: string, over: Record<string, unknown> = {}) => ({
  id,
  feedId: 3,
  title: `Article ${id}`,
  url: `https://example.com/${id}`,
  content: '',
  pubDate: '',
  read: false,
  ...over
})

function fakeApi (over: Partial<ReaderApi> = {}): ReaderApi {
  return {
    listFeeds: vi.fn().mockResolvedValue([]),
    listUnread: vi.fn().mockResolvedValue([]),
    getArticle: vi.fn(),
    markPageRead: vi.fn().mockResolvedValue([]),
    previousArticle: vi.fn(),
    syncAll: vi.fn().mockResolvedValue(undefined),
    syncOne: vi.fn().mockResolvedValue(undefined),
    ...over
  } as ReaderApi
}

describe('reader store — orchestrates the three panes', () => {
  it('opens a feed: loads its unread window into the middle column and records the selection', async () => {
    const api = fakeApi({
      listUnread: vi.fn().mockResolvedValue([article('a'), article('b')])
    })
    const store = createReaderStore(api)

    await store.openFeed(3)

    expect(api.listUnread).toHaveBeenCalledWith(3)
    expect(store.state.selectedFeedId).toBe(3)
    expect(store.state.list.articles.map((a) => a.id)).toEqual(['a', 'b'])
  })

  it('opens an article: fetches it for the reading pane and greys it in the middle column', async () => {
    const api = fakeApi({
      listUnread: vi.fn().mockResolvedValue([article('a'), article('b')]),
      getArticle: vi.fn().mockResolvedValue({
        article: article('b', { content: '<p>hi</p>', read: true }),
        nextId: 'c'
      })
    })
    const store = createReaderStore(api)
    await store.openFeed(3)

    await store.openArticle('b')

    expect(api.getArticle).toHaveBeenCalledWith(3, 'b')
    expect(store.state.open!.article.content).toBe('<p>hi</p>')
    expect(store.state.open!.nextId).toBe('c')
    expect(store.state.list.articles.find((a) => a.id === 'b')!.read).toBe(true)
    expect(store.state.list.selectedId).toBe('b')
  })

  it('loads more: asks for the article before the oldest one shown and appends it', async () => {
    const api = fakeApi({
      listUnread: vi.fn().mockResolvedValue([article('c'), article('b')]),
      previousArticle: vi.fn()
        .mockResolvedValue({ article: article('a'), nextId: 'b' })
    })
    const store = createReaderStore(api)
    await store.openFeed(3)

    await store.loadMore()

    expect(api.previousArticle).toHaveBeenCalledWith(3, 'b')
    expect(store.state.list.articles.map((a) => a.id)).toEqual(['c', 'b', 'a'])
  })

  it('opens the next article using the nextId carried by the reading pane', async () => {
    const api = fakeApi({
      listUnread: vi.fn().mockResolvedValue([article('a'), article('b')]),
      getArticle: vi.fn()
        .mockResolvedValueOnce({ article: article('a'), nextId: 'b' })
        .mockResolvedValueOnce({ article: article('b'), nextId: 'c' })
    })
    const store = createReaderStore(api)
    await store.openFeed(3)
    await store.openArticle('a')

    await store.openNext()

    expect(api.getArticle).toHaveBeenLastCalledWith(3, 'b')
    expect(store.state.open!.nextId).toBe('c')
  })

  it('loads the feed list for the left column', async () => {
    const api = fakeApi({
      listFeeds: vi.fn().mockResolvedValue([
        { id: -1, title: 'All', unread: 0 },
        { id: 3, title: 'InfoQ', unread: 12 }
      ])
    })
    const store = createReaderStore(api)

    await store.loadFeeds()

    expect(store.state.feeds.map((f) => f.id)).toEqual([-1, 3])
  })

  it('marks the current page read and shows the next unread window', async () => {
    const api = fakeApi({
      listUnread: vi.fn().mockResolvedValue([article('a'), article('b')]),
      markPageRead: vi.fn()
        .mockResolvedValue([article('x'), article('y')])
    })
    const store = createReaderStore(api)
    await store.openFeed(3)

    await store.markPageRead()

    expect(api.markPageRead).toHaveBeenCalledWith(3)
    expect(store.state.list.articles.map((a) => a.id)).toEqual(['x', 'y'])
  })

  it('subscribes to a feed and refreshes the left column', async () => {
    const api = fakeApi({
      listFeeds: vi
        .fn()
        .mockResolvedValueOnce([{ id: 3, title: 'InfoQ', unread: 0 }])
        .mockResolvedValueOnce([
          { id: 3, title: 'InfoQ', unread: 0 },
          { id: 9, title: 'Lobsters', unread: 0 }
        ]),
      addFeed: vi
        .fn()
        .mockResolvedValue({ id: 9, title: 'Lobsters', unread: 0 })
    })
    const store = createReaderStore(api)
    await store.loadFeeds()

    await store.addFeed('https://lobste.rs/rss', 'Lobsters')

    expect(api.addFeed).toHaveBeenCalledWith('https://lobste.rs/rss', 'Lobsters')
    expect(store.state.feeds.map((f) => f.id)).toEqual([3, 9])
  })

  it('syncs one feed, then refreshes the left column and the open window', async () => {
    const api = fakeApi({
      listFeeds: vi
        .fn()
        .mockResolvedValueOnce([{ id: 3, title: 'InfoQ', unread: 0 }])
        .mockResolvedValueOnce([{ id: 3, title: 'InfoQ', unread: 4 }]),
      listUnread: vi
        .fn()
        .mockResolvedValueOnce([])
        .mockResolvedValueOnce([article('a'), article('b')]),
      syncOne: vi.fn().mockResolvedValue(undefined)
    })
    const store = createReaderStore(api)
    await store.loadFeeds()
    await store.openFeed(3)

    await store.syncFeed(3)

    expect(api.syncOne).toHaveBeenCalledWith(3)
    expect(store.state.feeds).toEqual([{ id: 3, title: 'InfoQ', unread: 4 }])
    expect(store.state.list.articles.map((a) => a.id)).toEqual(['a', 'b'])
  })

  it('syncing a feed that is not open refreshes the left column only', async () => {
    const api = fakeApi({
      listFeeds: vi.fn().mockResolvedValue([{ id: 9, title: 'Lobsters', unread: 1 }]),
      syncOne: vi.fn().mockResolvedValue(undefined)
    })
    const store = createReaderStore(api)

    await store.syncFeed(9)

    expect(api.syncOne).toHaveBeenCalledWith(9)
    expect(api.listUnread).not.toHaveBeenCalled()
  })

  it('unsubscribes from a feed, refreshes the left column, and clears the panes if it was open', async () => {
    const api = fakeApi({
      listFeeds: vi
        .fn()
        .mockResolvedValueOnce([
          { id: 3, title: 'InfoQ', unread: 0 },
          { id: 9, title: 'Lobsters', unread: 0 }
        ])
        .mockResolvedValueOnce([{ id: 3, title: 'InfoQ', unread: 0 }]),
      listUnread: vi.fn().mockResolvedValue([article('a')]),
      removeFeed: vi.fn().mockResolvedValue(undefined)
    })
    const store = createReaderStore(api)
    await store.loadFeeds()
    await store.openFeed(9)

    await store.removeFeed(9)

    expect(api.removeFeed).toHaveBeenCalledWith(9)
    expect(store.state.feeds.map((f) => f.id)).toEqual([3])
    expect(store.state.selectedFeedId).toBeNull()
    expect(store.state.list.articles).toEqual([])
  })

  it('closes the reading pane but keeps the article list (narrow-view back from an article)', async () => {
    const api = fakeApi({
      listUnread: vi.fn().mockResolvedValue([article('a')]),
      getArticle: vi
        .fn()
        .mockResolvedValue({ article: article('a'), nextId: '' })
    })
    const store = createReaderStore(api)
    await store.openFeed(3)
    await store.openArticle('a')

    store.deselectArticle()

    expect(store.state.open).toBeNull()
    expect(store.state.list.selectedId).toBeNull()
    expect(store.state.list.articles).toHaveLength(1)
  })

  it('goes back to the feed column (narrow-view back from an article list)', async () => {
    const api = fakeApi({
      listUnread: vi.fn().mockResolvedValue([article('a')])
    })
    const store = createReaderStore(api)
    await store.openFeed(3)

    store.deselectFeed()

    expect(store.state.selectedFeedId).toBeNull()
    expect(store.state.open).toBeNull()
    expect(store.state.list.articles).toEqual([])
  })
})
