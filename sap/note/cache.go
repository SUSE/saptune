package note

import "github.com/SUSE/saptune/txtparser"

// parsedNoteCache is a command-scoped cache for INI files to avoid re-reading from disk.
var parsedNoteCache = make(map[string]*txtparser.INIFile)

// reportedMissingDropIns is a command-scoped cache to de-duplicate missing drop-in error messages.
var reportedMissingDropIns = make(map[string]bool)

// ResetCaches clears all package-level caches. It should be called at the beginning
// of any user-facing saptune command to ensure a clean state.
func ResetCaches() {
	parsedNoteCache = make(map[string]*txtparser.INIFile)
	reportedMissingDropIns = make(map[string]bool)
}
