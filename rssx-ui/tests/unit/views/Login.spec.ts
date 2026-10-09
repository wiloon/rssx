import { flushPromises, mount } from '@vue/test-utils'
import Login from '@/views/Login.vue'
import { axiosInstance } from '@/api/http'

vi.mock('@/api/http', () => ({
  axiosInstance: { post: vi.fn() }
}))

const push = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push })
}))

describe('Login', () => {
  beforeEach(() => {
    push.mockReset()
    vi.mocked(axiosInstance.post).mockReset()
    vi.mocked(axiosInstance.post).mockResolvedValue({
      data: { code: 20000, data: { token: 'session-token' } }
    })
  })

  function mountForm () {
    return mount(Login, {
      global: {
        stubs: {
          'v-container': { template: '<div><slot /></div>' },
          'v-form': {
            emits: ['submit'],
            template: '<form @submit.prevent="$emit(\'submit\', $event)"><slot /></form>'
          },
          'v-text-field': {
            props: ['modelValue', 'type'],
            emits: ['update:modelValue'],
            template:
              '<input :type="type || \'text\'" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
          },
          'v-btn': {
            props: ['type'],
            template: '<button :type="type"><slot /></button>'
          },
          'v-snackbar': true
        }
      }
    })
  }

  async function fillForm () {
    const wrapper = mountForm()
    await wrapper.get('[data-cy="user-name"]').setValue('alice')
    await wrapper.get('[data-cy="password"]').setValue('secret')
    return wrapper
  }

  it('signs in when the form is submitted with Enter', async () => {
    const wrapper = await fillForm()

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(axiosInstance.post).toHaveBeenCalledOnce()
    expect(axiosInstance.post).toHaveBeenCalledWith('/login', {
      name: 'alice',
      password: 'secret'
    })
    expect(push).toHaveBeenCalledWith({ name: 'Reader' })
  })

  it('uses a submit button so a click posts the form once', async () => {
    const wrapper = await fillForm()
    const button = wrapper.get('[data-cy="login"]')
    expect(button.attributes('type')).toBe('submit')

    button.element.closest('form')?.requestSubmit()
    await flushPromises()

    expect(axiosInstance.post).toHaveBeenCalledOnce()
  })
})
