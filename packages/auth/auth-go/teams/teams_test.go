package teams_test

import (
	"context"
	"testing"

	"github.com/oullin/alloy/auth/teams"
)

func TestServiceCreatesListsAndSwitchesTeams(t *testing.T) {
	service := teams.NewService(teams.NewMemoryRepository(), nil)

	team, err := service.Create(context.Background(), "1", "Core")

	if err != nil {
		t.Fatal(err)
	}

	list, err := service.List(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 1 || list[0].ID != team.ID {
		t.Fatalf("teams = %#v", list)
	}

	current, err := service.Current(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if current.ID != team.ID {
		t.Fatalf("current = %#v", current)
	}

	if err := service.SwitchCurrent(context.Background(), "1", team.ID); err != nil {
		t.Fatal(err)
	}
}

func TestServiceAuthorizesMemberManagementByRole(t *testing.T) {
	repo := teams.NewMemoryRepository()
	service := teams.NewService(repo, []teams.Role{{
		Name:        "admin",
		Permissions: []string{"members:create", "members:update"},
	}})

	team, err := service.Create(context.Background(), "owner", "Core")

	if err != nil {
		t.Fatal(err)
	}

	if err := service.AddMember(context.Background(), "owner", team.ID, "admin", "admin"); err != nil {
		t.Fatal(err)
	}

	if err := service.AddMember(context.Background(), "admin", team.ID, "member", "member"); err != nil {
		t.Fatal(err)
	}

	if err := service.RemoveMember(context.Background(), "admin", team.ID, "member"); err != teams.ErrForbidden {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}

	if err := service.UpdateRole(context.Background(), "admin", team.ID, "member", "admin"); err != nil {
		t.Fatal(err)
	}
}
