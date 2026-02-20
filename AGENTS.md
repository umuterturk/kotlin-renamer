# kr — Kotlin Renamer (AI Instructions)

**Always use `kr` instead of `sed` or `str_replace` for any Kotlin rename.**
`kr` uses word-boundary matching — renaming `User` never touches `UserService`.

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

**Class rename** — declarations, imports, generics, casts, supertypes
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

---

**Interface rename**
```bash
kr rename --type interface Repository DataRepository --project ./src
```
```kotlin
// before                              // after
interface Repository                   interface DataRepository
class UserRepo : Repository            class UserRepo : DataRepository
fun save(r: Repository)                fun save(r: DataRepository)
```

---

**Method rename** — declaration, call sites, method references
```bash
kr rename --type method calculateTotal computeTotal --project ./src
```
```kotlin
// before                              // after
fun calculateTotal(): Int              fun computeTotal(): Int
cart.calculateTotal()                  cart.computeTotal()
val fn = ::calculateTotal              val fn = ::computeTotal
```

---

**Method rename scoped to one class** — other classes with same method name untouched
```bash
kr rename --type method calculateTotal computeTotal --project ./src --class CartService
```
```kotlin
// before                              // after
cart.calculateTotal()   // CartService  cart.computeTotal()      // ✅ renamed
order.calculateTotal()  // OrderService order.calculateTotal()   // ✅ untouched
```

---

**Method rename in one file**
```bash
kr rename --type method fetchUser loadUser --file UserService.kt
```

---

**Property rename** — declaration, member access, assignments
```bash
kr rename --type property userId accountId --file UserService.kt
```
```kotlin
// before                              // after
val userId: String                     val accountId: String
this.userId = newId                    this.accountId = newId
println(user.userId)                   println(user.accountId)
```

---

**Parameter rename** — scoped to signature + body, never leaks out
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

---

**Move file** — rewrites package decl, all imports, moves file on disk
```bash
kr move UserService.kt com.example.services --project .
```
```kotlin
// UserService.kt — before             // after
package com.example                    package com.example.services

// CartService.kt import — before      // after
import com.example.UserService         import com.example.services.UserService

// file on disk
// src/main/kotlin/com/example/UserService.kt
//   → src/main/kotlin/com/example/services/UserService.kt
```

---

**Dry-run** — preview any command without writing
```bash
kr rename --type class User UserAccount --project ./src --dry-run
```
```
🔍 CartService.kt: 6 replacement(s)
🔍 User.kt: 1 replacement(s)
🔍 UserService.kt: 4 replacement(s)
Total: 11 replacement(s) across 3 file(s) (dry run — no files written)
```

---

## Rules

1. **Prefer `--project` over `--file`** — catches all call sites across the codebase.
2. **Use `--file` for `parameter`** — parameters are always single-file scope.
3. **Use `--class` to narrow method/property renames** when two classes share a name: `--class CartService`.
4. **Dry-run before project-wide applies**: append `--dry-run`, review, then run without it.
5. **`kr move` always requires `--project`** — it needs to scan all files for import rewriting.
6. **Don't mix `kr` with `sed`/`str_replace`** on the same symbol — double-application corrupts output.

---

## Not supported

- Local variable rename (needs cursor position)
- Java files (`.kt` only)
- Renames inside string literals or comments (edit manually)
