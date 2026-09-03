# RSSX

RSSX is a self-hosted RSS reader: a Go API (`rssx-api`) that syncs RSS sources into
storage and tracks read state, and a Vue frontend (`rssx-ui`) that presents them.

## Language

**Feed**:
An RSS/Atom source that RSSX polls, e.g. "InfoQ" or "InfoQ China". Identified by a
stable numeric id. A Feed exists independently of who subscribes to it.
_Avoid_: channel, source, site

**Subscription**:
The link between a user and a Feed. Adding a Feed to your list creates a
Subscription; removing it deletes the Subscription, not the Feed.
_Avoid_: follow, watch

**Purge**:
Deleting a Feed outright — its row, every Subscription to it, and all of its
Articles and read state in Redis. Distinct from removing a Subscription, which
leaves the Feed intact. Exposed as `DELETE /feed/:id/purge` and the delete
action on the feed management page.
_Avoid_: (using "delete" without saying which)

**Article**:
One item fetched from a Feed: title, publish date, source URL, and the content the
Feed provided (often only a summary). Identified by a per-Feed id.
_Avoid_: news, item, entry, post, story. (The `news` Go package predates this
glossary; new names should use "Article".)

**Sync**:
Fetching a Feed's current items from its source URL and storing any not seen before.
Runs on a timer and can be triggered manually per-Feed or for all Feeds.
_Avoid_: refresh, poll, crawl

**Read boundary**:
Per user, per Feed: the point in a Feed's Article history up to which everything is
considered read. Articles newer than the boundary are unread unless individually
marked. Moving the boundary forward is how "mark all as read" works.
_Avoid_: cursor, offset, watermark

**All view**:
A pseudo-Feed (id `-1`) in the left column that aggregates Articles across every
Subscription rather than representing one real Feed.
_Avoid_: inbox, everything, smart feed
