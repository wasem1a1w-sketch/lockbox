package vault

import (
	"testing"
	"time"
)

func TestEnsureIDsAssignsAndNormalizes(t *testing.T) {
	v := Vault{
		{Account: "legacy-a", Password: "x", SavedAt: time.Now(), SortOrder: 0},
		{Account: "legacy-b", Password: "y", SavedAt: time.Now(), SortOrder: 0},
		{ID: "existing-id", Account: "modern", Password: "z", SavedAt: time.Now(), SortOrder: 3},
	}

	changed := v.EnsureIDs()
	if changed == 0 {
		t.Fatal("changed=0 want >0")
	}
	if v[0].ID == "" || v[1].ID == "" || v[0].ID == v[1].ID {
		t.Errorf("IDs not uniquely assigned: %q %q", v[0].ID, v[1].ID)
	}
	if v[2].ID != "existing-id" {
		t.Errorf("existing ID clobbered: %q", v[2].ID)
	}
	if v[0].SortOrder != 1 || v[1].SortOrder != 2 || v[2].SortOrder != 3 {
		t.Errorf("SortOrder not normalized: %d %d %d", v[0].SortOrder, v[1].SortOrder, v[2].SortOrder)
	}

	// Idempotent.
	if again := v.EnsureIDs(); again != 0 {
		t.Errorf("second run changed=%d want 0", again)
	}
}

func TestEditFields(t *testing.T) {
	v := Vault{{ID: "id-1", Account: "old", Username: "old", Password: "old", SortOrder: 1}}
	if err := v.EditFields("id-1", "acc", "user", "pass"); err != nil {
		t.Fatal(err)
	}
	c := v[0]
	if c.Account != "acc" || c.Username != "user" || c.Password != "pass" {
		t.Errorf("fields not updated: %+v", c)
	}
	if c.SavedAt.IsZero() {
		t.Error("SavedAt not touched")
	}
	if err := v.EditFields("missing", "a", "u", "p"); err == nil {
		t.Error("want error for missing id")
	}
}

func TestSortByOrder(t *testing.T) {
	v := Vault{
		{ID: "c", SortOrder: 3},
		{ID: "a", SortOrder: 1},
		{ID: "b", SortOrder: 2},
	}
	v.SortByOrder()
	if v[0].ID != "a" || v[1].ID != "b" || v[2].ID != "c" {
		t.Errorf("order wrong: %s %s %s", v[0].ID, v[1].ID, v[2].ID)
	}
}
