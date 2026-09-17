# Audit 0001 — Security and correctness review (2026-09)

**Date:** 2026-09-16
**Scope:** `rssx-api` (Go backend) and `rssx-ui` (Vue 3 frontend), full read of the
routing layer, auth utilities, feed/news handlers, RSS sync pipeline, Redis
storage helpers, and the reading pane.
**Status:** findings recorded, remediation deferred.

This is a findings register, not a decision record. Each finding has a stable ID
so a later task spec (`rssx-api/docs/tasks/task-NNN-*.md`) can reference it.
Nothing here has been fixed yet.

---

## Summary

| ID | Severity | Area | Finding |
| --- | --- | --- | --- |
| [SEC-001](#sec-001--no-authentication-on-any-endpoint) | Critical | Backend / auth | No auth middleware; every endpoint is public |
| [SEC-002](#sec-002--all-users-share-one-hardcoded-account) | Critical | Backend / tenancy | Handlers hardcode `user.DefaultId = "0"` |
| [SEC-003](#sec-003--jwt-signing-key-committed-to-git-and-empty-in-the-container) | Critical | Backend / auth | Key is in version control; container default is empty |
| [SEC-004](#sec-004--xss-via-unsanitised-feed-html) | High | Frontend | `v-html` renders third-party feed HTML |
| [SEC-005](#sec-005--jwt-stored-in-localstorage) | Medium | Frontend | Token readable by any injected script |
| [BUG-001](#bug-001--rss-sync-can-stall-permanently) | High | Backend / sync | No HTTP timeout, blocking pool of 2, leaked pools |
| [SEC-006](#sec-006--tls-verification-disabled-process-wide) | High | Backend / sync | Mutates global `http.DefaultTransport` |
| [BUG-002](#bug-002--nil-pool-panic-crashes-the-process) | High | Backend / sync | `ants.NewPoolWithFunc` error discarded |
| [BUG-003](#bug-003--index-out-of-range-in-previousarticle) | Medium | Backend | `newsIds[0]` without a length check |
| [BUG-004](#bug-004--nil-type-assertions-after-redis-errors) | Medium | Backend | Error logged, execution continues onto `.([]byte)` |
| [BUG-005](#bug-005--userfeed-has-no-unique-constraint) | Medium | Backend / schema | Duplicate subscriptions possible |
| [BUG-006](#bug-006--logerror-used-with-a-format-string) | Low | Backend | `%v` printed literally |

---

## Critical

### SEC-001 — No authentication on any endpoint

**Where:** [`rssx-api/rssx-api.go`](../../rssx-api/rssx-api.go) — `setupRouter()`, lines 42–67

`setupRouter()` registers fourteen routes and attaches no middleware:

```go
router.GET("/feeds", feedHandler.LoadFeedList)
router.POST("/feed", feedHandler.AddFeed)
router.PUT("/feed/:id", feedHandler.UpdateFeed)
router.DELETE("/feed/:id", feedHandler.RemoveFeed)
router.DELETE("/feed/:id/purge", feedHandler.PurgeFeed)
router.GET("/news", list.LoadArticles)
```

`jwt.ParseToken()` is fully implemented and has thorough unit coverage in
[`utils/jwt/jwt_test.go`](../../rssx-api/utils/jwt/jwt_test.go) (expiry, bad signature,
malformed input, non-HMAC algorithms), but it is **never called from any handler
or middleware**. A grep for `ParseToken` outside the jwt package returns nothing.

**Impact:** anyone who can reach the service can list, modify, purge, and delete
feeds, and read all articles. `DELETE /feed/:id` needs no credentials at all.
The service is deployed on a public hostname.

**Note:** [`task-001`](../../rssx-api/docs/tasks/task-001-feed-management-api.md) states
"no auth middleware is required for now" and scopes auth out, so this is a known
unfinished item rather than a regression.
[`task-002`](../../rssx-api/docs/tasks/task-002-amazon-cognito-auth.md) (Cognito, P2,
Pending) is the planned replacement. The gap is that the service shipped publicly
before that task landed.

---

### SEC-002 — All users share one hardcoded account

**Where:**
[`feeds/feeds.go`](../../rssx-api/feeds/feeds.go) lines 38, 112, 123, 142 ·
[`feed/news/list/news-list.go`](../../rssx-api/feed/news/list/news-list.go) lines 291, 292, 300, 310, 351, 354, 361, 365, 386 ·
[`news/news.go`](../../rssx-api/news/news.go) line 94 ·
[`rss/gc.go`](../../rssx-api/rss/gc.go) line 27

Every data-access path substitutes the constant `user.DefaultId = "0"`
([`user/user.go:13`](../../rssx-api/user/user.go)) for the caller's identity:

```go
// feeds/feeds.go:38
userFeeds, err := h.repo.FindByUserID(user.DefaultId)

// feed/news/list/news-list.go:310, 354, 365
SetReadIndex(0, feedId, newIndex)
n.MarkRead(0)
```

**Impact:** subscriptions and per-article read state are global. One user marking
a page read changes it for everyone; one user unsubscribing removes the feed for
everyone. This is independent of SEC-001 — fixing the middleware alone would not
separate the data, because the handlers never consult the authenticated identity.

Any fix must thread a user ID from the Gin context through `feeds`, `news`, and
`list`. `task-002` §"User ID propagation" already anticipates this.

---

### SEC-003 — JWT signing key committed to git, and empty in the container

**Where:**
[`config.toml`](../../rssx-api/config.toml) line 3 ·
[`config-k8s.toml`](../../rssx-api/config-k8s.toml) line 3 ·
[`config-local.toml`](../../rssx-api/config-local.toml) line 3 ·
[`Containerfile`](../../rssx-api/Containerfile) line 54 ·
[`utils/jwt/jwt.go`](../../rssx-api/utils/jwt/jwt.go) lines 50, 87, 100

All three committed config files carry the same literal key:

```toml
security-key="fa6ee430-ebf6-44f2-a18b-64d691cd2dae"
```

The container image then overrides it with an empty string:

```dockerfile
ENV RSSX_SECURITY_KEY=""
```

and both the signing and parsing paths fall back to `""`:

```go
// jwt.go:50 (sign) and jwt.go:100 (verify)
config.GetString("rssx.security-key", "")
```

**Impact:** two separate problems. The committed key is public to anyone with
repo access, so tokens signed with it are forgeable. The empty container default
is worse — HMAC over an empty key means any party can mint a valid token for an
arbitrary user ID. Note `GetJwtToken`/line 87 reads a *different* config path
(`security-key`, not `rssx.security-key`), so the two token helpers disagree
about where the key lives.

**Also:** the key must be rotated, not just moved — it is already in git history.

---

## High

### SEC-004 — XSS via unsanitised feed HTML

**Where:** [`rssx-ui/src/components/ReadingPane.vue`](../../rssx-ui/src/components/ReadingPane.vue)

```vue
<!-- Feed-provided HTML; sanitising is a follow-up (docs/adr/0001). -->
<!-- eslint-disable-next-line vue/no-v-html -->
<div data-test="body" class="reading-pane__body" v-html="open.article.content" />
```

Article bodies are attacker-controllable end to end: `rss/sync.go` unmarshals the
feed XML, stores the description verbatim in Redis, the API returns it verbatim,
and the pane injects it with `v-html`. No sanitiser sits anywhere on that path.

**Impact:** any poisoned or hostile feed executes script in the app's origin.
The ESLint rule that would catch this is explicitly disabled inline, and the
comment acknowledges it as deferred work.

**Direction:** sanitise on render (DOMPurify) or on ingest; sanitising on ingest
alone is fragile because existing Redis content stays unsanitised.

---

### BUG-001 — RSS sync can stall permanently

**Where:** [`rss/sync.go`](../../rssx-api/rss/sync.go) lines 30–57

```go
func syncFeeds() {
	p, _ := ants.NewPoolWithFunc(2, syncOneFeed)
	// ...
	err := p.Invoke(oneFeed)
	// ...
	p.Release()
}

func syncOneFeed(data interface{}) {
	// ...
	client := &http.Client{}          // no Timeout
	result, err := client.Do(request)
```

Three compounding defects:

1. `http.Client{}` has no timeout, so a feed host that accepts the connection and
   never responds holds the worker forever.
2. The `ants` pool has two workers and blocks on `Invoke` by default. Two hung
   feeds block every subsequent `Invoke`, so the sync goroutine wedges and never
   reaches `p.Release()`.
3. `syncFeeds()` builds a fresh pool on every tick. Because wedged pools are
   never released, goroutines accumulate tick after tick.

**Impact:** feeds silently stop updating with no error surfaced, and the process
leaks goroutines until it is restarted. This is the most likely explanation for
any "feeds aren't refreshing" symptom.

**Direction:** set `Timeout` on the client (and a per-request `context`), raise
or reconsider the pool size, and `defer p.Release()`.

---

### SEC-006 — TLS verification disabled process-wide

**Where:** [`rss/sync.go`](../../rssx-api/rss/sync.go) line 49

```go
// insecure https
http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
client := &http.Client{}
```

This mutates the **shared** `http.DefaultTransport`, not a transport belonging to
this client. Because `client := &http.Client{}` leaves `Transport` nil, it uses
that same default — but so does every other outbound HTTPS call in the process.
It also runs on every feed sync, repeatedly, from pool workers.

**Impact:** certificate validation is off application-wide, enabling MITM against
any outbound request. Writing to a shared transport from concurrent workers is
also a data race.

**Direction:** if a specific feed genuinely needs a relaxed TLS config, give
*that* client its own `&http.Transport{}`; do not touch the default.

---

### BUG-002 — Nil pool panic crashes the process

**Where:** [`rss/sync.go`](../../rssx-api/rss/sync.go) line 31

```go
p, _ := ants.NewPoolWithFunc(2, syncOneFeed)
```

The error is discarded. On failure `p` is nil and the following `p.Invoke(...)`
panics.

**Impact:** worse than the handler-side panics below. This runs in a background
goroutine started by `go rss.Sync()`, where Gin's `Recovery` middleware does not
apply, so the panic takes down the whole process rather than returning a 500.

---

## Medium

### BUG-003 — Index out of range in `PreviousArticle`

**Where:** [`feed/news/list/news-list.go`](../../rssx-api/feed/news/list/news-list.go) lines 318–324

```go
index := FindIndexById(feedId, currentNewsId)
newsIds := FindNewsListByRange(NewsListKey(feedId), index-1, index-1)
previousNewsId := newsIds[0]
```

No length check. When the current article is the first in the feed (or the ID is
unknown, making `index` 0 and the range negative) the slice is empty and the
index panics.

**Impact:** 500 on a reachable path. `gin.Default()` includes `Recovery`, so the
process survives, but the request fails and the stack trace is logged.

Related: `feedId, _ := strconv.Atoi(c.Query("feedId"))` on line 320 silently
yields `0` for malformed input, so a bad query parameter quietly reads the wrong
feed instead of returning 400. The same pattern appears at lines 280 and 336.

---

### BUG-004 — Nil type assertions after Redis errors

**Where:**
[`news/news.go`](../../rssx-api/news/news.go) lines 103–108 ·
[`storage/redisx/redis.go`](../../rssx-api/storage/redisx/redis.go) lines 117–131

Both sites log the error and then continue as though it had not happened:

```go
// news/news.go:103
result, err := redis.Values(redisx.Exec("HMGET", ...))
if err != nil {
	log.Info(err.Error())   // logged, not returned
}
n.Title = string(result[0].([]byte))   // result is nil here
```

```go
// storage/redisx/redis.go:123
result, err := Exec("ZRANGE", key, rank, rank)
if err != nil {
	log.Errorf("failed to get news by rank: %v", err)
}
foo := result.([]interface{})          // panics on nil
```

`GetNewsIdListByScore` just above guards with `if r != nil`, so the surrounding
code already knows the pattern — `GetScoreByRank` simply omits it.

**Impact:** a Redis hiccup or a missing article key turns into a panic instead of
an error response. Recovered by Gin on handler paths; the severity of
`news.Load()` rises if it is ever called from a background goroutine.

---

### SEC-005 — JWT stored in localStorage

**Where:** [`rssx-ui/src/utils/auth.ts`](../../rssx-ui/src/utils/auth.ts)

```ts
const localStorageTokenKey = 'token'
export function getJwtToken (): string | null {
  return localStorage.getItem(localStorageTokenKey)
}
```

`localStorage` is readable by any script running in the origin.

**Impact:** on its own this is a common, accepted trade-off. Combined with
SEC-004 it completes an attack chain: a hostile feed injects script, the script
reads the token, and the token leaves the browser. Treat SEC-004 and SEC-005 as
one unit when prioritising.

---

### BUG-005 — `UserFeed` has no unique constraint

**Where:** [`common/sqlite.go`](../../rssx-api/common/sqlite.go) lines 45–50

```go
type UserFeed struct {
	UserId string `gorm:"index;not null"`
	FeedId int64  `gorm:"index;not null"`
	Sort   int    `gorm:"default:0"`
}
```

No primary key and no unique index on `(UserId, FeedId)`. `AddFeed`
([`feeds/feeds.go`](../../rssx-api/feeds/feeds.go) lines 112–124) does a
check-then-insert with no transaction:

```go
subscribed, err := h.repo.IsSubscribed(user.DefaultId, f.Id)
if subscribed { /* 409 */ }
if err := h.repo.Subscribe(user.DefaultId, f.Id); err != nil {
```

**Impact:** concurrent requests can both pass the check and insert, producing
duplicate subscriptions and duplicated articles in the list. The database cannot
reject them because the constraint does not exist.

---

## Low

### BUG-006 — `log.Error` used with a format string

**Where:** [`rss/sync.go`](../../rssx-api/rss/sync.go) line 79

```go
log.Error("failed to unmarshal: %v", err)
```

`log.Error` is not the formatting variant; `log.Errorf` is. The verb is printed
literally and the error value is dropped or appended unformatted.

**Impact:** cosmetic, but it hides the actual parse error on a path that matters
when a feed changes shape.

---

## Suggested remediation order

Sequencing reflects dependency and effort rather than raw severity.

1. **SEC-003** — rotate the key, move it out of git, fail fast on an empty key.
   Cheap, and a prerequisite for SEC-001 being worth anything.
2. **BUG-001 + SEC-006 + BUG-002** — one small, self-contained pass over
   `rss/sync.go`. Restores feed updates and removes a process-crash path.
3. **SEC-001 + SEC-002** — the real work. Middleware plus threading a user ID
   through three packages. Overlaps heavily with
   [`task-002`](../../rssx-api/docs/tasks/task-002-amazon-cognito-auth.md); decide whether
   to land an interim local-JWT middleware or wait for Cognito.
4. **SEC-004 + SEC-005** — sanitise feed HTML, then revisit token storage.
5. **BUG-003 / BUG-004 / BUG-005** — defensive fixes; fold into whichever task
   touches those files next.

## Not covered by this review

- No SQL injection was found; the GORM paths use parameter binding throughout.
- Password hashing in `user/` was not examined in depth.
- The frontend was reviewed only for the reading path and token handling.
- No dependency CVE scan was run on `go.mod` or `pnpm-lock.yaml`.
