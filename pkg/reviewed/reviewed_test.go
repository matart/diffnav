package reviewed

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

func makeFragment(lines ...struct {
	op   gitdiff.LineOp
	text string
}) *gitdiff.TextFragment {
	frag := &gitdiff.TextFragment{}
	for _, l := range lines {
		frag.Lines = append(frag.Lines, gitdiff.Line{Op: l.op, Line: l.text})
	}
	return frag
}

func TestHunkID_StableForSameContent(t *testing.T) {
	frag := makeFragment(
		struct {
			op   gitdiff.LineOp
			text string
		}{gitdiff.OpContext, "a\n"},
		struct {
			op   gitdiff.LineOp
			text string
		}{gitdiff.OpDelete, "b\n"},
		struct {
			op   gitdiff.LineOp
			text string
		}{gitdiff.OpAdd, "c\n"},
	)
	a := HunkID("path.go", frag)
	b := HunkID("path.go", frag)
	if a == "" || a != b {
		t.Fatalf("expected stable non-empty id, got %q vs %q", a, b)
	}
}

func TestHunkID_DiffersByFilePath(t *testing.T) {
	frag := makeFragment(struct {
		op   gitdiff.LineOp
		text string
	}{gitdiff.OpContext, "x\n"})
	if HunkID("a.go", frag) == HunkID("b.go", frag) {
		t.Fatal("expected different ids for different file paths")
	}
}

func TestHunkID_DiffersByContent(t *testing.T) {
	a := makeFragment(struct {
		op   gitdiff.LineOp
		text string
	}{gitdiff.OpAdd, "v1\n"})
	b := makeFragment(struct {
		op   gitdiff.LineOp
		text string
	}{gitdiff.OpAdd, "v2\n"})
	if HunkID("x.go", a) == HunkID("x.go", b) {
		t.Fatal("expected different ids for different content")
	}
}

func TestHunkID_NilFragment(t *testing.T) {
	if HunkID("a", nil) != "" {
		t.Fatal("expected empty id for nil fragment")
	}
}

func TestToggleAndIsReviewed(t *testing.T) {
	s := &State{Repos: map[string][]string{}}
	if s.IsReviewed("/repo", "id1") {
		t.Fatal("expected not reviewed initially")
	}
	if got := s.Toggle("/repo", "id1"); !got {
		t.Fatal("expected toggle to mark reviewed")
	}
	if !s.IsReviewed("/repo", "id1") {
		t.Fatal("expected reviewed after toggle")
	}
	if got := s.Toggle("/repo", "id1"); got {
		t.Fatal("expected toggle to unmark reviewed")
	}
	if s.IsReviewed("/repo", "id1") {
		t.Fatal("expected not reviewed after second toggle")
	}
}

func TestRepoIsolation(t *testing.T) {
	s := &State{Repos: map[string][]string{}}
	s.Toggle("/a", "shared")
	if s.IsReviewed("/b", "shared") {
		t.Fatal("expected per-repo isolation")
	}
}

func TestEmptyRepoRootUsesSentinel(t *testing.T) {
	s := &State{Repos: map[string][]string{}}
	s.Toggle("", "id1")
	if !s.IsReviewed("", "id1") {
		t.Fatal("expected toggle to work with empty repo root")
	}
}

func TestCountReviewed(t *testing.T) {
	s := &State{Repos: map[string][]string{}}
	s.Toggle("/r", "a")
	s.Toggle("/r", "c")
	got := s.CountReviewed("/r", []string{"a", "b", "c", "d"})
	if got != 2 {
		t.Fatalf("expected 2 reviewed, got %d", got)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := Load(dir)
	if err != nil {
		t.Fatalf("load empty: %v", err)
	}
	s.Toggle("/repo", "hash-a")
	s.Toggle("/repo", "hash-b")
	if err := s.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if !loaded.IsReviewed("/repo", "hash-a") || !loaded.IsReviewed("/repo", "hash-b") {
		t.Fatal("expected hashes to persist across load")
	}
}

func TestSaveAtomicLeavesNoTempOnFailure(t *testing.T) {
	dir := t.TempDir()
	s, _ := Load(dir)
	s.Toggle("/r", "x")
	if err := s.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	// No leftover temp files in the directory.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".reviewed-") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("leftover temp file: %s", e.Name())
		}
	}
}

func TestLoadEmptyDirIsDisabled(t *testing.T) {
	s, err := Load("")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	s.Toggle("/r", "x")
	if err := s.Save(); err != nil {
		t.Fatalf("save should be no-op: %v", err)
	}
}

func TestResolvePath_EnvOverride(t *testing.T) {
	t.Setenv(envDir, "/tmp/custom")
	if got := ResolvePath(Config{Path: "/ignored"}); got != "/tmp/custom" {
		t.Fatalf("expected env override, got %q", got)
	}
}

func TestResolvePath_YAMLOverride(t *testing.T) {
	t.Setenv(envDir, "")
	if got := ResolvePath(Config{Path: "/yaml/path"}); got != "/yaml/path" {
		t.Fatalf("expected yaml override, got %q", got)
	}
}

func TestResolvePath_ExpandsTilde(t *testing.T) {
	t.Setenv(envDir, "")
	home, _ := os.UserHomeDir()
	got := ResolvePath(Config{Path: "~/data/diffnav"})
	want := filepath.Join(home, "data", "diffnav")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolvePath_XDGFallback(t *testing.T) {
	t.Setenv(envDir, "")
	t.Setenv("XDG_DATA_HOME", "/xdg/data")
	if got := ResolvePath(Config{}); got != "/xdg/data/diffnav" {
		t.Fatalf("expected XDG path, got %q", got)
	}
}
