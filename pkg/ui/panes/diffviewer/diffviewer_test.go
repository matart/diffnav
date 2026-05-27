package diffviewer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderPreamble_Empty(t *testing.T) {
	if got := renderPreamble(""); got != "" {
		t.Fatalf("expected empty string for empty preamble, got %q", got)
	}
	if got := renderPreamble("   \n  \n  "); got != "" {
		t.Fatalf("expected empty string for whitespace-only preamble, got %q", got)
	}
}

func TestRenderPreamble_GitShow(t *testing.T) {
	preamble := `commit abc123def456
Author: Jane Doe <jane@example.com>
Date:   Mon Jan 1 00:00:00 2026 +0000

    feat: add new feature

    This is the body of the commit message.`

	got := renderPreamble(preamble)
	plain := ansi.Strip(got)

	// All original content lines should be preserved in the output.
	for _, want := range []string{
		"commit abc123def456",
		"Author: Jane Doe <jane@example.com>",
		"Date:   Mon Jan 1 00:00:00 2026 +0000",
		"feat: add new feature",
		"This is the body of the commit message.",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, plain)
		}
	}
}

func TestInjectPHPOpenTag_NonPHPFileUnchanged(t *testing.T) {
	body := "@@ -10,2 +10,2 @@\n a\n-b\n+c\n"
	if got := injectPHPOpenTag("foo.go", body); got != body {
		t.Fatalf("expected non-PHP file unchanged, got %q", got)
	}
}

func TestInjectPHPOpenTag_SkipsHunksThatAlreadyHaveTag(t *testing.T) {
	body := "@@ -1,3 +1,3 @@\n <?php\n-foo\n+bar\n"
	if got := injectPHPOpenTag("foo.php", body); got != body {
		t.Fatalf("expected unchanged when <?php already present, got %q", got)
	}
}

func TestInjectPHPOpenTag_SkipsHunksThatAlreadyHaveShortEcho(t *testing.T) {
	body := "@@ -1,3 +1,3 @@\n <?= $foo ?>\n-a\n+b\n"
	if got := injectPHPOpenTag("foo.php", body); got != body {
		t.Fatalf("expected unchanged when <?= already present, got %q", got)
	}
}

func TestInjectPHPOpenTag_InjectsAndShiftsLineNumbers(t *testing.T) {
	body := "@@ -10,3 +10,3 @@ class Foo\n class Foo {\n-    bar\n+    baz\n }\n"
	got := injectPHPOpenTag("src/Foo.php", body)
	want := "@@ -9,4 +9,4 @@ class Foo\n <?php /*diffnav_phpfix*/\n class Foo {\n-    bar\n+    baz\n }\n"
	if got != want {
		t.Fatalf("unexpected output:\n got:  %q\n want: %q", got, want)
	}
}

func TestInjectPHPOpenTag_SkipsWhenCannotShift(t *testing.T) {
	// If the first hunk starts at line 1, shifting would land on 0; skip injection
	// so stripping the synthetic line later doesn't desync real line numbers.
	body := "@@ -1,2 +1,2 @@\n-a\n+b\n"
	if got := injectPHPOpenTag("foo.php", body); got != body {
		t.Fatalf("expected unchanged when first hunk starts at line 1, got %q", got)
	}
}

func TestInjectPHPOpenTag_TouchesEveryHunk(t *testing.T) {
	body := "@@ -10,1 +10,1 @@\n a\n@@ -50,1 +50,1 @@\n b\n"
	got := injectPHPOpenTag("foo.php", body)
	want := "@@ -9,2 +9,2 @@\n <?php /*diffnav_phpfix*/\n a\n@@ -49,2 +49,2 @@\n <?php /*diffnav_phpfix*/\n b\n"
	if got != want {
		t.Fatalf("unexpected output:\n got:  %q\n want: %q", got, want)
	}
}

func TestInjectPHPOpenTag_MixedHunks(t *testing.T) {
	// First hunk has <?php (near top of file) → skipped. Second hunk further
	// down has no <?php → injected. This is the case the user reported.
	body := "@@ -1,3 +1,3 @@\n <?php\n-a\n+b\n@@ -50,1 +50,1 @@\n c\n"
	got := injectPHPOpenTag("foo.php", body)
	want := "@@ -1,3 +1,3 @@\n <?php\n-a\n+b\n@@ -49,2 +49,2 @@\n <?php /*diffnav_phpfix*/\n c\n"
	if got != want {
		t.Fatalf("unexpected output:\n got:  %q\n want: %q", got, want)
	}
}

func TestInjectPHPOpenTag_OmittedHunkCount(t *testing.T) {
	body := "@@ -10 +10 @@\n a\n"
	got := injectPHPOpenTag("foo.php", body)
	want := "@@ -9,2 +9,2 @@\n <?php /*diffnav_phpfix*/\n a\n"
	if got != want {
		t.Fatalf("unexpected output:\n got:  %q\n want: %q", got, want)
	}
}

func TestStripPHPSyntheticLines_RemovesMarkedLines(t *testing.T) {
	in := "header\n<?php /*diffnav_phpfix*/\nreal line\nanother\n"
	got := stripPHPSyntheticLines(in)
	want := "header\nreal line\nanother\n"
	if got != want {
		t.Fatalf("unexpected output:\n got:  %q\n want: %q", got, want)
	}
}

func TestStripPHPSyntheticLines_StripsThroughANSI(t *testing.T) {
	// Simulate delta's ANSI-coloured output around our sentinel.
	in := "\x1b[34mheader\x1b[0m\n\x1b[38;5;240m<?php \x1b[0m\x1b[38;5;245m/*diffnav_phpfix*/\x1b[0m\nreal\n"
	got := stripPHPSyntheticLines(in)
	want := "\x1b[34mheader\x1b[0m\nreal\n"
	if got != want {
		t.Fatalf("unexpected output:\n got:  %q\n want: %q", got, want)
	}
}

func TestStripPHPSyntheticLines_NoMarkerNoOp(t *testing.T) {
	in := "header\nreal\nanother\n"
	if got := stripPHPSyntheticLines(in); got != in {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

func TestInjectPHPOpenTag_PHPExtensionVariants(t *testing.T) {
	body := "@@ -10,1 +10,1 @@\n a\n"
	for _, ext := range []string{".php", ".PHP", ".phtml", ".php3", ".php7", ".phps", ".phpt"} {
		if injectPHPOpenTag("foo"+ext, body) == body {
			t.Errorf("expected injection for extension %q, got no change", ext)
		}
	}
}

func TestRenderPreamble_MergeCommit(t *testing.T) {
	preamble := `commit abc123def456
Merge: aaa111 bbb222
Author: Jane Doe <jane@example.com>
Date:   Mon Jan 1 00:00:00 2026 +0000

    Merge branch 'feature' into main`

	got := renderPreamble(preamble)
	plain := ansi.Strip(got)

	for _, want := range []string{
		"Merge: aaa111 bbb222",
		"Merge branch 'feature' into main",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, plain)
		}
	}
}
