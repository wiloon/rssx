/**
 * The article list is the middle column of the reader: the ordered set of
 * Articles shown for the selected Feed, plus which one is open in the reading
 * pane. It is fed by the backend's unread-window endpoints (see
 * docs/adr/0001-three-pane-reader-layout.md) so it behaves as an unread queue,
 * not a full archive.
 *
 * Every function here is pure: (state, input) => new state. No DOM, no HTTP.
 */

export interface Article {
  id: string
  feedId: number
  title: string
  url: string
  content: string
  pubDate: string
  read: boolean
}

export interface ArticleList {
  articles: Article[]
  selectedId: string | null
}

export function createArticleList (): ArticleList {
  return { articles: [], selectedId: null }
}

/**
 * Replace the list with a freshly loaded unread window (from GET /news-list),
 * keeping the order the feed returned. Clears the current selection.
 */
export function loadWindow (_list: ArticleList, window: Article[]): ArticleList {
  return { articles: [...window], selectedId: null }
}

/**
 * Load one older article (from GET /previous-news) onto the tail of the list.
 * The list is newest-first, so "older" means the end. A no-op if the article is
 * already present. The selection is untouched.
 */
export function loadOlder (list: ArticleList, older: Article): ArticleList {
  if (list.articles.some((a) => a.id === older.id)) {
    return list
  }
  return { ...list, articles: [...list.articles, older] }
}

/**
 * Open an article in the reading pane. Opening an article marks it read
 * immediately (the backend does the same server-side on GET /news), so the
 * middle column greys it out without waiting for a refetch.
 */
export function selectArticle (list: ArticleList, id: string): ArticleList {
  return {
    articles: list.articles.map((a) =>
      a.id === id ? { ...a, read: true } : a
    ),
    selectedId: id
  }
}

/** Close the reading pane, keeping the loaded list and its read state. */
export function clearSelection (list: ArticleList): ArticleList {
  return { ...list, selectedId: null }
}
