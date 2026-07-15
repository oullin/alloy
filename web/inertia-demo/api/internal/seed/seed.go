package seed

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"alloy.dev/inertia-demo/internal/database"
)

// SeedPasswordEnvVar names the environment variable that supplies the password
// for the seeded demo users. When unset, Run generates a random password per
// invocation so a deployed demo never ships known credentials.
const SeedPasswordEnvVar = "INERTIA_DEMO_SEED_PASSWORD"

// Run seeds the database using a demo-user password resolved from the
// environment, falling back to a freshly generated random password that is
// printed to stdout. Use RunWithPassword when a deterministic password is
// needed (e.g. tests).
func Run(db *sql.DB) error {
	password, err := resolveSeedPassword()

	if err != nil {
		return fmt.Errorf("seed password: %w", err)
	}

	return RunWithPassword(db, password)
}

// resolveSeedPassword returns the configured seed password, or a random one
// (printed to stdout) when SeedPasswordEnvVar is unset.
func resolveSeedPassword() (string, error) {
	if p := strings.TrimSpace(os.Getenv(SeedPasswordEnvVar)); p != "" {
		return p, nil
	}

	password, err := randomPassword()

	if err != nil {
		return "", err
	}

	fmt.Printf("seed: %s is unset; generated a random demo-user password: %s\n", SeedPasswordEnvVar, password)

	return password, nil
}

// randomPassword returns a URL-safe, base64-encoded 24-byte CSPRNG password.
func randomPassword() (string, error) {
	buf := make([]byte, 24)

	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random password: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// RunWithPassword seeds the database, assigning password to every seeded user.
func RunWithPassword(db *sql.DB, password string) error {
	if err := database.Truncate(db); err != nil {
		return fmt.Errorf("seed truncate: %w", err)
	}

	now := time.Now()

	if err := seedUsers(db, now, password); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}

	orgIDs, err := seedOrganizations(db)

	if err != nil {
		return fmt.Errorf("seed organizations: %w", err)
	}

	if err := seedContacts(db, orgIDs); err != nil {
		return fmt.Errorf("seed contacts: %w", err)
	}

	if err := seedNotes(db, now); err != nil {
		return fmt.Errorf("seed notes: %w", err)
	}

	if err := seedInvites(db, now); err != nil {
		return fmt.Errorf("seed invites: %w", err)
	}

	if err := seedUploads(db, now); err != nil {
		return fmt.Errorf("seed uploads: %w", err)
	}

	if err := seedApprovals(db, now); err != nil {
		return fmt.Errorf("seed approvals: %w", err)
	}

	if err := database.SetCounter(db, "priority_escalations", 18); err != nil {
		return fmt.Errorf("seed counter: %w", err)
	}

	return nil
}
