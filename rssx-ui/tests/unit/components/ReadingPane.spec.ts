import { mount } from '@vue/test-utils'
import ReadingPane from '@/components/ReadingPane.vue'
import type { OpenArticle } from '@/api/reader'

const open = (over: Partial<OpenArticle['article']> = {}, nextId = 'n1'): OpenArticle => ({
  article: {
    id: '100',
    feedId: 3,
    title: 'The headline',
    url: 'https://example.com/100',
    content: '<p>Body <strong>text</strong></p>',
    pubDate: '2026-08-30',
    read: true,
    ...over
  },
  nextId
})

describe('ReadingPane — the right pane', () => {
  it('shows a placeholder when nothing is open', () => {
    const wrapper = mount(ReadingPane, { props: { open: null } })

    expect(wrapper.find('[data-test="article"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="empty"]').exists()).toBe(true)
  })

  it('renders the open article: title, feed-provided HTML content, and a link to the original', () => {
    const wrapper = mount(ReadingPane, { props: { open: open() } })

    expect(wrapper.get('[data-test="title"]').text()).toBe('The headline')
    expect(wrapper.get('[data-test="body"]').html()).toContain(
      '<strong>text</strong>'
    )
    expect(wrapper.get('[data-test="original"]').attributes('href')).toBe(
      'https://example.com/100'
    )
  })

  it('emits next, unless there is no next article', async () => {
    const wrapper = mount(ReadingPane, { props: { open: open({}, 'n1') } })
    await wrapper.get('[data-test="next"]').trigger('click')
    expect(wrapper.emitted('next')).toHaveLength(1)

    const last = mount(ReadingPane, { props: { open: open({}, '') } })
    expect(
      last.get('[data-test="next"]').attributes('disabled')
    ).toBeDefined()
  })
})
