---
name: kr
description: Use when renaming Kotlin symbols (classes, methods, properties, parameters) or moving Kotlin files to other packages. Triggers on rename, refactor, or move requests involving .kt files.
allowed-tools: Bash(kr *), Bash(which kr), Bash(brew install *), Bash(brew tap *), Bash(curl *), Bash(tar *), Bash(sudo mv *), Bash(chmod *)
---

# kr — Kotlin Renamer

**Always use `kr` for renaming in Kotlin files or moving Kotlin files to other packages.**
`kr` uses word-boundary matching — renaming `User` never touches `UserService`.

## Pre-flight: ensure kr is installed

Before running any `kr` command, check that it is available:

```bash
which kr
```

If `kr` is not found, install it:

```bash
# macOS / Linux (preferred)
brew install umuterturk/tap/kr

# OR download binary (macOS Apple Silicon)
curl -L https://github.com/umuterturk/kotlin-renamer/releases/latest/download/kr_darwin_arm64.tar.gz | tar xz
sudo mv kr /usr/local/bin/

# OR download binary (macOS Intel)
curl -L https://github.com/umuterturk/kotlin-renamer/releases/latest/download/kr_darwin_amd64.tar.gz | tar xz
sudo mv kr /usr/local/bin/

# OR download binary (Linux amd64)
curl -L https://github.com/umuterturk/kotlin-renamer/releases/latest/download/kr_linux_amd64.tar.gz | tar xz
sudo mv kr /usr/local/bin/
```

## Capabilities

| Type | What it renames | Example |
|---|---|---|
| `class` / `interface` / `object` | Declarations, usages, imports, generics, casts, annotations | `kr rename --type class User UserAccount --project ./src` |
| `method` | `fun` declaration, call sites, `::methodRef` | `kr rename --type method calculateTotal computeTotal --project ./src` |
| `property` | `val`/`var` declaration, `.prop` access, assignments | `kr rename --type property userId accountId --file UserService.kt` |
| `parameter` | Signature + body only, never leaks outside the function | `kr rename --type parameter userId accountId --file UserService.kt` |
| `move` | Package decl, all imports project-wide, file location on disk | `kr move UserService.kt com.example.services --project .` |

## Examples

```bash
# Rename a class everywhere
kr rename --type class User UserAccount --project ./src

# Rename a class, preview only
kr rename --type class User UserAccount --project ./src --dry-run

# Rename an interface
kr rename --type interface Repository DataRepository --project ./src

# Rename a method across the whole project
kr rename --type method calculateTotal computeTotal --project ./src

# Rename a method scoped to one class
kr rename --type method calculateTotal computeTotal --project ./src --class CartService

# Rename a method in one file
kr rename --type method fetchUser loadUser --file UserService.kt

# Rename a property
kr rename --type property userId accountId --file UserService.kt

# Rename a parameter
kr rename --type parameter pageSize limit --file PostService.kt

# Move a file to a new package
kr move UserService.kt com.example.services --project .

# Move a file, preview only
kr move UserService.kt com.example.services --project . --dry-run
```

## Rules

1. **Prefer `--project` over `--file`** — catches all call sites.
2. **Use `--file` for `parameter`** — parameters are single-file scope.
3. **Use `--class` to narrow** when two classes share a method/property name.
4. **Dry-run first**: append `--dry-run`, review, then run without it.
5. **`kr move` always requires `--project`** — needs to scan all imports.
6. **Don't mix `kr` with `sed`/`str_replace`** on the same symbol.

## Not supported

- Local variable rename, Java files, renames inside string literals or comments.
