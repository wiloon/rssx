import { selectionToQuery, queryToSelection } from '@/reader/urlSync'

describe('reader URL sync — a reload or a shared link restores the same view', () => {
  it('serialises the current selection to query params, omitting what is not selected', () => {
    expect(selectionToQuery({ feedId: 3, articleId: '100' })).toEqual({
      feed: '3',
      article: '100'
    })
    expect(selectionToQuery({ feedId: 3, articleId: null })).toEqual({ feed: '3' })
    expect(selectionToQuery({ feedId: null, articleId: null })).toEqual({})
  })

  it('restores a selection from query params', () => {
    expect(queryToSelection({ feed: '3', article: '100' })).toEqual({
      feedId: 3,
      articleId: '100'
    })
    // The "All" view is feed -1.
    expect(queryToSelection({ feed: '-1' })).toEqual({
      feedId: -1,
      articleId: null
    })
    expect(queryToSelection({})).toEqual({ feedId: null, articleId: null })
  })

  it('drops a malformed feed id, and an article with no feed', () => {
    expect(queryToSelection({ feed: 'nope', article: '100' })).toEqual({
      feedId: null,
      articleId: null
    })
    expect(queryToSelection({ article: '100' })).toEqual({
      feedId: null,
      articleId: null
    })
  })
})
