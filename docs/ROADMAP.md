# RSSX Roadmap

Planned upgrades that are agreed in direction but not scheduled. Each entry
states where we are today, what has to be true first, and what triggers the work.

---

## Authentication: local password now, Clerk later

**Status:** Planned (decided 2026-09-29). Not scheduled.
**Supersedes:** [`task-002`](../rssx-api/docs/tasks/task-002-amazon-cognito-auth.md) (Cognito + Casdoor).

### Today — local username/password

- `POST /register` and `POST /login` in `rssx-api/user/`; passwords hashed with bcrypt
  in the SQLite `users` table.
- rssx-api signs its own HS256 JWT (`rssx.security-key`, env `RSSX_SECURITY_KEY`),
  valid for one day.
- `jwt.RequireAuth()` guards every route except `/ping`, `/login`, `/register` and
  stores the token's user id in the gin context under `jwt.ContextKeyUserId`.
- rssx-ui keeps the token in `localStorage`, sends it as `Authorization: Bearer`,
  and returns to `/login` on 401.

This stays the auth model until the trigger below fires. Keep new code behind
the same seam — handlers take the user id from the gin context, never from the
token or the request directly — so swapping the token issuer later does not
touch business logic.

### Target — Clerk, shared with ENX (Catglish)

ENX already uses Clerk ([enx ADR-015](https://github.com/wiloon/enx/blob/main/docs/architecture/adr-015-cognito-to-clerk-auth-migration.md)).
RSSX joins **the same Clerk application**, so one Clerk user id (`sub`) identifies
a person in both apps and a user signed in to ENX is signed in to RSSX (SSO).
A separate Clerk application would mean a separate user pool and no SSO.

Whether the session is shared silently depends on the domains:

| Setup | Behaviour | Cost |
| --- | --- | --- |
| Same root domain (e.g. `catglish.com` + `rss.catglish.com`) | Clerk shares the session across subdomains by default | Free plan |
| Different root domains (e.g. `catglish.com` + `rssx.wiloon.com`) | RSSX is a Clerk **satellite domain**: "Sign in" bounces to ENX's sign-in page and straight back if already signed in; sign-in/sign-up happen on ENX; sign-out is global | Paid plan in production; free on development instances |

See Clerk: [Authentication across different domains](https://clerk.com/docs/guides/dashboard/dns-domains/satellite-domains).

### Prerequisites

1. ~~**Per-user data (audit [SEC-002](audit/0001-security-and-correctness-review-2026-09.md#sec-002--all-users-share-one-hardcoded-account)).**~~
   Done 2026-09-29: subscriptions and read state are keyed on the user id from
   the gin context, so swapping in Clerk only changes where that id comes from.
2. **ENX production domain settled.** Decides free subdomain sharing vs a paid
   satellite domain (see table).

### Steps

1. **Homelab on the Clerk development instance** (`rational-deer-4450`, the one
   enx-ui-lab uses). ENX and RSSX are both under `wiloon.com` there; if the
   session is not shared automatically, configure RSSX as a satellite (free in
   development).
2. **rssx-api:** replace `RequireAuth`'s local HS256 check with Clerk session JWT
   verification — JWKS, `iss`, `exp`/`nbf`, `azp` allowlist — mirroring enx-api's
   `clerk_auth.go`. Key local users on the Clerk user id (`users.clerk_user_id`,
   provisioned on first request, like ENX's `GetOrCreateByClerkUserId`).
3. **Access control:** ENX is public, so any ENX user could reach RSSX. Add an
   allowlist (Clerk user id or email) unless RSSX is opened to everyone.
4. **rssx-ui:** adopt `@clerk/vue` (RSSX is a client-only SPA, which satellite
   domains support); get tokens from Clerk instead of `localStorage`.
5. **Data migration:** move the existing user-`"0"` subscriptions and Redis read
   state (`read_index:0:*`, `read_mark:0:*`) to the owner's Clerk-backed user id.
6. **Add the RSSX origin** to Clerk's authorized parties / `CLERK_AUTHORIZED_PARTIES`.
7. **Retire local auth:** remove `/login`, `/register`, the password column, the
   HS256 signing path, and `RSSX_SECURITY_KEY`.
8. **Production:** switch to the production instance (`clerk.catglish.com`) with
   the domain setup chosen in prerequisite 2.

### Trigger

Start when any of these becomes true:

- RSSX gets users other than the owner, and managing their passwords is a burden.
- SSO with ENX/Catglish is wanted in daily use (e.g. RSSX offered alongside Catglish).
- Social login (Google/GitHub) is needed.

### Costs and risks to recheck then

- RSSX users count toward the shared Clerk MAU quota (free up to 10k MAU).
- Satellite domains need a paid Clerk plan in production.
- Vendor lock-in is inherited from ENX; ENX ADR-015 keeps Logto as the exit path.
