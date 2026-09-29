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

  it('strips scripts, event handlers and javascript: URLs from the content', () => {
    const wrapper = mount(ReadingPane, {
      props: {
        open: open({
          content:
            '<p>ok</p><script>alert(1)</script>' +
            '<img src="x.png" onerror="alert(2)">' +
            '<a href="javascript:alert(3)">bad</a>' +
            '<iframe src="https://evil.example"></iframe>'
        })
      }
    })

    const body = wrapper.get('[data-test="body"]')
    expect(body.html()).toContain('<p>ok</p>')
    expect(body.find('script').exists()).toBe(false)
    expect(body.find('iframe').exists()).toBe(false)
    expect(body.get('img').attributes('onerror')).toBeUndefined()
    expect(body.get('a').attributes('href')).toBeUndefined()
  })

  it('opens content links in a new tab without an opener', () => {
    const wrapper = mount(ReadingPane, {
      props: { open: open({ content: '<a href="https://example.com/x">x</a>' }) }
    })

    const link = wrapper.get('[data-test="body"] a')
    expect(link.attributes('href')).toBe('https://example.com/x')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toBe('noopener noreferrer')
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
