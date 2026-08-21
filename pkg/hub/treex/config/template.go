package config

// Template is the starter file treex config init writes.
//
// It is deliberately almost entirely commented out. The built-in catalog
// already knows every agent treex supports, and a file that restated all of it
// would go stale the moment a provider changed; what a user actually needs is a
// small file that overrides one or two things.
const Template = `# treex configuration
#
# Everything here is optional. treex ships with a catalog of the agent tools it
# knows about, and this file patches that catalog rather than replacing it, so
# you only need to write the parts you want to change.
#
#   treex providers    show the catalog and where each provider resolves to
#   treex scan         measure without touching anything
#   treex clean        show what would be removed (add --apply to act)

version: 1

defaults:
    # Worker count for the size walk. 0 derives it from the machine.
    jobs: 0

    # Ignore anything touched more recently than this. Staleness is measured
    # from the newest file anywhere in the tree, so a worktree you were in an
    # hour ago is never a candidate.
    older-than: 7d

    # Skip candidates smaller than this.
    # min-size: 100MB

safety:
    # Each of these can be turned off, but each one exists to stop treex
    # deleting something you cannot get back.
    require-clean-worktree: true
    require-pushed-commits: true
    require-no-stash: true

    protected-branches: [main, master, develop]

    # Never deleted, and never descended into. Additions are merged with the
    # built-in guards rather than replacing them.
    protect-paths:
        - ~/Sites
        - ~/Documents

# Directories treated as regenerable build output. This list does double duty:
# it is what "clean --categories artifact" removes, and it is what treex ignores
# when deciding whether a worktree holds real uncommitted work.
#
# artifacts:
#     names: [node_modules, dist, build, .turbo, target, vendor, Pods]

# Providers patch the built-in catalog by name. To switch one off:
#
# providers:
#     - name: cursor
#       enabled: false
#
# To add a directory to one treex already knows:
#
#     - name: codex
#       caches: [cache, runtimes, android, maestro, my-extra-cache]
#
# To describe a tool treex has never heard of:
#
#     - name: my-agent
#       root: ~/.my-agent
#       worktrees:
#           - { path: worktrees, depth: 2 }
#       caches: [cache]
`
