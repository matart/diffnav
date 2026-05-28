# Ideas

Roadmap for evolving `diffnav` from a stateless diff pager into an AI-era code review
workspace.

The unifying goal: make local PR review fast enough that the loop
**write → review locally with Claude → fix → push** beats the GitHub round-trip.

## Foundation: statefulness

Today `diffnav` is a stateless pager. The features below need a place to put comments,
reviewed-markers, cached AI reviews, and fetched GitHub data. A configurable storage
path is the wedge that unlocks the rest.

### Storage config

Add a `storage.path` config option. Default to XDG (`$XDG_DATA_HOME/diffnav`, falling
back to `~/.local/share/diffnav`); allow override:

```yaml
storage:
  path: ~/.local/share/diffnav
```

### Layout

```
<storage.path>/
  state.db                                # sqlite — reviewed markers, comments
  cache/
    reviews/<repo-hash>/<diff-hash>.json    # Claude review output
    gh-comments/<repo-hash>/<pr-id>.json    # fetched GH comments
```

SQLite for anything we query by lookup (reviewed status, comments by file/line).
JSON-per-diff for blobby outputs (review reports). Key everything by
`<repo-root-hash, base-sha..head-sha>` so caches invalidate naturally when the diff
changes.

## Features, ranked by leverage vs effort

| # | Feature                                                                                                    | Leverage    | Effort   | State?           | Why this rank                                                                                                                  |
| - | ---------------------------------------------------------------------------------------------------------- | ----------- | -------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| 1 | **Reviewed markers per hunk** — toggle a hunk as "seen", persist it, dim on reopen ✅ shipped               | High        | Low      | Yes (small)      | Mechanical, immediate win on big PRs. Validates the whole state layer with the simplest possible schema.                       |
| 2 | **Smart ordering / collapse** — push generated files, lockfiles, snapshots to bottom or collapse by default | High        | Low–Med  | No (config only) | One-time UX upgrade. For a 40-file patch with mostly noise, halves scroll time.                                                |
| 3 | **Explain / critique this hunk** — keybind on a hunk, Claude responds in a side pane                       | High        | Low–Med  | Optional cache   | Cheap to build, AI-native, scoped to one hunk so prompts stay tight. Lays the Claude-invocation plumbing for #4.               |
| 4 | **Pre-push Claude review** — run a full-diff review locally, render comments inline                        | Very High   | Med–High | Yes (cache)      | Biggest single AI-era win: collapses push → CI → human → fix → push into one local step. Reuses #3's plumbing and #5's pinning.|
| 5 | **GitHub PR comments inline** — fetch via `gh`, render at the right lines                                  | High (PRs)  | Med      | Cache            | Hard part is mapping GH's commit-SHA-relative positions to local diff lines. Same line-pinning code as #4.                     |
| 6 | **Local line comments** — drafts that can optionally sync to GitHub                                        | Med–High    | Med      | Yes              | Useful as a "first pass" workspace. Lower priority because #4 covers AI feedback and #5 covers GH feedback.                    |
| 7 | **AI conversation thread per hunk** — multi-turn chat anchored to a hunk                                   | Med–High    | High     | Yes (heavy)      | Overlaps #3 heavily. Save for after the others land — by then we'll know if the single-shot version was enough.                |

## Suggested build order

**1 → 2 → 3 → 4 → 5 → 6 → 7**.

- 1–3 are independent and small. Each is independently shippable.
- 4 and 5 share infrastructure with 3 (Claude invocation, comment-to-line pinning), so
  they get cheaper once 3 lands.
- 6 builds on 5's GitHub plumbing.
- 7 is a possible follow-up to 3 if the single-shot UX feels insufficient.

## Notes

- Keep `diffnav` usable as a pure pager — every stateful feature must degrade
  cleanly when storage is unwritable or unconfigured.
- Cache invalidation must key on the diff content, not just file paths; same file
  path across two different PRs should not share review state.
- Claude integration should pick a model deliberately rather than defaulting to
  whatever's latest, so review style stays stable across runs.
