package renamer

import (
	"fmt"
	"io"
	"sort"
)

// PrintResults writes the standard kr output format to w.
//
//	✅ CartService.kt: 4 replacement(s)
//	✅ InvoiceService.kt: 2 replacement(s)
//	Total: 6 replacement(s) across 2 file(s)
func PrintResults(w io.Writer, results []FileResult, dryRun bool) {
	// Sort for deterministic output
	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})

	totalReplacements := 0
	filesChanged := 0

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "❌ %s: error: %v\n", r.Path, r.Err)
			continue
		}
		if r.Replacements > 0 {
			mark := "✅"
			if dryRun {
				mark = "🔍"
			}
			fmt.Fprintf(w, "%s %s: %d replacement(s)\n", mark, r.Path, r.Replacements)
			totalReplacements += r.Replacements
			filesChanged++
		}
	}

	if totalReplacements == 0 {
		fmt.Fprintln(w, "No replacements found.")
		return
	}

	suffix := ""
	if dryRun {
		suffix = " (dry run — no files written)"
	}
	fmt.Fprintf(w, "Total: %d replacement(s) across %d file(s)%s\n",
		totalReplacements, filesChanged, suffix)
}

// PrintMoveResult writes the move command output.
func PrintMoveResult(w io.Writer, r *MoveResult, dryRun bool) {
	verb := "Moved"
	if dryRun {
		verb = "Would move"
	}
	fmt.Fprintf(w, "%s: %s\n    → %s\n", verb, r.MovedFrom, r.MovedTo)

	if len(r.ImportResults) > 0 {
		fmt.Fprintln(w, "Import updates:")
		PrintResults(w, r.ImportResults, dryRun)
	} else {
		fmt.Fprintln(w, "No import statements needed updating.")
	}
}
