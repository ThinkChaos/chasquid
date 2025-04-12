package courier

import (
	"bytes"
	"os"
	"testing"
	"time"

	"blitiri.com.ar/go/chasquid/internal/testlib"
)

func TestMDA(t *testing.T) {
	dir := testlib.MustTempDir(t)
	defer testlib.RemoveIfOk(t, dir)

	p := MDA{
		Binary:  "tee",
		Args:    []string{dir + "/%to_user%"},
		Timeout: 1 * time.Minute,
	}

	err, _ := p.Deliver("from@x", "to@local", []byte("data"))
	if err != nil {
		t.Fatalf("Deliver: %v", err)
	}

	data, err := os.ReadFile(dir + "/to")
	if err != nil || !bytes.Equal(data, []byte("data")) {
		t.Errorf("Invalid data: %q - %v", string(data), err)
	}
}

func TestMDATimeout(t *testing.T) {
	p := MDA{"/bin/sleep", []string{"1"}, 100 * time.Millisecond}

	err, permanent := p.Deliver("from", "to@local", []byte("data"))
	if err != errTimeout {
		t.Errorf("Unexpected error: %v", err)
	}
	if permanent {
		t.Errorf("expected transient, got permanent")
	}
}

func TestMDABadCommandLine(t *testing.T) {
	// Non-existent binary.
	p := MDA{"thisdoesnotexist", nil, 1 * time.Minute}
	err, permanent := p.Deliver("from", "to", []byte("data"))
	if err == nil {
		t.Errorf("unexpected success for non-existent binary")
	}
	if !permanent {
		t.Errorf("expected permanent, got transient")
	}

	// Incorrect arguments.
	p = MDA{"cat", []string{"--fail_unknown_option"}, 1 * time.Minute}
	err, _ = p.Deliver("from", "to", []byte("data"))
	if err == nil {
		t.Errorf("unexpected success for incorrect arguments")
	}
}

// Test that local delivery failures are considered permanent or not
// according to the exit code.
func TestExitCode(t *testing.T) {
	// TODO: This can happen when building under unusual circumstances, such
	// as Debian package building. Are they reasonable enough for us to keep
	// this?
	if _, err := os.Stat("../../test/util/exitcode"); os.IsNotExist(err) {
		t.Skipf("util/exitcode not found, running from outside repo?")
	}

	cases := []struct {
		cmd             string
		args            []string
		expectPermanent bool
	}{
		{"does/not/exist", nil, true},
		{"../../test/util/exitcode", []string{"1"}, true},
		{"../../test/util/exitcode", []string{"75"}, false},
	}
	for _, c := range cases {
		p := &MDA{c.cmd, c.args, 5 * time.Second}
		err, permanent := p.Deliver("from", "to", []byte("data"))
		if err == nil {
			t.Errorf("%q: pipe delivery worked, expected failure", c.cmd)
		}
		if c.expectPermanent != permanent {
			t.Errorf("%q: permanent expected=%v, got=%v",
				c.cmd, c.expectPermanent, permanent)
		}
	}
}

func TestSanitize(t *testing.T) {
	cases := []struct{ v, expected string }{
		// These are the same.
		{"thisisfine", "thisisfine"},
		{"ñaca", "ñaca"},
		{"123-456_789", "123-456_789"},
		{"123+456~789", "123+456~789"},

		// These have problematic characters that get dropped.
		{"with spaces", "withspaces"},
		{"with/slash", "withslash"},
		{"quote';andsemicolon", "quoteandsemicolon"},
		{"a;b", "ab"},
		{`"test"`, "test"},

		// Interesting cases taken from
		// http://www.user.uni-hannover.de/nhtcapri/bidirectional-text.html
		// We allow them, they're the same on both sides.
		{"١٩٩٩–١٢–٣١", "١٩٩٩–١٢–٣١"},

		//lint:ignore ST1018 The use of a literal U+200C is intentional.
		{"موزه‌ها", "موزه\u200cها"},
	}
	for _, c := range cases {
		out := sanitizeForMDA(c.v)
		if out != c.expected {
			t.Errorf("%q: expected %q, got %q", c.v, c.expected, out)
		}
	}
}

func TestForward(t *testing.T) {
	p := MDA{"thisdoesnotexist", nil, 1 * time.Minute}
	err, permanent := p.Forward(
		"from", "to", []byte("data"), []string{"server"})
	if err != errForwardNotSupported {
		t.Errorf("unexpected error: %v", err)
	}
	if !permanent {
		t.Errorf("expected permanent, got transient")
	}
}
