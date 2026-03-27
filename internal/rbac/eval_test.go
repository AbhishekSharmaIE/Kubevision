package rbac

import "testing"

func TestRank(t *testing.T) {
	if Rank(PermRead) >= Rank(PermWrite) {
		t.Fatal("read < write")
	}
	if Rank(PermWrite) >= Rank(PermAdmin) {
		t.Fatal("write < admin")
	}
	if Rank("bogus") != 0 {
		t.Fatal("unknown permission rank 0")
	}
}

func TestParsePermission(t *testing.T) {
	if err := ParsePermission(PermRead); err != nil {
		t.Fatal(err)
	}
	if err := ParsePermission("invalid"); err == nil {
		t.Fatal("expected error")
	}
}
