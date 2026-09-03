/**
 * The feed-admin API client backs the feed maintenance page (src/views/
 * FeedManager.vue). It is deliberately separate from the reader client
 * (src/api/reader.ts): that one deals in the unread-window reading flow and its
 * quirky feed titles, this one deals in the plain editable feed record.
 *
 * The HTTP layer is injected so the client can be tested without a network.
 */

import type { HttpClient } from '@/api/reader'

/** One feed as shown on the maintenance page — the editable fields only. */
export interface AdminFeed {
  id: number
  title: string
  url: string
}

/** The backend's feed JSON (PascalCase, from the `feed` Go package). */
interface RawFeed {
  Id: number
  Title: string
  Url: string
}

function toAdminFeed (raw: RawFeed): AdminFeed {
  return { id: raw.Id, title: raw.Title, url: raw.Url }
}

export function createFeedAdminApi (http: HttpClient) {
  return {
    /** GET /feeds/detail — the subscribed feeds, without unread decoration. */
    async list (): Promise<AdminFeed[]> {
      const { data } = await http.get('/feeds/detail')
      return (data as RawFeed[]).map(toAdminFeed)
    },

    /** POST /feed — create the feed and subscribe to it. */
    async create (title: string, url: string): Promise<AdminFeed> {
      const { data } = await http.post('/feed', { url, title })
      return toAdminFeed(data as RawFeed)
    },

    /** PUT /feed/:id — rename or repoint an existing feed. */
    async update (id: number, title: string, url: string): Promise<AdminFeed> {
      const { data } = await http.put(`/feed/${id}`, { url, title })
      return toAdminFeed(data as RawFeed)
    },

    /** DELETE /feed/:id/purge — delete the feed and all its articles. */
    async remove (id: number): Promise<void> {
      await http.delete(`/feed/${id}/purge`)
    },

    /** POST /sync/:id — trigger a one-off sync of this feed. */
    async sync (id: number): Promise<void> {
      await http.post(`/sync/${id}`)
    }
  }
}

export type FeedAdminApi = ReturnType<typeof createFeedAdminApi>
