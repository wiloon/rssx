import { visiblePanes } from '@/reader/layout'

describe('reader layout — which panes are on screen', () => {
  it('shows all three panes on a wide viewport regardless of selection', () => {
    expect(visiblePanes({ feedId: 3, articleId: '1' }, false)).toEqual([
      'feeds',
      'articles',
      'reading'
    ])
    expect(visiblePanes({ feedId: null, articleId: null }, false)).toEqual([
      'feeds',
      'articles',
      'reading'
    ])
  })

  it('on a narrow viewport shows one pane, as deep as the selection goes', () => {
    expect(visiblePanes({ feedId: null, articleId: null }, true)).toEqual([
      'feeds'
    ])
    expect(visiblePanes({ feedId: 3, articleId: null }, true)).toEqual([
      'articles'
    ])
    expect(visiblePanes({ feedId: 3, articleId: '1' }, true)).toEqual(['reading'])
  })
})
