package repositories

import (
	"path/filepath"
	"strings"
)

var fileAliasRootMarkers = []string{
	"/dashboard_files/",
	"/dashboard/files/",
	"/data/files/",
}

// normalizeFileAliasPath lowercases, uses forward slashes, and strips a trailing .bin.
func normalizeFileAliasPath(p string) string {
	n := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(p), `\`, `/`))
	n = strings.TrimSuffix(n, ".bin")
	// Collapse duplicate slashes without using filepath (keeps forward slashes on Windows).
	for strings.Contains(n, "//") {
		n = strings.ReplaceAll(n, "//", "/")
	}
	return strings.TrimPrefix(n, "./")
}

// logicalFileRelPath extracts a logical relative path under the files root when possible.
func logicalFileRelPath(p string) string {
	n := normalizeFileAliasPath(p)
	if n == "" || n == "." {
		return ""
	}

	for _, marker := range fileAliasRootMarkers {
		if idx := strings.LastIndex(n, marker); idx >= 0 {
			return strings.TrimPrefix(n[idx+len(marker):], "/")
		}
	}

	if idx := strings.LastIndex(n, "/files/"); idx >= 0 {
		return strings.TrimPrefix(n[idx+len("/files/"):], "/")
	}

	// Already relative (no drive / leading slash).
	if !strings.HasPrefix(n, "/") && !hasWindowsDrivePrefix(n) {
		return strings.TrimPrefix(n, "/")
	}

	// Absolute but under an unknown root: keep path after first segment that looks like a files top dir.
	parts := strings.Split(strings.TrimPrefix(n, "/"), "/")
	if len(parts) > 0 && hasWindowsDrivePrefix(parts[0]+"/") {
		// "c:/users/..." → drop "c:"
		parts = parts[1:]
	}
	for i, part := range parts {
		if isFilesTopDir(part) {
			return strings.Join(parts[i:], "/")
		}
	}

	return ""
}

func hasWindowsDrivePrefix(n string) bool {
	return len(n) >= 2 && n[1] == ':' && n[0] >= 'a' && n[0] <= 'z'
}

func isFilesTopDir(name string) bool {
	switch name {
	case "tasks", "problems", "questions", "actions", "definitions",
		"knowledge-bits", "knowledge-nodes", "stories", "epics",
		"scheduled-tasks", "states":
		return true
	default:
		return false
	}
}

// fileAliasPathMatches reports whether a stored alias path refers to the same logical file
// as target (absolute or relative). Handles Windows/Unix separators, case, and optional .bin.
func fileAliasPathMatches(stored, target string) bool {
	storedNorm := normalizeFileAliasPath(stored)
	targetNorm := normalizeFileAliasPath(target)
	if storedNorm == "" || targetNorm == "" {
		return false
	}
	if storedNorm == targetNorm {
		return true
	}

	storedRel := logicalFileRelPath(stored)
	targetRel := logicalFileRelPath(target)
	if storedRel != "" && targetRel != "" && storedRel == targetRel {
		return true
	}

	// Target is a logical relative path; stored is absolute (or vice versa).
	if targetRel == "" && !filepath.IsAbs(target) && !hasWindowsDrivePrefix(targetNorm) {
		targetRel = strings.TrimPrefix(targetNorm, "/")
	}
	if storedRel == "" && !filepath.IsAbs(stored) && !hasWindowsDrivePrefix(storedNorm) {
		storedRel = strings.TrimPrefix(storedNorm, "/")
	}

	if storedRel != "" && targetRel != "" && storedRel == targetRel {
		return true
	}
	if targetRel != "" && (storedNorm == targetRel || strings.HasSuffix(storedNorm, "/"+targetRel)) {
		return true
	}
	if storedRel != "" && (targetNorm == storedRel || strings.HasSuffix(targetNorm, "/"+storedRel)) {
		return true
	}
	return false
}
