# Storage Contracts

These schemas describe the storage shape expected by the auth package. SQL
repositories accept custom table names, but unsafe identifiers fall back to the
defaults shown here.

## Password Reset Tokens

```sql
CREATE TABLE password_reset_tokens (
  email      TEXT PRIMARY KEY,
  token      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
```

`token` stores a SHA-256 hash, not the plaintext reset token.

## Sessions

The database session handler and browser-session repository expect:

```sql
CREATE TABLE sessions (
  id            TEXT PRIMARY KEY,
  user_id       TEXT,
  ip_address    TEXT,
  user_agent    TEXT,
  payload       TEXT NOT NULL,
  last_activity BIGINT NOT NULL
);

CREATE INDEX sessions_user_id_last_activity
  ON sessions (user_id, last_activity DESC);
```

## Personal Access Tokens

```sql
CREATE TABLE personal_access_tokens (
  id          TEXT PRIMARY KEY,
  user_id     TEXT NOT NULL,
  name        TEXT NOT NULL,
  token_hash  TEXT NOT NULL,
  abilities   TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL,
  last_used_at TIMESTAMPTZ,
  expires_at   TIMESTAMPTZ,
  revoked_at   TIMESTAMPTZ
);

CREATE INDEX personal_access_tokens_user_id
  ON personal_access_tokens (user_id);
```

`abilities` stores JSON. `token_hash` stores SHA-256 of the secret part of the
`id|secret` plaintext token.

## WebAuthn Users

```sql
CREATE TABLE webauthn_users (
  rpid    TEXT NOT NULL,
  user_id TEXT NOT NULL,
  handle  BYTEA NOT NULL,
  PRIMARY KEY (rpid, user_id)
);

CREATE UNIQUE INDEX webauthn_users_handle
  ON webauthn_users (rpid, handle);
```

The handle is the stable WebAuthn user handle. It is not the application user
ID and should not be displayed.

## WebAuthn Credentials

```sql
CREATE TABLE webauthn_credentials (
  rpid          TEXT NOT NULL,
  user_id       TEXT NOT NULL,
  credential_id BYTEA NOT NULL,
  credential    BYTEA NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at  TIMESTAMPTZ,
  PRIMARY KEY (rpid, credential_id)
);

CREATE INDEX webauthn_credentials_user_id
  ON webauthn_credentials (rpid, user_id);
```

`credential` stores the serialized `webauthn.Credential`. For higher auditability
or database-level policy enforcement, split the credential into explicit columns
for public key, sign count, flags, attestation data, and transports.

## WebAuthn Ceremony Sessions

```sql
CREATE TABLE webauthn_sessions (
  rpid       TEXT NOT NULL,
  id         TEXT NOT NULL,
  challenge  TEXT NOT NULL,
  payload    BYTEA NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (rpid, id)
);

CREATE UNIQUE INDEX webauthn_sessions_challenge
  ON webauthn_sessions (rpid, challenge);

CREATE INDEX webauthn_sessions_expires_at
  ON webauthn_sessions (expires_at);
```

Rows must be scoped to the same relying party ID used by WebAuthn verification.
Expired rows should be removed with `SQLSessionStore.DeleteExpired`.

## Teams

`teams.SQLRepository` persists the following logical shape. The in-memory
repository is included for tests and local apps.

```sql
CREATE TABLE teams (
  id         TEXT PRIMARY KEY,
  name       TEXT NOT NULL,
  owner_id   TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE team_members (
  team_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  role    TEXT NOT NULL,
  PRIMARY KEY (team_id, user_id)
);

CREATE TABLE current_teams (
  user_id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL
);
```

Member-management operations must be scoped through the actor's membership and
role permissions.
