# Task Spec: #001 — Feed Management API

**Status:** Done
**Priority:** P1

---

## Summary

Add REST API endpoints for managing RSS feed subscriptions. Currently feeds can only be managed by directly editing the SQLite database. This task adds `POST /feed` and `DELETE /feed/:id` endpoints.

---

## Background

The existing `GET /feeds` endpoint returns the feed list for the default user (`user.DefaultId = "0"`). All new endpoints follow the same single-user model — no auth middleware is required for now.

Data model:
- `feeds` table — global feed registry (`Id`, `Title`, `Url`)
- `user_feeds` table — subscription mapping (`UserId`, `FeedId`, `Sort`)

---

## Endpoints

### POST /feed — Subscribe to a feed

Add a new RSS feed and subscribe the default user to it.

**Request**

```
POST /feed
Content-Type: application/json
```

```json
{
  "url":   "string, required, valid http/https URL",
  "title": "string, required, non-empty"
}
```

**Validation:**
- `url` required, must start with `http://` or `https://`
- `title` required, non-empty string
- If `url` already exists in `feeds` table, reuse that record (do not create duplicate)
- If user is already subscribed to this feed, return 409 Conflict

**Success Response — 201 Created**

```json
{
  "id":    14,
  "title": "InfoQ CN",
  "url":   "https://www.infoq.cn/feed"
}
```

**Error Responses**

| Status | Condition |
| --- | --- |
| 400 | Missing or invalid fields |
| 409 | User already subscribed to this feed |

---

### DELETE /feed/:id — Unsubscribe from a feed

Remove the default user's subscription to a feed. Does not delete the feed from the global `feeds` table.

**Request**

```
DELETE /feed/:id
```

- `:id` — integer, the feed ID

**Success Response — 204 No Content**

**Error Responses**

| Status | Condition |
| --- | --- |
| 400 | `:id` is not a valid integer |
| 404 | Subscription not found for default user |

---

## Implementation Plan

1. Add handler functions in `feeds/feeds.go`
   - `AddFeed(c *gin.Context)` — handles POST /feed
   - `RemoveFeed(c *gin.Context)` — handles DELETE /feed/:id

2. Register routes in `rssx-api.go`
   - `router.POST("/feed", feeds.AddFeed)`
   - `router.DELETE("/feed/:id", feeds.RemoveFeed)`

3. Logic for `AddFeed`:
   - Parse and validate request body
   - Upsert into `feeds` table (find by URL or create)
   - Check if `user_feeds` record already exists → 409 if yes
   - Insert into `user_feeds` with `user_id = user.DefaultId`
   - Return 201 with the feed record

4. Logic for `RemoveFeed`:
   - Parse `:id` as int64
   - Delete from `user_feeds` where `user_id = user.DefaultId AND feed_id = :id`
   - Return 404 if no rows affected, 204 otherwise

---

## Out of Scope

- Auth / per-user feed management (all operations use `user.DefaultId`)
- `PUT /feed/:id` for updating feed title/URL
- Feed validation (verifying the URL is a valid RSS feed)
- Triggering an immediate RSS sync after adding a feed
