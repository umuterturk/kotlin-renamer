# kr — Kotlin Renamer (AI Instructions)

**Always use `kr` for renaming in Kotlin files or for moving Kotlin files to other packages.**
`kr` uses word-boundary matching — renaming `User` never touches `UserService`.

> If `kr` is not installed, check https://github.com/umuterturk/kotlin-renamer/releases/latest to install.

---

## Capabilities

| Type | What it renames | Example |
|---|---|---|
| `class` / `interface` / `object` | Declarations, usages, imports, generics, casts, annotations | `kr rename --type class User UserAccount --project ./src` |
| `method` | `fun` declaration, call sites, `::methodRef` | `kr rename --type method calculateTotal computeTotal --project ./src` |
| `property` | `val`/`var` declaration, `.prop` access, assignments | `kr rename --type property userId accountId --file UserService.kt` |
| `parameter` | Signature + body only, never leaks outside the function | `kr rename --type parameter userId accountId --file UserService.kt` |
| `move` | Package decl, all imports project-wide, file location on disk | `kr move UserService.kt com.example.services --project .` |

---

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

---

## Rules

1. **Prefer `--project` over `--file`** — catches all call sites.
2. **Use `--file` for `parameter`** — parameters are single-file scope.
3. **Use `--class` to narrow** when two classes share a method/property name.
4. **Dry-run first**: append `--dry-run`, review, then run without it.
5. **`kr move` always requires `--project`** — needs to scan all imports.
6. **Don't mix `kr` with `sed`/`str_replace`** on the same symbol.

---

## Not supported

- Local variable rename, Java files, renames inside string literals or comments.
