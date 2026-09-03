import { mount } from '@vue/test-utils'
import ArticleColumn from '@/components/ArticleColumn.vue'
import type { Article } from '@/reader/articleList'

const article = (id: string, over: Partial<Article> = {}): Article => ({
  id,
  feedId: 3,
  title: `Article ${id}`,
  url: `https://example.com/${id}`,
  content: '',
  pubDate: '',
  read: false,
  ...over
})

const articles = [article('a', { read: true }), article('b')]

describe('ArticleColumn — the middle pane', () => {
  it('lists articles and greys out the ones already read', () => {
    const wrapper = mount(ArticleColumn, {
      props: { articles, selectedArticleId: null, feedSelected: true }
    })

    const items = wrapper.findAll('[data-test="article"]')
    expect(items).toHaveLength(2)
    expect(items[0].classes()).toContain('is-read')
    expect(items[1].classes()).not.toContain('is-read')
  })

  it('emits select with the article id when one is clicked', async () => {
    const wrapper = mount(ArticleColumn, {
      props: { articles, selectedArticleId: null, feedSelected: true }
    })

    await wrapper.findAll('[data-test="article"]')[1].trigger('click')

    expect(wrapper.emitted('select')).toEqual([['b']])
  })

  it('shows the page controls only when a feed is selected, and they emit', async () => {
    const withoutFeed = mount(ArticleColumn, {
      props: { articles: [], selectedArticleId: null, feedSelected: false }
    })
    expect(withoutFeed.find('[data-test="mark-page-read"]').exists()).toBe(false)

    const wrapper = mount(ArticleColumn, {
      props: { articles, selectedArticleId: null, feedSelected: true }
    })
    await wrapper.get('[data-test="mark-page-read"]').trigger('click')
    await wrapper.get('[data-test="load-more"]').trigger('click')

    expect(wrapper.emitted('mark-page-read')).toHaveLength(1)
    expect(wrapper.emitted('load-more')).toHaveLength(1)
  })
})
