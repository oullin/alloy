package session

import csession "alloy.dev/backend/contracts/session"

// Handler abstracts the session storage backend.
// SessionHandlerInterface.
type Handler = csession.Handler

// ExistenceAware is implemented by handlers that can distinguish between
// inserting a new session and updating an existing one (e.g. DatabaseHandler).
type ExistenceAware = csession.ExistenceAware

// Encrypter encrypts and decrypts session payloads.
type Encrypter = csession.Encrypter

// RequestAware is implemented by handlers that need the current HTTP request
// (e.g. CookieHandler).
type RequestAware = csession.RequestAware

// Cache is the minimal cache interface required by CacheHandler.
type Cache = csession.Cache
