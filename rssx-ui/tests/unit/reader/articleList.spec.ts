import {
  createArticleList,
  loadWindow,
  selectArticle,
  loadOlder,
  clearSelection
} from '@/reader/articleList'
import { Article, ArticleList } from '@/reader/articleList'

const article = (id: string, over: Partial<Article> = {}): Article => ({
  id,
  feedId: 1,
  title: `Article ${id}`,
  url: `https://example.com/${id}`,
  content: '',
  pubDate: '',
  read: false,
  ...over
})

describe('article list — the middle column of the reader', () => {
  it('shows the articles from the unread window in the order the feed returned them', () => {
    const list: ArticleList = loadWindow(createArticleList(), [
      article('a'),
      article('b'),
      article('c')
    ])

    expect(list.articles.map((a) => a.id)).toEqual(['a', 'b', 'c'])
    expect(list.selectedId).toBeNull()
  })

  it('marks an article as read when it is opened in the reading pane', () => {
    let list = loadWindow(createArticleList(), [article('a'), article('b')])

    list = selectArticle(list, 'b')

    expect(list.selectedId).toBe('b')
    expect(list.articles.find((a) => a.id === 'b')!.read).toBe(true)
    expect(list.articles.find((a) => a.id === 'a')!.read).toBe(false)
  })

  it('appends older articles at the tail when loading more, keeps the selection, and ignores duplicates', () => {
    let list = loadWindow(createArticleList(), [article('c'), article('b')])
    list = selectArticle(list, 'c')

    list = loadOlder(list, article('a'))
    expect(list.articles.map((a) => a.id)).toEqual(['c', 'b', 'a'])
    expect(list.selectedId).toBe('c')

    list = loadOlder(list, article('a'))
    expect(list.articles.map((a) => a.id)).toEqual(['c', 'b', 'a'])
  })

  it('clears the selection (narrow-view back) without dropping the loaded articles or their read state', () => {
    let list = loadWindow(createArticleList(), [article('a'), article('b')])
    list = selectArticle(list, 'a')

    list = clearSelection(list)

    expect(list.selectedId).toBeNull()
    expect(list.articles.map((a) => a.id)).toEqual(['a', 'b'])
    expect(list.articles.find((a) => a.id === 'a')!.read).toBe(true)
  })
})
