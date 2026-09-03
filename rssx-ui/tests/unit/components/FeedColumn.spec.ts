import { mount } from '@vue/test-utils'
import FeedColumn from '@/components/FeedColumn.vue'
import type { Feed } from '@/api/reader'

const feeds: Feed[] = [
  { id: -1, title: 'All', unread: 0 },
  { id: 3, title: 'InfoQ', unread: 12 },
  { id: 4, title: 'InfoQ China', unread: 0 }
]

describe('FeedColumn — the left pane', () => {
  it('lists each feed title, with an unread badge only when there are unread items', () => {
    const wrapper = mount(FeedColumn, { props: { feeds, selectedFeedId: null } })

    const items = wrapper.findAll('[data-test="feed"]')
    expect(
      items.map((i) => i.get('[data-test="feed-title"]').text())
    ).toEqual(['All', 'InfoQ', 'InfoQ China'])
    expect(items.map((i) => i.find('[data-test="feed-badge"]').exists())).toEqual(
      [false, true, false]
    )
    expect(items[1].get('[data-test="feed-badge"]').text()).toBe('12')
  })

  it('emits select with the feed id when a feed is clicked', async () => {
    const wrapper = mount(FeedColumn, { props: { feeds, selectedFeedId: null } })

    await wrapper.findAll('[data-test="feed"]')[1].trigger('click')

    expect(wrapper.emitted('select')).toEqual([[3]])
  })

  it('marks the selected feed', () => {
    const wrapper = mount(FeedColumn, { props: { feeds, selectedFeedId: 3 } })

    const selected = wrapper.get('[data-test="feed"].is-selected')
    expect(selected.text()).toContain('InfoQ')
  })
})
