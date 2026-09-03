import { createReaderApi, HttpClient } from '@/api/reader'

type Method = 'get' | 'post' | 'put' | 'delete'

interface Call {
  method: Method
  url: string
  data?: unknown
  config?: unknown
}

interface RecordingHttp extends HttpClient {
  calls: Call[]
}

/** A fake HTTP client that answers from a routing table and records its calls. */
function fakeHttp (routes: Record<string, unknown>): RecordingHttp {
  const calls: Call[] = []
  const respond = (call: Call) => {
    calls.push(call)
    if (!(call.url in routes)) {
      return Promise.reject(
        new Error(`no stub for ${call.method.toUpperCase()} ${call.url}`)
      )
    }
    return Promise.resolve({ data: routes[call.url] })
  }
  return {
    calls,
    get: (url, config) => respond({ method: 'get', url, config }),
    post: (url, data, config) => respond({ method: 'post', url, data, config }),
    put: (url, data, config) => respond({ method: 'put', url, data, config }),
    delete: (url, config) => respond({ method: 'delete', url, config })
  }
}

describe('reader API client — the seam between the UI and the backend', () => {
  it('parses the feed list, splitting off the unread count the backend appends to each title', async () => {
    const api = createReaderApi(
      fakeHttp({
        '/feeds': [
          { Id: -1, Title: 'All', Url: '' },
          { Id: 3, Title: 'InfoQ - 12', Url: 'https://infoq.com/rss' },
          { Id: 4, Title: 'InfoQ China - 0', Url: 'https://infoq.cn/rss' }
        ]
      })
    )

    const feeds = await api.listFeeds()

    expect(feeds).toEqual([
      { id: -1, title: 'All', unread: 0 },
      { id: 3, title: 'InfoQ', unread: 12 },
      { id: 4, title: 'InfoQ China', unread: 0 }
    ])
  })

  it('only strips the trailing count, keeping " - " that is part of the feed title itself', async () => {
    const api = createReaderApi(
      fakeHttp({
        '/feeds': [
          { Id: 7, Title: 'InfoQ - Software Architecture - 5', Url: 'x' }
        ]
      })
    )

    const [feed] = await api.listFeeds()

    expect(feed).toEqual({
      id: 7,
      title: 'InfoQ - Software Architecture',
      unread: 5
    })
  })

  it('loads the unread window for a feed and maps the backend article shape to the reader shape', async () => {
    const http = fakeHttp({
      '/news-list': [
        {
          Id: '100',
          FeedId: 3,
          Title: 'Something happened',
          Url: 'https://infoq.com/a/100',
          Description: '<p>body</p>',
          PubDate: '2026-08-30',
          ReadFlag: false
        },
        {
          Id: '99',
          FeedId: 3,
          Title: 'Older thing',
          Url: 'https://infoq.com/a/99',
          Description: '<p>older</p>',
          PubDate: '2026-08-29',
          ReadFlag: true
        }
      ]
    })
    const api = createReaderApi(http)

    const articles = await api.listUnread(3)

    expect(http.calls).toContainEqual({
      method: 'get',
      url: '/news-list',
      config: { params: { id: 3 } }
    })
    expect(articles).toEqual([
      {
        id: '100',
        feedId: 3,
        title: 'Something happened',
        url: 'https://infoq.com/a/100',
        content: '<p>body</p>',
        pubDate: '2026-08-30',
        read: false
      },
      {
        id: '99',
        feedId: 3,
        title: 'Older thing',
        url: 'https://infoq.com/a/99',
        content: '<p>older</p>',
        pubDate: '2026-08-29',
        read: true
      }
    ])
  })

  it('loads one article for the reading pane along with the id of the next one', async () => {
    const http = fakeHttp({
      '/news': {
        Id: '100',
        FeedId: 3,
        Title: 'T',
        Url: 'u',
        Description: 'd',
        PubDate: 'p',
        ReadFlag: true,
        NextId: '101'
      }
    })
    const api = createReaderApi(http)

    const result = await api.getArticle(3, '100')

    expect(http.calls).toContainEqual({
      method: 'get',
      url: '/news',
      config: { params: { id: '100', feedId: 3 } }
    })
    expect(result).toEqual({
      article: {
        id: '100',
        feedId: 3,
        title: 'T',
        url: 'u',
        content: 'd',
        pubDate: 'p',
        read: true
      },
      nextId: '101'
    })
  })

  it('marks the current page read and returns the next unread window', async () => {
    const http = fakeHttp({
      '/mark-read': [
        {
          Id: '80',
          FeedId: 3,
          Title: 'Next page item',
          Url: 'u',
          Description: 'd',
          PubDate: 'p',
          ReadFlag: false
        }
      ]
    })
    const api = createReaderApi(http)

    const next = await api.markPageRead(3)

    expect(http.calls).toContainEqual({
      method: 'get',
      url: '/mark-read',
      config: { params: { feedId: 3 } }
    })
    expect(next.map((a) => a.id)).toEqual(['80'])
  })

  it('loads the article just before a given one, for the reading pane back button', async () => {
    const http = fakeHttp({
      '/previous-news': {
        Id: '98',
        FeedId: 3,
        Title: 'Previous',
        Url: 'u',
        Description: 'd',
        PubDate: 'pd',
        ReadFlag: true,
        NextId: '99'
      }
    })
    const api = createReaderApi(http)

    const result = await api.previousArticle(3, '99')

    expect(http.calls).toContainEqual({
      method: 'get',
      url: '/previous-news',
      config: { params: { newsId: '99', feedId: 3 } }
    })
    expect(result).toEqual({
      article: {
        id: '98',
        feedId: 3,
        title: 'Previous',
        url: 'u',
        content: 'd',
        pubDate: 'pd',
        read: true
      },
      nextId: '99'
    })
  })

  it('triggers a sync of all feeds', async () => {
    const http = fakeHttp({ '/sync': { message: 'sync started' } })
    const api = createReaderApi(http)

    await api.syncAll()

    expect(http.calls).toContainEqual({ method: 'post', url: '/sync', data: undefined, config: undefined })
  })

  it('triggers a sync of a single feed by id', async () => {
    const http = fakeHttp({ '/sync/3': { message: 'sync started', feed_id: 3 } })
    const api = createReaderApi(http)

    await api.syncOne(3)

    expect(http.calls).toContainEqual({ method: 'post', url: '/sync/3', data: undefined, config: undefined })
  })

  it('subscribes to a new feed by url and title', async () => {
    const http = fakeHttp({
      '/feed': { Id: 9, Title: 'Lobsters', Url: 'https://lobste.rs/rss' }
    })
    const api = createReaderApi(http)

    const feed = await api.addFeed('https://lobste.rs/rss', 'Lobsters')

    expect(http.calls).toContainEqual({
      method: 'post',
      url: '/feed',
      data: { url: 'https://lobste.rs/rss', title: 'Lobsters' },
      config: undefined
    })
    expect(feed).toEqual({ id: 9, title: 'Lobsters', unread: 0 })
  })

  it('unsubscribes from a feed by id', async () => {
    const http = fakeHttp({ '/feed/9': '' })
    const api = createReaderApi(http)

    await api.removeFeed(9)

    expect(http.calls).toContainEqual({
      method: 'delete',
      url: '/feed/9',
      config: undefined
    })
  })
})
