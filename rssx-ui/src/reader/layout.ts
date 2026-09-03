/**
 * Which panes are on screen. On a wide viewport all three show side by side; on
 * a narrow one only the deepest-selected pane shows, and "back" is clearing the
 * selection a level (docs/adr/0001).
 */

import type { Selection } from '@/reader/urlSync'

export type Pane = 'feeds' | 'articles' | 'reading'

const ALL: Pane[] = ['feeds', 'articles', 'reading']

export function visiblePanes (selection: Selection, isNarrow: boolean): Pane[] {
  if (!isNarrow) return ALL
  if (selection.articleId !== null) return ['reading']
  if (selection.feedId !== null) return ['articles']
  return ['feeds']
}
