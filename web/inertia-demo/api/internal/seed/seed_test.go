package seed

import (
	"testing"

	"alloy.dev/inertia-demo/internal/database"
)

func TestRunSeedsDatabase(t *testing.T) {
	t.Parallel()

	db, err := database.Open(":memory:")

	if err != nil {
		t.Fatalf("Open(:memory:) error = %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := Run(db); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if got, err := database.GetCounter(db, "priority_escalations"); err != nil || got != 18 {
		t.Fatalf("GetCounter(priority_escalations) = %d, %v, want 18, nil", got, err)
	}

	if got, err := database.InviteCount(db); err != nil || got != 20 {
		t.Fatalf("InviteCount() = %d, %v, want 20, nil", got, err)
	}

	if got, err := database.UploadCount(db); err != nil || got != 15 {
		t.Fatalf("UploadCount() = %d, %v, want 15, nil", got, err)
	}

	if got, err := database.ApprovalCount(db); err != nil || got != 15 {
		t.Fatalf("ApprovalCount() = %d, %v, want 15, nil", got, err)
	}

	users := []struct {
		email string
		name  string
	}{
		{email: "test@example.com", name: "Demo User"},
		{email: "alice@example.com", name: "Alice Manager"},
	}

	for _, want := range users {
		user, err := database.FindUserByEmail(db, want.email)

		if err != nil {
			t.Fatalf("FindUserByEmail(%q) error = %v", want.email, err)
		}

		if user == nil || user.Name != want.name {
			t.Fatalf("FindUserByEmail(%q) = %#v, want name %q", want.email, user, want.name)
		}
	}

	contacts, err := database.ListContacts(db, "", false)

	if err != nil {
		t.Fatalf("ListContacts() error = %v", err)
	}

	if len(contacts) == 0 {
		t.Fatal("ListContacts() returned no contacts")
	}

	organizations, err := database.ListOrganizations(db, "")

	if err != nil {
		t.Fatalf("ListOrganizations() error = %v", err)
	}

	if len(organizations) == 0 {
		t.Fatal("ListOrganizations() returned no organizations")
	}

	notes, err := database.ListRecentNotes(db, 5)

	if err != nil {
		t.Fatalf("ListRecentNotes() error = %v", err)
	}

	if len(notes) == 0 {
		t.Fatal("ListRecentNotes() returned no notes")
	}

	if invites, err := database.ListInvites(db); err != nil || len(invites) != 20 {
		t.Fatalf("ListInvites() len = %d, %v, want 20, nil", len(invites), err)
	}

	if uploads, err := database.ListUploads(db); err != nil || len(uploads) != 15 {
		t.Fatalf("ListUploads() len = %d, %v, want 15, nil", len(uploads), err)
	}

	if approvals, err := database.ListApprovals(db); err != nil || len(approvals) != 15 {
		t.Fatalf("ListApprovals() len = %d, %v, want 15, nil", len(approvals), err)
	}
}

func TestRunTruncatesExistingData(t *testing.T) {
	t.Parallel()

	db, err := database.Open(":memory:")

	if err != nil {
		t.Fatalf("Open(:memory:) error = %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if _, err := database.CreateUser(db, "Extra User", "extra@example.com", "password", nil); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if err := database.CreateInvite(db, "legacy", "Legacy User", "legacy@example.com", "Owner", "Queued"); err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	if err := Run(db); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	user, err := database.FindUserByEmail(db, "extra@example.com")

	if err != nil {
		t.Fatalf("FindUserByEmail(extra@example.com) error = %v", err)
	}

	if user != nil {
		t.Fatalf("FindUserByEmail(extra@example.com) = %#v, want nil after truncate", user)
	}

	if invites, err := database.ListInvites(db); err != nil || len(invites) != 20 {
		t.Fatalf("ListInvites() len = %d, %v, want 20, nil after reseed", len(invites), err)
	}
}
