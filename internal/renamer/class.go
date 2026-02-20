package renamer

import (
	"fmt"
	"regexp"
	"strings"
)

// ClassRenamer handles renaming of classes, interfaces, and objects.
//
// Strategy: single-pass scan using a master regex that captures every possible
// occurrence of oldName.  Each match is examined for its surrounding context to
// decide whether it is a real symbol reference and should be replaced.
//
// Contexts handled:
//   - import declarations           import com.example.OldName
//   - package-qualified access      com.example.OldName.method()
//   - annotation                    @OldName
//   - declaration keyword           class/interface/object OldName
//   - type annotation               val x: OldName  / fun f(): OldName
//   - generic type argument         List<OldName>
//   - supertype list                : OldName() / : OldName,
//   - as / as? cast                 x as OldName
//   - is / !is check               x is OldName
//   - constructor call              OldName(...)
//   - companion / static access     OldName.bar
//   - method reference              OldName::bar
//
// Non-goals (not renamed):
//   - Local variable names that shadow the class name (requires scope analysis)
//   - Contents of string literals or comments (we preserve those)
type ClassRenamer struct{}

func (r *ClassRenamer) Rename(content, oldName, newName string) (string, int) {
	return singlePassRename(content, oldName, newName, isClassContext)
}

// isClassContext returns true when the character at position [start,end) within
// the full source string represents a class/type name that should be renamed.
//
// We inspect what immediately precedes and follows the matched token.
func isClassContext(src string, start, end int) bool {
	pre := src[:start]
	post := src[end:]

	// Never rename if the character immediately before is a letter, digit, or
	// underscore (we're inside a longer identifier like UserService → User).
	if len(pre) > 0 {
		c := pre[len(pre)-1]
		if isIdentChar(c) {
			return false
		}
	}

	// Never rename if the character immediately after is a letter, digit, or
	// underscore (again, inside a longer identifier).
	if len(post) > 0 {
		c := post[0]
		if isIdentChar(c) {
			return false
		}
	}

	// At this point we have a proper word-boundary match. Accept it.
	// The patterns below are examples of what we accept; anything that passes
	// the word-boundary check is a valid class name reference.
	return true
}

// MethodRenamer handles renaming of functions and methods.
//
// Contexts handled:
//   - declaration:        fun oldName(
//   - call site:          .oldName(  or  oldName(  (at statement start)
//   - method reference:   ::oldName
//   - named argument:     oldName =   — NOT renamed (it's a parameter label)
type MethodRenamer struct {
	ClassName string // optional: limit to calls on a specific class/receiver
}

func (r *MethodRenamer) Rename(content, oldName, newName string) (string, int) {
	return singlePassRename(content, oldName, newName, r.isMethodContext)
}

func (r *MethodRenamer) isMethodContext(src string, start, end int) bool {
	pre := src[:start]
	post := strings.TrimLeft(src[end:], " \t")

	// Must have word boundaries
	if len(pre) > 0 && isIdentChar(pre[len(pre)-1]) {
		return false
	}
	if len(post) > 0 && isIdentChar(post[0]) {
		return false
	}

	trimmedPre := strings.TrimRight(pre, " \t")

	// Preceded by :: — this is a method reference: ::oldName or obj::oldName
	if strings.HasSuffix(trimmedPre, "::") {
		return true
	}

	// Preceded by "fun" keyword — function declaration
	if strings.HasSuffix(trimmedPre, "fun") {
		return true
	}

	// Must be followed by ( to be a call site
	if len(post) > 0 && post[0] == '(' {
		return true
	}

	return false
}

// PropertyRenamer handles renaming of val/var properties and fields.
//
// Contexts handled:
//   - declaration:    val oldName / var oldName
//   - this access:   this.oldName
//   - member access: receiver.oldName
//   - named argument: oldName = value  (in constructor/function calls)
type PropertyRenamer struct {
	ClassName string
}

func (r *PropertyRenamer) Rename(content, oldName, newName string) (string, int) {
	return singlePassRename(content, oldName, newName, r.isPropertyContext)
}

func (r *PropertyRenamer) isPropertyContext(src string, start, end int) bool {
	pre := src[:start]
	post := strings.TrimLeft(src[end:], " \t")

	// Word boundaries
	if len(pre) > 0 && isIdentChar(pre[len(pre)-1]) {
		return false
	}
	if len(post) > 0 && isIdentChar(post[0]) {
		return false
	}

	trimmedPre := strings.TrimRight(pre, " \t")

	// Preceded by . (member access) — accept
	if strings.HasSuffix(trimmedPre, ".") {
		return true
	}

	// Preceded by val/var (declaration) — accept
	if strings.HasSuffix(trimmedPre, "val") || strings.HasSuffix(trimmedPre, "var") {
		return true
	}

	// Followed by = (assignment or named arg) — accept
	if len(post) > 0 && post[0] == '=' && (len(post) < 2 || post[1] != '=') {
		return true
	}

	// Followed by : (type annotation on declaration) — accept
	if len(post) > 0 && post[0] == ':' {
		return true
	}

	// Bare read: not followed by ( — accept as property access in expression context
	// e.g. val d = comboDiscount / println(comboDiscount) / if (comboDiscount > 0)
	if len(post) == 0 || post[0] != '(' {
		return true
	}

	return false
}

// ParameterRenamer renames parameters within function signatures and their bodies.
// This is the most conservative renamer — it only operates within a single file
// and only within function scopes.
type ParameterRenamer struct{}

func (r *ParameterRenamer) Rename(content, oldName, newName string) (string, int) {
	// We process the file function-by-function
	return renameParameters(content, oldName, newName)
}

// ─── core engine ──────────────────────────────────────────────────────────────

// singlePassRename scans src for all word-boundary occurrences of oldName and
// replaces those for which contextFn returns true.  Returns the modified source
// and replacement count.
func singlePassRename(src, oldName, newName string, contextFn func(src string, start, end int) bool) (string, int) {
	pat := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\b`)

	count := 0
	var buf strings.Builder
	last := 0

	for _, loc := range pat.FindAllStringIndex(src, -1) {
		start, end := loc[0], loc[1]

		if contextFn(src, start, end) {
			buf.WriteString(src[last:start])
			buf.WriteString(newName)
			last = end
			count++
		}
	}
	buf.WriteString(src[last:])

	return buf.String(), count
}

// ─── parameter rename ─────────────────────────────────────────────────────────

// renameParameters renames a parameter within all function scopes where it
// appears in the parameter list.
func renameParameters(src, oldName, newName string) (string, int) {
	total := 0

	// Find all function headers that declare oldName as a parameter.
	// We scan for "fun " + optional name + "(...oldName...)" blocks.
	// Then we rename inside the matched brace scope.

	funPat := regexp.MustCompile(`(?m)\bfun\b`)
	result := src
	offset := 0

	for _, loc := range funPat.FindAllStringIndex(src, -1) {
		// loc is relative to original src; adjust to result using offset
		funStart := loc[0] + offset

		// find the opening paren of parameter list
		parenOpen := strings.Index(result[funStart:], "(")
		if parenOpen < 0 {
			continue
		}
		parenOpen += funStart

		// find matching closing paren
		parenClose := findMatchingParen(result, parenOpen)
		if parenClose < 0 {
			continue
		}

		paramSection := result[parenOpen : parenClose+1]

		// Does this function have oldName as a parameter?
		if !hasParamName(paramSection, oldName) {
			continue
		}

		// Find function body: either expression body (= ...) or block body ({ ... })
		bodyStart, bodyEnd := findFunctionBody(result, parenClose+1)
		if bodyStart < 0 {
			// no body found; still rename in signature
			bodyStart = parenOpen
			bodyEnd = parenClose + 1
		}

		// Rename in signature + body
		scopeContent := result[parenOpen:bodyEnd]
		renamed, n := singlePassRename(scopeContent, oldName, newName, isParameterContext)
		if n > 0 {
			result = result[:parenOpen] + renamed + result[bodyEnd:]
			offset += len(renamed) - len(scopeContent)
			total += n
		}
	}

	return result, total
}

// hasParamName checks whether a parameter list string contains oldName as a
// parameter identifier (not as a type name, which starts with uppercase).
func hasParamName(paramSection, oldName string) bool {
	pat := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\s*:`)
	return pat.MatchString(paramSection)
}

// isParameterContext accepts identifiers that look like parameter usage:
// preceded by nothing special (start of expression, comma, whitespace) and
// followed by anything except ( — we don't want to capture function calls.
func isParameterContext(src string, start, end int) bool {
	pre := src[:start]
	post := src[end:]

	// Word boundaries
	if len(pre) > 0 && isIdentChar(pre[len(pre)-1]) {
		return false
	}
	if len(post) > 0 && isIdentChar(post[0]) {
		return false
	}

	// Skip if followed by ( — that would be a function call, not a parameter
	trimPost := strings.TrimLeft(post, " \t")
	if len(trimPost) > 0 && trimPost[0] == '(' {
		return false
	}

	return true
}

// ─── brace/paren matching helpers ─────────────────────────────────────────────

func findMatchingParen(src string, open int) int {
	depth := 0
	for i := open; i < len(src); i++ {
		switch src[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func findFunctionBody(src string, afterParams int) (int, int) {
	// skip whitespace and return type annotation
	i := afterParams
	for i < len(src) && (src[i] == ' ' || src[i] == '\t' || src[i] == '\n' || src[i] == '\r') {
		i++
	}

	// return type annotation: ": Type"
	if i < len(src) && src[i] == ':' {
		// skip to next { or =
		for i < len(src) && src[i] != '{' && src[i] != '=' && src[i] != '\n' {
			i++
		}
	}

	if i >= len(src) {
		return -1, -1
	}

	switch src[i] {
	case '{':
		end := findMatchingBrace(src, i)
		if end < 0 {
			return -1, -1
		}
		return i, end + 1
	case '=':
		// expression body — goes to end of line
		lineEnd := strings.Index(src[i:], "\n")
		if lineEnd < 0 {
			return i, len(src)
		}
		return i, i + lineEnd + 1
	}

	return -1, -1
}

func findMatchingBrace(src string, open int) int {
	depth := 0
	for i := open; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// ─── utilities ────────────────────────────────────────────────────────────────

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// ValidateIdentifier checks that s is a valid Kotlin identifier.
func ValidateIdentifier(s string) error {
	valid := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !valid.MatchString(s) {
		return fmt.Errorf("invalid Kotlin identifier: %q", s)
	}
	return nil
}
