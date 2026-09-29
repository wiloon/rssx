import DOMPurify from 'dompurify'

// Feed links should never replace the reader, nor hand it to the target page.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName === 'A' && node.hasAttribute('href')) {
    node.setAttribute('target', '_blank')
    node.setAttribute('rel', 'noopener noreferrer')
  }
})

/**
 * Strips scripts, event handlers, `javascript:` URLs and other active content
 * from feed-provided HTML so it can be rendered with `v-html`.
 */
export function sanitizeFeedHtml (html: string): string {
  return DOMPurify.sanitize(html, {
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'textarea', 'select'],
    ADD_ATTR: ['target']
  })
}
