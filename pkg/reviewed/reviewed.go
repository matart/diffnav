// Package reviewed persists per-hunk "reviewed" markers across diffnav sessions.
package reviewed

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

// Config controls where review state is persisted.
type Config struct {
	Path string `yaml:"path"`
}

const (
	envDir   = "DIFFNAV_STORAGE_DIR"
	fileName = "reviewed.json"
	noRepo   = "$norepo"
)

// State is the persisted set of reviewed hunk IDs, keyed by repo root.
type State struct {
	Repos map[string][]string `json:"repos"`

	path string // empty if persistence is disabled
}

// HunkID computes a stable identifier for a hunk: sha256 of file path plus the
// canonical content of the fragment (ops + line text, in order). Renames or
// content edits invalidate the marker.
func HunkID(filePath string, frag *gitdiff.TextFragment) string {
	if frag == nil {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(filePath))
	h.Write([]byte{'\n'})
	for _, line := range frag.Lines {
		h.Write([]byte{byte(opGlyph(line.Op))})
		h.Write([]byte(line.Line))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func opGlyph(op gitdiff.LineOp) rune {
	switch op {
	case gitdiff.OpAdd:
		return '+'
	case gitdiff.OpDelete:
		return '-'
	default:
		return ' '
	}
}

// DefaultPath returns the default storage directory following XDG conventions.
func DefaultPath() string {
	if p := os.Getenv("XDG_DATA_HOME"); p != "" {
		return filepath.Join(p, "diffnav")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if runtime.GOOS == "windows" {
			if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
				return filepath.Join(appData, "diffnav")
			}
		}
		return filepath.Join(home, ".local", "share", "diffnav")
	}
	return ""
}

// ResolvePath applies env override → yaml override → default, in priority order.
func ResolvePath(cfg Config) string {
	if dir := os.Getenv(envDir); dir != "" {
		return dir
	}
	if cfg.Path != "" {
		return expandHome(cfg.Path)
	}
	return DefaultPath()
}

func expandHome(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~"))
}

// Load reads the reviewed state. Returns an empty state pinned to dir/reviewed.json
// if the file doesn't yet exist. If dir is empty, returns a disabled state that
// silently ignores Save calls.
func Load(dir string) (*State, error) {
	if dir == "" {
		return &State{Repos: map[string][]string{}}, nil
	}
	path := filepath.Join(dir, fileName)
	state := &State{Repos: map[string][]string{}, path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, err
	}
	if err := json.Unmarshal(data, state); err != nil {
		return state, err
	}
	if state.Repos == nil {
		state.Repos = map[string][]string{}
	}
	return state, nil
}

// Save atomically writes the state to disk. No-op if persistence is disabled.
func (s *State) Save() error {
	if s == nil || s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".reviewed-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, s.path)
}

// IsReviewed reports whether the hunk is marked reviewed in this repo.
func (s *State) IsReviewed(repoRoot, hunkID string) bool {
	if s == nil || hunkID == "" {
		return false
	}
	key := repoKey(repoRoot)
	return slices.Contains(s.Repos[key], hunkID)
}

// Toggle flips the reviewed state for a hunk and returns the new state.
func (s *State) Toggle(repoRoot, hunkID string) bool {
	if s == nil || hunkID == "" {
		return false
	}
	key := repoKey(repoRoot)
	existing := s.Repos[key]
	if i := slices.Index(existing, hunkID); i >= 0 {
		s.Repos[key] = slices.Delete(existing, i, i+1)
		return false
	}
	s.Repos[key] = append(existing, hunkID)
	return true
}

// CountReviewed counts how many of the given hunk IDs are marked reviewed.
func (s *State) CountReviewed(repoRoot string, hunkIDs []string) int {
	if s == nil || len(hunkIDs) == 0 {
		return 0
	}
	key := repoKey(repoRoot)
	have := s.Repos[key]
	if len(have) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(have))
	for _, id := range have {
		set[id] = struct{}{}
	}
	n := 0
	for _, id := range hunkIDs {
		if _, ok := set[id]; ok {
			n++
		}
	}
	return n
}

func repoKey(repoRoot string) string {
	if repoRoot == "" {
		return noRepo
	}
	return repoRoot
}
