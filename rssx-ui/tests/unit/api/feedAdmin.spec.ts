import { createFeedAdminApi } from '@/api/feedAdmin'
import type { HttpClient } from '@/api/reader'

type Method = 'get' | 'post' | 'put' | 'delete'

interface Call {
  method: Method
  url: string
  data?: unknown
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
    get: (url) => respond({ method: 'get', url }),
    post: (url, data) => respond({ method: 'post', url, data }),
    put: (url, data) => respond({ method: 'put', url, data }),
    delete: (url) => respond({ method: 'delete', url })
  }
}

describe('feed-admin API client', () => {
  it('lists feeds, mapping the backend PascalCase to the editable record', async () => {
    const api = createFeedAdminApi(
      fakeHttp({
        '/feeds/detail': [
          { Id: 1, Title: 'Hacker News', Url: 'https://hnrss.org/newest' },
          { Id: 2, Title: 'r/golang', Url: 'https://www.reddit.com/r/golang/.rss' }
        ]
      })
    )

    expect(await api.list()).toEqual([
      { id: 1, title: 'Hacker News', url: 'https://hnrss.org/newest' },
      { id: 2, title: 'r/golang', url: 'https://www.reddit.com/r/golang/.rss' }
    ])
  })

  it('creates a feed with a { url, title } body and maps the response back', async () => {
    const http = fakeHttp({
      '/feed': { Id: 7, Title: 'New', Url: 'https://example.com/rss' }
    })
    const api = createFeedAdminApi(http)

    const created = await api.create('New', 'https://example.com/rss')

    expect(created).toEqual({ id: 7, title: 'New', url: 'https://example.com/rss' })
    expect(http.calls[0]).toEqual({
      method: 'post',
      url: '/feed',
      data: { url: 'https://example.com/rss', title: 'New' }
    })
  })

  it('updates a feed via PUT /feed/:id', async () => {
    const http = fakeHttp({
      '/feed/5': { Id: 5, Title: 'Renamed', Url: 'https://example.com/rss' }
    })
    const api = createFeedAdminApi(http)

    const updated = await api.update(5, 'Renamed', 'https://example.com/rss')

    expect(updated).toEqual({ id: 5, title: 'Renamed', url: 'https://example.com/rss' })
    expect(http.calls[0]).toEqual({
      method: 'put',
      url: '/feed/5',
      data: { url: 'https://example.com/rss', title: 'Renamed' }
    })
  })

  it('purges a feed via DELETE /feed/:id/purge', async () => {
    const http = fakeHttp({ '/feed/9/purge': null })
    const api = createFeedAdminApi(http)

    await api.remove(9)

    expect(http.calls[0]).toEqual({ method: 'delete', url: '/feed/9/purge' })
  })

  it('triggers a one-off sync via POST /sync/:id', async () => {
    const http = fakeHttp({ '/sync/3': null })
    const api = createFeedAdminApi(http)

    await api.sync(3)

    expect(http.calls[0]).toEqual({ method: 'post', url: '/sync/3', data: undefined })
  })
})
