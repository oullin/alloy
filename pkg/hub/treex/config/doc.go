// Package config is treex's public configuration contract: the YAML schema
// users write and the typed values the rest of the program reads.
//
// A configuration file never starts from nothing. Default() carries a catalog
// of the agent tools treex knows about, and a file layered on top patches that
// catalog field by field, so a user who only wants to disable one provider
// writes three lines rather than restating every root treex should sweep.
//
// Nothing in this package touches the filesystem beyond reading the config
// file itself. Path expansion happens in Resolve, and the values it returns are
// absolute and cleaned, so no consumer downstream ever handles a "~".
package config
