/**
 * Maps the reader's selection to and from URL query params so a reload or a
 * shared link restores the same view (docs/adr/0001). Pure functions; the
 * router wiring lives in the Reader view.
 */

export interface Selection {
  feedId: number | null
  articleId: string | null
}

type Query = Record<string, string | undefined>

export function selectionToQuery (selection: Selection): Record<string, string> {
  const query: Record<string, string> = {}
  if (selection.feedId !== null) {
    query.feed = String(selection.feedId)
    // An article only makes sense in the context of a feed.
    if (selection.articleId !== null) {
      query.article = selection.articleId
    }
  }
  return query
}

export function queryToSelection (query: Query): Selection {
  const empty: Selection = { feedId: null, articleId: null }

  if (query.feed === undefined) return empty
  const feedId = Number(query.feed)
  if (!Number.isInteger(feedId)) return empty

  return {
    feedId,
    articleId: query.article ?? null
  }
}
