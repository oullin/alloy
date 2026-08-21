# treex is built by goreleaser, which cross-compiles the binary and passes it
# in; this image only assembles the runtime.
#
# git is the one runtime dependency that genuinely matters. treex removes a
# linked worktree by asking git to do it, and without git in the image every
# linked worktree would be reported as un-removable.
FROM alpine:3.22

RUN apk add --no-cache git ca-certificates \
    && adduser -D -u 1000 treex

COPY treex /usr/local/bin/treex

# treex reads a mounted host home rather than the container's own filesystem:
#
#   docker run --rm -v "$HOME:/host" ghcr.io/oullin/treex scan --root /host
#
# A read-write mount is what --apply needs, and is also the most dangerous way
# to run this. Mount :ro for scanning, which is the common case.
WORKDIR /host

USER treex

ENTRYPOINT ["/usr/local/bin/treex"]
CMD ["scan", "--root", "/host"]
