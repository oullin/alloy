// Command modzip builds the Go module artifacts that the Hara module proxy
// serves: a .zip, a .mod, and an .info per module per version, laid out as a
// GOPROXY file tree so the result can be verified offline with
// GOPROXY=file://<dir> before anything is uploaded.
//
// Two things make this more than `zip -r`. A module zip must prefix every entry
// with <module>@<version>/ and must not contain a go.mod outside the top level,
// so nested modules have to be excluded -- golang.org/x/mod/zip does both.
// And a replace directive is ignored by Go in a non-main module, so the local
// replace that pkg/hub/auth/passkeys needs during development must be stripped
// and its require pinned to the version being released, or consumers get an
// unresolvable hara.sh/alloy v0.0.0.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/sumdb/dirhash"
	"golang.org/x/mod/zip"
)

// target is one publishable module: where it lives and what it is called.
type target struct {
	dir  string
	path string
}

// rootModule is the module whose version the nested modules are pinned to.
const rootModule = "hara.sh/alloy"

var targets = []target{
	{dir: "pkg/hub", path: rootModule},
	{dir: "pkg/hub/auth/passkeys", path: rootModule + "/auth/passkeys"},
	{dir: "pkg/hub/queue/drivers/sqs", path: rootModule + "/queue/drivers/sqs"},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "modzip: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		repoRoot = flag.String("repo", ".", "repository root")
		version  = flag.String("version", "", "release version, e.g. v0.3.0")
		outDir   = flag.String("out", "", "output directory for the GOPROXY tree")
		stamp    = flag.String("time", "", "RFC3339 commit time for .info (default: now)")
	)

	flag.Parse()

	if *version == "" || *outDir == "" {
		return fmt.Errorf("both -version and -out are required")
	}

	if module.CanonicalVersion(*version) != *version {
		return fmt.Errorf("version %q is not canonical", *version)
	}

	published := time.Now().UTC()

	if *stamp != "" {
		parsed, err := time.Parse(time.RFC3339, *stamp)

		if err != nil {
			return fmt.Errorf("parsing -time: %w", err)
		}

		published = parsed.UTC()
	}

	root, err := filepath.Abs(*repoRoot)

	if err != nil {
		return err
	}

	for _, item := range targets {
		if err := build(root, item, *version, published, *outDir); err != nil {
			return fmt.Errorf("%s: %w", item.path, err)
		}
	}

	return nil
}

func build(root string, item target, version string, published time.Time, outDir string) error {
	source := filepath.Join(root, item.dir)

	if _, err := os.Stat(filepath.Join(source, "go.mod")); err != nil {
		return fmt.Errorf("no go.mod in %s: %w", item.dir, err)
	}

	// Stage a copy so the published go.mod can differ from the working tree's
	// without ever mutating the checkout.
	staging, err := os.MkdirTemp("", "modzip-")

	if err != nil {
		return err
	}

	defer os.RemoveAll(staging)

	staged := filepath.Join(staging, "module")

	if err := copyTree(source, staged); err != nil {
		return fmt.Errorf("staging: %w", err)
	}

	goMod, err := publishableGoMod(filepath.Join(staged, "go.mod"), version)

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(staged, "go.mod"), goMod, 0o644); err != nil {
		return err
	}

	escaped, err := module.EscapePath(item.path)

	if err != nil {
		return err
	}

	escapedVersion, err := module.EscapeVersion(version)

	if err != nil {
		return err
	}

	versionDir := filepath.Join(outDir, filepath.FromSlash(escaped), "@v")

	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return err
	}

	zipPath := filepath.Join(versionDir, escapedVersion+".zip")
	archive, err := os.Create(zipPath)

	if err != nil {
		return err
	}

	mv := module.Version{Path: item.path, Version: version}

	if err := zip.CreateFromDir(archive, mv, staged); err != nil {
		archive.Close()

		return fmt.Errorf("creating zip: %w", err)
	}

	if err := archive.Close(); err != nil {
		return err
	}

	// CheckZip re-reads the archive with the same rules the go command applies
	// on extract, so a layout mistake fails here rather than in a customer build.
	if _, err := zip.CheckZip(mv, zipPath); err != nil {
		return fmt.Errorf("invalid module zip: %w", err)
	}

	// The .mod endpoint must serve exactly these bytes forever: go.sum records
	// dirhash.Hash1 over them, so any later reformatting breaks every consumer.
	if err := os.WriteFile(filepath.Join(versionDir, escapedVersion+".mod"), goMod, 0o644); err != nil {
		return err
	}

	info, err := json.Marshal(struct {
		Version string
		Time    time.Time
	}{Version: version, Time: published})

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(versionDir, escapedVersion+".info"), info, 0o644); err != nil {
		return err
	}

	if err := appendVersion(filepath.Join(versionDir, "list"), version); err != nil {
		return err
	}

	zipHash, err := dirhash.HashZip(zipPath, dirhash.DefaultHash)

	if err != nil {
		return err
	}

	// Matches cmd/go's goModSum: Hash1 over a single entry literally named
	// "go.mod", not <module>@<version>/go.mod. Getting this wrong produces an
	// audit hash that silently disagrees with every consumer's go.sum.
	modHash, err := dirhash.Hash1([]string{"go.mod"}, func(string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(goMod)), nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n  zip    %s\n  go.mod %s\n", item.path, version, zipHash, modHash)

	return nil
}

// publishableGoMod drops the local replace directives, which Go ignores in a
// non-main module and which would leave the require pointing at v0.0.0, and
// pins any dependency on the root module to the version being released.
func publishableGoMod(path, version string) ([]byte, error) {
	raw, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	parsed, err := modfile.Parse(path, raw, nil)

	if err != nil {
		return nil, err
	}

	for _, rep := range parsed.Replace {
		if err := parsed.DropReplace(rep.Old.Path, rep.Old.Version); err != nil {
			return nil, fmt.Errorf("dropping replace %s: %w", rep.Old.Path, err)
		}
	}

	for _, req := range parsed.Require {
		if req.Mod.Path == rootModule || strings.HasPrefix(req.Mod.Path, rootModule+"/") {
			if err := parsed.AddRequire(req.Mod.Path, version); err != nil {
				return nil, fmt.Errorf("pinning %s: %w", req.Mod.Path, err)
			}
		}
	}

	parsed.Cleanup()

	return parsed.Format()
}

// appendVersion keeps @v/list sorted-by-arrival and free of duplicates so the
// same version can be rebuilt without corrupting the listing.
func appendVersion(path, version string) error {
	existing, err := os.ReadFile(path)

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for _, line := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(line) == version {
			return nil
		}
	}

	return os.WriteFile(path, append(existing, []byte(version+"\n")...), 0o644)
}

// copyTree copies regular files and directories only. Symlinks and irregular
// files are rejected by the module zip format anyway.
func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, path)

		if err != nil {
			return err
		}

		if relative == "." {
			return os.MkdirAll(destination, 0o755)
		}

		// .git never belongs in a module zip and can be very large.
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "node_modules") {
			return filepath.SkipDir
		}

		target := filepath.Join(destination, relative)

		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		data, err := os.ReadFile(path)

		if err != nil {
			return err
		}

		return os.WriteFile(target, data, 0o644)
	})
}
