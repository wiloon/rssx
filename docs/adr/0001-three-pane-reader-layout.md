# Three-pane reader layout

## Status

accepted

## Context

The UI navigates as three full-page routes drilled through one at a time:
`FeedList` (`/`) → `FeedNewsList` (`/feed-news-list?feedId=`) → `News`
(`/news?newsid=&feedid=`). Going from one Article back to the Feed list is a
browser back-navigation, and the Feed list and Article list are never visible at
the same time. This is unlike every mainstream desktop RSS reader (NetNewsWire,
Reeder, Feedly, Inoreader, Miniflux), which show three persistent columns:
Feeds | Articles | Reading pane.

## Decision

Replace the route-per-view drill-down with a single persistent three-pane layout:

- **Left — Feed column.** The "All view" plus a flat list of subscribed Feeds,
  each with its unread count. Selecting a Feed drives the middle column.
- **Middle — Article column.** The selected Feed's Articles, newest first, unread
  in normal weight and read greyed. Selecting an Article drives the right pane.
- **Right — Reading pane.** The selected Article's title, date, source, and
  Feed-provided content rendered as sanitised HTML, with "open original" and
  previous/next controls.

Selection state (`selectedFeedId`, `selectedArticleId`) lives in the Vuex store
and in the URL query so a reload and a shared link both restore the same view.
The three route components collapse into one `Reader` view; `/login` and
`/register` stay separate routes.

On narrow viewports the three panes degrade to one visible pane at a time with
back navigation between them — the current drill-down, but as in-layout state
rather than router history.

### The Article column keeps the existing unread-window API

The middle column is fed by the existing endpoints unchanged:
`GET /news-list` (a page of unread Articles from the read boundary),
`GET /mark-read` (advance the boundary a page, return the next page),
`GET /previous-news` (step back one Article), `GET /news` (one Article, marks it
read as a side effect).

Consequence: the middle column is an **unread queue**, not a full archive. You
page forward through unread Articles and can step back over ones you have read,
but you cannot freely scroll a Feed's entire history. A full chronological
timeline with random access would need a new paginated list endpoint and a
weaker role for the read boundary; that is deferred, not rejected — see the
"considered options".

`GET /feeds` currently returns each unread count concatenated into the Feed's
title string (`"InfoQ - 12"`). The layout needs the count as its own field; this
is a small backend change assumed alongside this ADR, not a redesign.

## Considered options

- **Keep the drill-down, just restyle it.** Cheapest, but leaves the core
  interaction — no side-by-side context, back-button to change Feeds — that
  motivated the change.
- **Three-pane with a new full-timeline list endpoint now.** The ideal end
  state, but it reworks the read-state model (boundary + out-of-order read set)
  at the same time as the layout. Splitting them keeps this change reviewable;
  the timeline can be its own ADR once the layout is in place.

## Consequences

- Deep links change shape: one `/reader?feed=&article=` instead of three routes.
  Old bookmarks break; acceptable for a single-user self-hosted tool.
- E2E tests that assert on route transitions between the three pages need
  rewriting against in-layout selection.
- The "All view" (feed id `-1`) still has no backend list implementation
  (`LoadNewsListByFeed` returns nothing for `-1`); the left column shows it but
  it stays empty until that is filled in.
