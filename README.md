# kr — Kotlin Renamer

[![Release](https://img.shields.io/github/v/release/umuterturk/kotlin-renamer)](https://github.com/umuterturk/kotlin-renamer/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/umuterturk/kotlin-renamer)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/umuterturk/kotlin-renamer/release.yml?label=build)](https://github.com/umuterturk/kotlin-renamer/actions)

A fast, syntax-aware rename CLI for Kotlin projects. Feels closer to IntelliJ's **Shift+F6** than a bash script.

```
✅ CartService.kt: 4 replacement(s)
✅ InvoiceService.kt: 2 replacement(s)
Total: 6 replacement(s) across 2 file(s)
```

---

## The problem

When you can't have IntelliJ open, renaming Kotlin symbols with `sed` or AI assistants breaks:

| Problem | Example |
|---|---|
| Hits substrings | Renaming `User` → `Account` also mangles `UserService` |
| Misses qualified names | `com.example.User` left unchanged |
| Misses imports | Other files still import the old name |
| No type context | Can't scope a method rename to one class |

`kr` fixes all of this with word-boundary matching and Kotlin-specific context rules.

---

## Install

### Homebrew (macOS / Linux)
```bash
brew install umuterturk/tap/kr
```

### Download binary
Grab the latest binary for your platform from the [Releases](https://github.com/umuterturk/kotlin-renamer/releases/latest) page.

```bash
# macOS (Apple Silicon)
curl -L https://github.com/umuterturk/kotlin-renamer/releases/latest/download/kr_darwin_arm64.tar.gz | tar xz
sudo mv kr /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/umuterturk/kotlin-renamer/releases/latest/download/kr_darwin_amd64.tar.gz | tar xz
sudo mv kr /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/umuterturk/kotlin-renamer/releases/latest/download/kr_linux_amd64.tar.gz | tar xz
sudo mv kr /usr/local/bin/
```

### Build from source
```bash
git clone https://github.com/umuterturk/kotlin-renamer.git
cd kotlin-renamer
go build -o kr .
sudo mv kr /usr/local/bin/
```

---

## Usage

```
kr rename <old> <new> [flags]
kr move   <file> <new.package> [flags]
```

### Flags — rename

| Flag | Description |
|---|---|
| `--type` | Symbol type: `class`, `interface`, `object`, `method`, `property`, `parameter` (default: `class`) |
| `--project` | Project root — scans all `.kt` files recursively |
| `--file` | Restrict to a single file |
| `--class` | Scope `method`/`property` rename to a specific class |
| `--dry-run` | Preview changes without writing |

### Flags — move

| Flag | Description |
|---|---|
| `--project` | Project root — required, used to scan all `.kt` files for import rewriting |
| `--dry-run` | Preview changes without writing |

---

## Examples

**Rename a class everywhere**
```bash
kr rename --type class User UserAccount --project ./src
```
```kotlin
// before                              // after
class User(val name: String)           class UserAccount(val name: String)
import com.example.User                import com.example.UserAccount
val list: List<User>                   val list: List<UserAccount>
fun save(u: User): User                fun save(u: UserAccount): UserAccount
x as? User                            x as? UserAccount
class UserService { ... }              class UserService { ... }  // ← untouched
```

**Rename a class, preview only**
```bash
kr rename --type class User UserAccount --project ./src --dry-run
```

**Rename an interface**
```bash
kr rename --type interface Repository DataRepository --project ./src
```
```kotlin
// before                              // after
interface Repository                   interface DataRepository
class UserRepo : Repository            class UserRepo : DataRepository
fun save(r: Repository)                fun save(r: DataRepository)
```

**Rename a method across the whole project**
```bash
kr rename --type method calculateTotal computeTotal --project ./src
```
```kotlin
// before                              // after
fun calculateTotal(): Int              fun computeTotal(): Int
cart.calculateTotal()                  cart.computeTotal()
val fn = ::calculateTotal              val fn = ::computeTotal
```

**Rename a method scoped to one class**
```bash
kr rename --type method calculateTotal computeTotal --project ./src --class CartService
```
```kotlin
// before                              // after
cart.calculateTotal()   // CartService  cart.computeTotal()      // ✅ renamed
order.calculateTotal()  // OrderService order.calculateTotal()   // ✅ untouched
```

**Rename a method in one file**
```bash
kr rename --type method fetchUser loadUser --file UserService.kt
```

**Rename a property**
```bash
kr rename --type property userId accountId --file UserService.kt
```
```kotlin
// before                              // after
val userId: String                     val accountId: String
this.userId = newId                    this.accountId = newId
println(user.userId)                   println(user.accountId)
```

**Rename a parameter**
```bash
kr rename --type parameter pageSize limit --file PostService.kt
```
```kotlin
// before                              // after
fun fetch(pageSize: Int) {             fun fetch(limit: Int) {
    query(pageSize)                        query(limit)
}
val pageSize = 20  // class-level       val pageSize = 20  // ← untouched
```

**Move a file to a new package**
```bash
kr move UserService.kt com.example.services --project .
```
```kotlin
// UserService.kt                      // after
package com.example                    package com.example.services

// CartService.kt                      // after
import com.example.UserService         import com.example.services.UserService

// file on disk
// src/main/kotlin/com/example/UserService.kt
//   → src/main/kotlin/com/example/services/UserService.kt
```

**Move a file, preview only**
```bash
kr move UserService.kt com.example.services --project . --dry-run
```

---

## What kr does NOT handle

| Scenario | Reason |
|---|---|
| Local variable rename | Requires cursor position context |
| Java files | `.kt` only |
| Names inside string literals / comments | Intentionally skipped — edit manually |
| Rename a whole package | Run `kr move` on each file in the package |

---

## AI & Editor Integration

### Quick setup (recommended)

After installing `kr`, run:

```bash
kr setup
```

This installs both integrations at once:
- **Claude Code** skill → `~/.claude/skills/kr/SKILL.md` (global, all projects)
- **Cursor** rule → `.cursor/rules/kotlin-renamer.mdc` (current project)

You can install them individually with `kr setup --claude-code` or `kr setup --cursor`.

### Claude Code (skill — no project config needed)

The skill is installed globally. Claude Code will automatically use `kr` for any Kotlin rename or move task across all your projects. If `kr` is not on your PATH, Claude will offer to install it.

<details>
<summary>Manual install (without kr setup)</summary>

```bash
mkdir -p ~/.claude/skills/kr
curl -o ~/.claude/skills/kr/SKILL.md \
  https://raw.githubusercontent.com/umuterturk/kotlin-renamer/main/skills/claude-code/SKILL.md
```
</details>

### Cursor

The rule activates automatically when you open any `.kt` file.

<details>
<summary>Manual install (without kr setup)</summary>

```bash
mkdir -p .cursor/rules
curl -o .cursor/rules/kotlin-renamer.mdc \
  https://raw.githubusercontent.com/umuterturk/kotlin-renamer/main/skills/cursor/kotlin-renamer.mdc
```
</details>

### Other agents (AGENTS.md)
Some agent frameworks read `AGENTS.md` from the project root.

```bash
curl -o AGENTS.md \
  https://raw.githubusercontent.com/umuterturk/kotlin-renamer/main/AGENTS.md
```

---

## License

MIT
