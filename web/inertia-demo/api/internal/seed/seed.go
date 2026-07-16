package seed

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"alloy.dev/inertia-demo/internal/database"
)

// SeedPasswordEnvVar names the environment variable that supplies the password
// for the seeded demo users. When unset, Run falls back to
// DefaultSeedPassword.
const SeedPasswordEnvVar = "INERTIA_DEMO_SEED_PASSWORD"

// DefaultSeedPassword is the well-known password assigned to the seeded demo
// users when SeedPasswordEnvVar is unset. It is deliberately a known demo
// credential (owner decision, for demo convenience); set the environment
// variable to override it.
const DefaultSeedPassword = "12345678"

// Run seeds the database using a demo-user password resolved from the
// environment, falling back to DefaultSeedPassword. Use RunWithPassword when a
// deterministic per-test password is needed.
func Run(db *sql.DB) error {
	return RunWithPassword(db, resolveSeedPassword())
}

// resolveSeedPassword returns the configured seed password, or
// DefaultSeedPassword when SeedPasswordEnvVar is unset.
func resolveSeedPassword() string {
	if p := strings.TrimSpace(os.Getenv(SeedPasswordEnvVar)); p != "" {
		return p
	}

	return DefaultSeedPassword
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
