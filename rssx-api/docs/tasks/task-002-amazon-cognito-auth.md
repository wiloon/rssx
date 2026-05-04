# Task Spec: #002 — Social Login via Cognito (Google native + WeChat via Casdoor)

**Status:** Pending
**Priority:** P2
**Depends on:** None (can be done independently of #001)

---

## Summary

Replace the current self-managed username/password authentication with Amazon Cognito as the central auth layer. Google login uses Cognito's native social IdP federation. WeChat login is routed through **Casdoor** (self-hosted, natively supports WeChat OAuth) which federates into Cognito as an OIDC IdP. The rssx-api validates only Cognito-issued tokens via JWKS regardless of which login method the user chose.

---

## Architecture

```
                    ┌─────────────────────────────┐
                    │         rssx-ui              │
                    └────────────┬────────────────-┘
                                 │ login redirect
                                 ▼
                    ┌─────────────────────────────┐
                    │      Amazon Cognito          │
                    │  (User Pool + Hosted UI)     │
                    └──────┬──────────────┬───────-┘
                           │              │
              Google IdP   │              │  OIDC federation
              (native)     │              │
                           ▼              ▼
                      Google OAuth  ┌──────────────┐
                                    │   Casdoor    │ (k8s pod, self-hosted)
                                    └──────┬───────┘
                                           │ WeChat OAuth2 (native)
                                           ▼
                                       WeChat IDP
```

- **Google login**: Cognito natively supports Google OAuth — no Casdoor needed
- **WeChat login**: Casdoor natively wraps WeChat OAuth into standard OIDC; Cognito federates with Casdoor as an external OIDC IdP
- **rssx-api** only validates Cognito-issued Access Tokens (RS256 via JWKS) — no knowledge of Casdoor or individual social providers
- **rssx-ui** uses Cognito Hosted UI — no direct dependency on Casdoor

---

## Why Casdoor over Dex for WeChat

| | Casdoor | Dex |
| --- | --- | --- |
| WeChat support | ✅ Native, built-in | ❌ Community plugin only, unmaintained |
| Config | UI + API | YAML file only |
| User management UI | ✅ Yes | ❌ No |
| Language | Go | Go |
| Memory | ~100MB | ~50MB |

---

## Phased Delivery

| Phase | Scope | Casdoor required |
| --- | --- | --- |
| **v1** | Cognito + Google login | No |
| **v2** | Add WeChat login via Casdoor | Yes |

v1 can be shipped independently. v2 adds Casdoor on top without changing rssx-api or rssx-ui.

---

## Background

### Current auth flow

```
POST /register → bcrypt hash → store in SQLite users table → return self-signed HS256 JWT
POST /login    → bcrypt verify against SQLite          → return self-signed HS256 JWT
All protected routes → validate HS256 JWT with local secret key
```

### Target auth flow

```
[Google]  rssx-ui → Cognito Hosted UI → Google OAuth → Cognito → rssx-ui
[WeChat]  rssx-ui → Cognito Hosted UI → Casdoor → WeChat OAuth → Casdoor → Cognito → rssx-ui

Both: rssx-ui → rssx-api with Authorization: Bearer <cognito_access_token>
      rssx-api → Cognito JWKS → verify RS256 → extract sub (user ID)
```

The `POST /login` and `POST /register` endpoints in rssx-api are removed.

---

## Component 1 (v1): Cognito — Google native federation

Add Google as a built-in social IdP in Cognito. No Casdoor required.

New resource in `aws/opentofu/rssx/main.tf`:

```hcl
resource "aws_cognito_identity_provider" "google" {
  user_pool_id  = aws_cognito_user_pool.rssx.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    client_id        = var.google_client_id
    client_secret    = var.google_client_secret
    authorize_scopes = "openid email profile"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
  }
}
```

Update App Client to allow Google:

```hcl
resource "aws_cognito_user_pool_client" "rssx_api" {
  # ... existing config ...
  supported_identity_providers = ["COGNITO", "Google"]
  allowed_oauth_flows          = ["code"]
  allowed_oauth_scopes         = ["openid", "email", "profile"]
  callback_urls                = ["https://rssx.wiloon.com/callback"]
  logout_urls                  = ["https://rssx.wiloon.com/logout"]
}
```

Google OAuth credentials are created in Google Cloud Console and passed via `terraform.tfvars`.

---

## Component 1 (v2): Casdoor — WeChat federation

Casdoor is deployed in the homelab k8s cluster. Required only for WeChat login.

### Casdoor configuration

In Casdoor admin UI:
1. Create a new **Provider** of type `WeChat` — enter WeChat App ID and App Secret
2. Create an **Application** with the WeChat provider attached
3. Note the Casdoor OIDC issuer URL: `https://casdoor.wiloon.com`
4. Create an **OAuth client** for Cognito with redirect URI:
   `https://<cognito_domain>.auth.<region>.amazoncognito.com/oauth2/idpresponse`

Add Casdoor as an OIDC IdP in Cognito (new OpenTofu resource):

```hcl
resource "aws_cognito_user_pool_identity_provider" "casdoor" {
  user_pool_id  = aws_cognito_user_pool.rssx.id
  provider_name = "Casdoor"
  provider_type = "OIDC"

  provider_details = {
    client_id                 = var.casdoor_client_id
    client_secret             = var.casdoor_client_secret
    attributes_request_method = "GET"
    oidc_issuer               = "https://casdoor.wiloon.com"
    authorize_scopes          = "openid email profile"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
  }
}
```

Update App Client to include Casdoor:

```hcl
supported_identity_providers = ["COGNITO", "Google", "Casdoor"]
```

---

## Component 2: rssx-api changes

### Configuration

Add to `config.toml` / `config-k8s.toml`:

```toml
[cognito]
region       = "us-east-1"
user_pool_id = "us-east-1_xxxxxxxx"
client_id    = "xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

K8s env vars:

```
COGNITO_REGION
COGNITO_USER_POOL_ID
COGNITO_CLIENT_ID
```

### Removed endpoints

| Method | Path | Reason |
| --- | --- | --- |
| POST | /login | Replaced by Cognito |
| POST | /register | Replaced by Cognito |

### Token validation middleware

JWKS URL:
```
https://cognito-idp.<region>.amazonaws.com/<user_pool_id>/.well-known/jwks.json
```

Middleware requirements:
1. Fetch and cache JWKS (1-hour TTL, retry with fresh keys on signature failure)
2. Validate: RS256 signature, `exp`, `iss`, `aud` (matches `client_id`), `token_use = "access"`
3. Extract `sub` claim as user ID, set on Gin context

---

## Implementation Plan

### v1 — Google login (no Casdoor)

1. **OpenTofu** — add `aws_cognito_identity_provider` for Google, add User Pool Domain, update App Client (`supported_identity_providers`, `callback_urls`)
2. **rssx-api** — add `utils/cognito` package (JWKS fetch + RS256 validation), add `RequireAuth` middleware, remove `/login` and `/register` routes
3. **User ID propagation** — replace `user.DefaultId` with user ID from Gin context in `feeds` package
4. **rssx-ui** — redirect to Cognito Hosted UI for login, handle callback, store and attach access token
5. **Cleanup** — archive `user/login.go`, `user/register.go`, `user/user.go`, `utils/jwt/jwt.go`

### v2 — WeChat login (requires Casdoor)

1. Deploy Casdoor in k8s, configure WeChat provider in Casdoor admin UI
2. **OpenTofu** — add `aws_cognito_user_pool_identity_provider` for Casdoor, update App Client to include `"Casdoor"` in `supported_identity_providers`
3. rssx-api and rssx-ui require no additional changes

---

## Security Considerations

- Never store Cognito `client_secret` or Casdoor `client_secret` in client-side code
- JWKS must be fetched over HTTPS; validate `iss` matches Cognito endpoint
- Cache JWKS but support key rotation (retry with fresh JWKS on signature failure)
- Use Access Token (not ID Token) for API authorization
- Casdoor `client_secret` for Cognito must be stored as a k8s Secret, not in config files

---

## Known Risks

| Risk | Mitigation |
| --- | --- |
| Casdoor becomes a single point of failure for WeChat login | Run 2 replicas; Cognito sessions persist after Casdoor restarts |
| Two-hop redirect (WeChat → Casdoor → Cognito → app) increases login latency | Acceptable for homelab |
| WeChat OAuth requires verified public account + domain registration | Evaluate WeChat account requirements before starting v2 |

---

## Out of Scope

- WeChat Mini Program login (different protocol from WeChat OAuth)
- Cognito Lambda triggers
- Fine-grained authorization / Cognito groups
- Migrating existing SQLite users to Cognito
