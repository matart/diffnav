package filenode

import (
	"fmt"
	"image/color"
	"path/filepath"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"
	"github.com/bluekeyes/go-gitdiff/gitdiff"

	"github.com/dlvhdr/diffnav/pkg/config"
	"github.com/dlvhdr/diffnav/pkg/icons"
	"github.com/dlvhdr/diffnav/pkg/utils"
)

// Icon style constants.
const (
	IconsNerdStatus   = "nerd-fonts-status"
	IconsNerdSimple   = "nerd-fonts-simple"
	IconsNerdFiletype = "nerd-fonts-filetype"
	IconsNerdFull     = "nerd-fonts-full"
	IconsUnicode      = "unicode"
	IconsASCII        = "ascii"
)

type FileNode struct {
	File       *gitdiff.File
	Depth      int
	YOffset    int
	Selected   bool
	PanelWidth int
	Cfg        config.Config
	// ReviewedFn, if set, returns (reviewedHunkCount, totalHunkCount) for this
	// file at render time so the badge stays fresh without rebuilding the tree.
	ReviewedFn func(file *gitdiff.File) (int, int)
}

func (f *FileNode) Path() string {
	return GetFileName(f.File)
}

func (f *FileNode) Value() string {
	name := filepath.Base(f.Path())

	// full has a special layout: [status icon] [filename] [file-type icon]
	if f.Cfg.UI.Icons == IconsNerdFull {
		return utils.RemoveReset(f.renderFullLayout(name))
	}

	// All other styles: [icon] [filename] with optional coloring
	return utils.RemoveReset(f.renderStandardLayout(name))
}

// renderStandardLayout renders: [icon colored] [filename]
// Used by status, simple, filetype, unicode, ascii.
func (f *FileNode) renderStandardLayout(name string) string {
	icon := f.getIcon() + " "
	iconWidth := lipgloss.Width(icon) + 1

	stats := ""
	if f.Cfg.UI.ShowDiffStats {
		stats = " " + ViewFileDiffStats(f.File, lipgloss.NewStyle())
	}

	badge := f.reviewedBadge()
	fullyReviewed := f.allHunksReviewed()
	suffix := stats + badge

	nameMaxWidth := f.PanelWidth - f.Depth - iconWidth - lipgloss.Width(suffix)
	truncatedName := utils.TruncateString(name, nameMaxWidth)
	coloredIcon := lipgloss.NewStyle().Foreground(f.StatusColor()).Render(icon)
	if fullyReviewed {
		coloredIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(icon)
	}

	if f.Selected {
		bgStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(f.StatusColor())
		if fullyReviewed {
			bgStyle = bgStyle.Foreground(lipgloss.Color("8"))
		}
		if f.PanelWidth > 0 {
			availableWidth := f.PanelWidth - iconWidth - f.Depth
			if availableWidth > 0 {
				bgStyle = bgStyle.Width(availableWidth)
			}
		}
		return coloredIcon + bgStyle.Render(truncatedName) + suffix
	}

	if fullyReviewed {
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		return coloredIcon + dim.Render(truncatedName) + suffix
	}

	if f.Cfg.UI.ColorFileNames {
		styledName := lipgloss.NewStyle().Foreground(f.StatusColor()).Render(truncatedName)
		return coloredIcon + styledName + suffix
	}

	return coloredIcon + truncatedName + suffix
}

// reviewedBadge renders ` n/m` after the row. Always shown when the file has
// hunks: dim grey for 0 reviewed (a "not started" marker), green when all done.
func (f *FileNode) reviewedBadge() string {
	if f.ReviewedFn == nil {
		return ""
	}
	done, total := f.ReviewedFn(f.File)
	if total == 0 {
		return ""
	}
	color := lipgloss.Color("8")
	if done == total {
		color = lipgloss.Green
	}
	return " " + lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%d/%d", done, total))
}

func (f *FileNode) allHunksReviewed() bool {
	if f.ReviewedFn == nil {
		return false
	}
	done, total := f.ReviewedFn(f.File)
	return total > 0 && done == total
}

// renderFullLayout renders: [status icon colored] [file-type icon colored] [filename]
// All icons colored by git status.
func (f *FileNode) renderFullLayout(name string) string {
	statusIcon := f.getStatusIcon()
	fileIcon := icons.GetIcon(name, false)
	style := lipgloss.NewStyle().Foreground(f.StatusColor())
	if f.allHunksReviewed() {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	}

	stats := ""
	if f.Cfg.UI.ShowDiffStats {
		stats = " " + ViewFileDiffStats(f.File, lipgloss.NewStyle())
	}
	suffix := stats + f.reviewedBadge()

	iconsPrefix := style.Render(statusIcon) + " " + style.Render(fileIcon) + " "
	iconsWidth := lipgloss.Width(statusIcon) + 1 + lipgloss.Width(fileIcon) + 1

	nameMaxWidth := f.PanelWidth - f.Depth - iconsWidth - lipgloss.Width(suffix)
	truncatedName := utils.TruncateString(name, nameMaxWidth)

	if f.Selected {
		bgStyle := style.Bold(true)
		if f.PanelWidth > 0 {
			if w := f.PanelWidth - iconsWidth - f.Depth; w > 0 {
				bgStyle = bgStyle.Width(w)
			}
		}
		return iconsPrefix + bgStyle.Render(truncatedName) + suffix
	}

	if f.Cfg.UI.ColorFileNames {
		return iconsPrefix + style.Render(truncatedName) + suffix
	}
	return iconsPrefix + lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Render(truncatedName) +
		suffix
}

// getIcon returns the left icon based on the icon style.
func (f *FileNode) getIcon() string {
	name := filepath.Base(f.Path())
	switch f.Cfg.UI.Icons {
	case IconsNerdStatus:
		if f.File.IsNew {
			return ""
		} else if f.File.IsDelete {
			return ""
		}
		return ""
	case IconsNerdSimple:
		return ""
	case IconsNerdFiletype:
		return icons.GetIcon(name, false) // File-type specific icon (colored by status)
	case IconsUnicode:
		if f.File.IsNew {
			return "+"
		} else if f.File.IsDelete {
			return "⛌"
		}
		return "●"
	default: // ascii (fallback for unknown values)
		if f.File.IsNew {
			return "+"
		} else if f.File.IsDelete {
			return "x"
		}
		return "*"
	}
}

// getStatusIcon returns the git status indicator icon (used by full layout).
// Uses the same boxed icons as status style.
func (f *FileNode) getStatusIcon() string {
	if f.File.IsNew {
		return "\uf457" //
	} else if f.File.IsDelete {
		return "\ueadf" //
	}
	return "\uf459" //
}

// StatusColor returns the color for this file based on its git status.
func (f *FileNode) StatusColor() color.Color {
	if f.File.IsNew {
		return lipgloss.Green
	} else if f.File.IsDelete {
		return lipgloss.Red
	}
	return lipgloss.Yellow
}

func (f *FileNode) String() string {
	return f.Value()
}

func (f *FileNode) Children() tree.Children {
	return tree.NodeChildren(nil)
}

func (f *FileNode) Hidden() bool {
	return false
}

func (f *FileNode) SetHidden(bool) {}

func (f *FileNode) SetValue(any) {}

func DiffStats(file *gitdiff.File) (int64, int64) {
	if file == nil {
		return 0, 0
	}
	var added int64 = 0
	var deleted int64 = 0
	frags := file.TextFragments
	for _, frag := range frags {
		added += frag.LinesAdded
		deleted += frag.LinesDeleted
	}
	return added, deleted
}

func ViewDiffStats(added, deleted int64, base lipgloss.Style) string {
	addedView := ""
	deletedView := ""

	if added > 0 {
		addedView = base.Foreground(lipgloss.Green).Render(fmt.Sprintf("+%d", added))
	}

	if added > 0 && deleted > 0 {
		addedView += base.Render(" ")
	}

	if deleted > 0 {
		deletedView = base.Foreground(lipgloss.Red).Render(fmt.Sprintf("-%d", deleted))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, addedView, deletedView)
}

func ViewFileDiffStats(file *gitdiff.File, base lipgloss.Style) string {
	added, deleted := DiffStats(file)

	return ViewDiffStats(added, deleted, base)
}

func GetFileName(file *gitdiff.File) string {
	if file.NewName != "" {
		return file.NewName
	}
	return file.OldName
}
