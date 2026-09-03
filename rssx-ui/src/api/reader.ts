/**
 * The reader API client wraps the backend's unread-window endpoints
 * (see docs/adr/0001-three-pane-reader-layout.md) behind a small typed surface
 * the store and views can call. It also absorbs the backend's quirks — notably
 * that GET /feeds concatenates each feed's unread count into its title string.
 *
 * The HTTP layer is injected so the client can be tested without a network.
 */

import { Article } from '@/reader/articleList'

export interface HttpClient {
  get (url: string, config?: unknown): Promise<{ data: any }>
  post (url: string, data?: unknown, config?: unknown): Promise<{ data: any }>
  delete (url: string, config?: unknown): Promise<{ data: any }>
}

export interface Feed {
  id: number
  title: string
  unread: number
}

interface RawFeed {
  Id: number
  Title: string
  Url: string
}

/** The backend's article JSON (PascalCase, from the `news` package). */
interface RawArticle {
  Id: string
  FeedId: number
  Title: string
  Url: string
  Description: string
  PubDate: string
  ReadFlag: boolean
  NextId?: string
}

/** One article shown in the reading pane, plus where "next" goes. */
export interface OpenArticle {
  article: Article
  nextId: string
}

function toArticle (raw: RawArticle): Article {
  return {
    id: raw.Id,
    feedId: raw.FeedId,
    title: raw.Title,
    url: raw.Url,
    content: raw.Description,
    pubDate: raw.PubDate,
    read: raw.ReadFlag
  }
}

function toOpenArticle (raw: RawArticle): OpenArticle {
  return { article: toArticle(raw), nextId: raw.NextId ?? '' }
}

/**
 * The backend appends " - <n>" to each feed title. Split it back off the tail,
 * but only when the tail really is a number — feed titles can contain " - "
 * themselves.
 */
function parseFeedTitle (raw: string): { title: string; unread: number } {
  const match = raw.match(/^(.*) - (\d+)$/)
  if (!match) {
    return { title: raw, unread: 0 }
  }
  return { title: match[1], unread: Number(match[2]) }
}

function toFeed (raw: RawFeed): Feed {
  const { title, unread } = parseFeedTitle(raw.Title)
  return { id: raw.Id, title, unread }
}

export function createReaderApi (http: HttpClient) {
  return {
    async listFeeds (): Promise<Feed[]> {
      const { data } = await http.get('/feeds')
      return (data as RawFeed[]).map(toFeed)
    },

    async listUnread (feedId: number): Promise<Article[]> {
      const { data } = await http.get('/news-list', { params: { id: feedId } })
      return (data as RawArticle[]).map(toArticle)
    },

    async getArticle (feedId: number, articleId: string): Promise<OpenArticle> {
      const { data } = await http.get('/news', {
        params: { id: articleId, feedId }
      })
      return toOpenArticle(data as RawArticle)
    },

    async markPageRead (feedId: number): Promise<Article[]> {
      const { data } = await http.get('/mark-read', { params: { feedId } })
      return (data as RawArticle[]).map(toArticle)
    },

    async previousArticle (
      feedId: number,
      articleId: string
    ): Promise<OpenArticle> {
      const { data } = await http.get('/previous-news', {
        params: { newsId: articleId, feedId }
      })
      return toOpenArticle(data as RawArticle)
    },

    async syncAll (): Promise<void> {
      await http.post('/sync')
    },

    async syncOne (feedId: number): Promise<void> {
      await http.post(`/sync/${feedId}`)
    },

    async addFeed (url: string, title: string): Promise<Feed> {
      const { data } = await http.post('/feed', { url, title })
      return toFeed(data as RawFeed)
    },

    async removeFeed (feedId: number): Promise<void> {
      await http.delete(`/feed/${feedId}`)
    }
  }
}

export type ReaderApi = ReturnType<typeof createReaderApi>
