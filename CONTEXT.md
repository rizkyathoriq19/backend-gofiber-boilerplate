# Authentication

Vocabulary for independent user logins and their revocation.

## Language

**Login session**:
A user's authenticated presence established by one successful sign-in, independent of that user's other sign-ins. A login session is not a physical device or an individual token.
_Avoid_: Device, token (when referring to the whole login session)

**Logout**:
Termination of the current login session, without terminating the user's other login sessions. Its credentials no longer authorize subsequent requests; requests already underway may finish.
_Avoid_: Logout all (when referring to ordinary logout)

**Session renewal**:
Extension of a login session's lifetime in exchange for its current refresh token, which can be used for renewal only once. Repeating that exchange is rejected without terminating the successfully renewed login session.
_Avoid_: New login (when referring to renewal of an existing login session)
