package def

import (
	"strings"
	"testing"
)

// Ikemen GO reads [Info] mugenversion and ikemenversion (common.go:931-970
// parses them); they decide several MUGEN-versus-Ikemen behaviours.
func TestParse_InfoVersions(t *testing.T) {
	info, err := Parse(strings.NewReader("[Info]\nname = KFM\nmugenversion = 1.1\nikemenversion = 1.0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if info.MugenVersion != "1.1" || info.IkemenVersion != "1.0" {
		t.Errorf("versions = %q %q, want 1.1 and 1.0", info.MugenVersion, info.IkemenVersion)
	}
	info, err = Parse(strings.NewReader("[Info]\nname = Old\n"))
	if err != nil || info.MugenVersion != "" || info.IkemenVersion != "" {
		t.Errorf("no versions: %q %q (%v), want empty", info.MugenVersion, info.IkemenVersion, err)
	}
}
