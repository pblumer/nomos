package update

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.2.0", "0.1.0", true},
		{"0.2.0", "v0.1.0", true},
		{"v1.0.0", "0.9.9", true},
		{"v0.1.1", "0.1.0", true},
		{"v0.1.0", "0.1.0", false},
		{"v0.1.0", "0.2.0", false},
		{"v0.2.0", "dev", false},    // non-semver current never updatable
		{"nightly", "0.1.0", false}, // non-semver latest never updatable
		{"v1.2.3-rc1", "1.2.3", false},
	}
	for _, c := range cases {
		if got := isNewer(c.latest, c.current); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestEnabledDefaultsOff(t *testing.T) {
	t.Setenv("NOMOS_UPDATE_CHECK", "")
	if Enabled() {
		t.Fatal("update check should be opt-in (off by default)")
	}
	t.Setenv("NOMOS_UPDATE_CHECK", "1")
	if !Enabled() {
		t.Fatal("NOMOS_UPDATE_CHECK=1 should enable the check")
	}
}

func TestStatusDisabledReportsCurrentVersion(t *testing.T) {
	t.Setenv("NOMOS_UPDATE_CHECK", "")
	s := NewChecker().Status()
	if s.Enabled {
		t.Error("expected disabled status")
	}
	if s.Current == "" {
		t.Error("current version should always be reported")
	}
	if s.UpdateAvailable {
		t.Error("disabled check must never report an available update")
	}
}
