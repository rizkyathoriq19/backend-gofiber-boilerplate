---
status: accepted
---

# Redis-backed login sessions with fail-closed validation

Authentication currently splits issuance, refresh-token storage, and access-token revocation across modules with incompatible revocation meanings. Deepen the existing security package into a Login session module whose interface owns issuance, access validation, session renewal, and logout, keeping JWT encoding and the Redis adapter inside its implementation. This concentrates token rules and verification at one seam rather than adding another pass-through module.

**Implementation status:** Implemented.

## Agreed behavior

- Each successful sign-in establishes an independent login session. Tokens carry that session's identity; Redis holds its active state and current refresh-token identity.
- Logout terminates only the current login session. Every access and refresh token from that session is rejected on subsequent requests; requests already underway may finish. Other login sessions remain valid.
- Session renewal atomically replaces the current refresh-token identity and extends the session by the configured refresh lifetime. Exactly one concurrent exchange of a refresh token succeeds.
- Reusing a consumed refresh token is rejected without terminating the successfully renewed session. No grace window or replay response is retained. A lost successful refresh response can require the affected client to sign in again.
- Access validation checks both token validity and active session state. Invalid, expired, revoked, or legacy tokens without a session identity receive HTTP 401. Failure to determine revocation state because Redis is unavailable receives HTTP 503, using the existing error envelope; it does not grant access.
- Existing HTTP routes and request/response field names remain unchanged. Token lifetimes come from configuration, and `expires_in` reflects the issued access token's actual lifetime rather than a hardcoded duration.
- A one-time re-login during rollout is acceptable. Legacy-token compatibility is not part of this change.

## Scope and verification

Include the narrow typed-cache correction in user lookup because the existing warm-cache type mismatch prevents the production refresh flow from succeeding. Keep broader cache redesign, RBAC changes, WebSocket changes, logout-all, and a dependency-injection framework out of scope.

Tests cross the production module interface and verify session isolation, logout, atomic concurrent renewal, rejected reuse, lost-response behavior, Redis outages, expiry, and warm-cache refresh. Replace copied authentication behavior in tests rather than maintaining a second token implementation. Real Redis verifies atomicity; the existing CI workflow provides Redis 7 and race-enabled Go tests. Local integration verification requires an isolated Redis instance and must not use application data stores.

The implementation must wire both configured access and refresh lifetimes into token generation and refresh-token storage. The current startup path passes only the access lifetime, while the handlers hardcode `expires_in` to 24 hours; both are part of the correction. Existing refresh and logout cache failures currently become HTTP 500, so the implementation must map revocation-storage outages to the existing `ServiceUnavailable` or `AuthServiceUnavailable` 503 error codes.

## Trade-offs

- **Availability versus revocation:** fail-closed validation makes protected requests dependent on Redis availability. This is preferred to allowing revoked credentials during an outage.
- **Retry convenience versus simplicity:** strict single-use refresh avoids a replay window, accepting re-login after a lost successful response. A duplicate request does not revoke the winning session.
- **Compatibility versus a coherent model:** a one-time re-login avoids maintaining both legacy token behavior and session-aware validation.

## Source evidence

- `internal/module/auth/service.go:73–89`: refresh retains account lookup while delegating session validation and atomic renewal.
- `internal/shared/security/login_session.go`: login-session issuance, active-state validation, single-use renewal, logout, and the Redis atomic compare-and-swap adapter now have one owner.
- `internal/middleware/auth.go`: protected requests validate both JWT claims and active session state, failing closed with HTTP 503 when Redis is unavailable.
- `internal/module/auth/repository.go` and `internal/shared/utils/cache.go`: user lookup uses the existing typed cache-aside helper for warm-cache refresh support.
- `internal/module/auth/handler.go`: response expiry is supplied from configured access-token lifetime.
- `internal/config/config.go`, `cmd/server/main.go`, and `internal/shared/security/jwt_manager.go`: configured access and refresh lifetimes are wired through token generation and session storage.
- `internal/module/auth/service.go`: authentication delegates token lifecycle rules to the login-session module while retaining account lookup and password verification.
- `internal/shared/enum/error_http_code.go:34–43`: `ServiceUnavailable` and `AuthServiceUnavailable` already map to HTTP 503.
- `.github/workflows/test.yml:29–36,64`: Redis 7 and race-enabled test execution are already configured.

Domain vocabulary is defined in [CONTEXT.md](../../CONTEXT.md).
