package renamer

import (
	"strings"
	"testing"
)

// ─── Class Rename Tests ────────────────────────────────────────────────────────

func TestClassRename_Declaration(t *testing.T) {
	r := &ClassRenamer{}
	src := `class User(val name: String)`
	got, n := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "class UserAccount")
	assertNotContains(t, got, "class User(")
	assertCount(t, n, 1)
}

func TestClassRename_DoesNotTouchSubstring(t *testing.T) {
	r := &ClassRenamer{}
	src := `
class User(val name: String)
class UserService(val user: User)
`
	got, n := r.Rename(src, "User", "UserAccount")
	// UserService must NOT be touched
	assertContains(t, got, "UserService")
	assertNotContains(t, got, "UserAccountService")
	// The standalone User in the parameter type should be renamed
	assertContains(t, got, "UserAccount")
	_ = n
}

func TestClassRename_Import(t *testing.T) {
	r := &ClassRenamer{}
	src := `import com.example.User
import com.example.UserService`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "import com.example.UserAccount")
	// UserService import must NOT change
	assertContains(t, got, "import com.example.UserService")
	assertNotContains(t, got, "import com.example.UserAccountService")
}

func TestClassRename_TypeAnnotation(t *testing.T) {
	r := &ClassRenamer{}
	src := `fun doSomething(user: User): User {
    return user
}`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "user: UserAccount")
	assertContains(t, got, "): UserAccount")
}

func TestClassRename_Generics(t *testing.T) {
	r := &ClassRenamer{}
	src := `val list: List<User> = mutableListOf<User>()`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "List<UserAccount>")
	assertContains(t, got, "mutableListOf<UserAccount>")
}

func TestClassRename_Inheritance(t *testing.T) {
	r := &ClassRenamer{}
	src := `class AdminUser : User(), Serializable`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, ": UserAccount()")
	assertNotContains(t, got, ": User()")
}

func TestClassRename_Annotation(t *testing.T) {
	r := &ClassRenamer{}
	src := `@User
class Something`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "@UserAccount")
}

func TestClassRename_As(t *testing.T) {
	r := &ClassRenamer{}
	src := `val u = something as User
val v = something as? User`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "as UserAccount")
	assertContains(t, got, "as? UserAccount")
}

func TestClassRename_Is(t *testing.T) {
	r := &ClassRenamer{}
	src := `when (x) {
    is User -> true
    !is User -> false
}`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "is UserAccount")
	assertContains(t, got, "!is UserAccount")
	assertNotContains(t, got, "is User ")
}

func TestClassRename_QualifiedName(t *testing.T) {
	r := &ClassRenamer{}
	src := `val x = com.example.User()`
	got, _ := r.Rename(src, "User", "UserAccount")
	assertContains(t, got, "com.example.UserAccount")
}

func TestClassRename_Interface(t *testing.T) {
	r := &ClassRenamer{}
	src := `interface Repository {
    fun findAll(): List<Repository>
}`
	got, _ := r.Rename(src, "Repository", "DataRepository")
	assertContains(t, got, "interface DataRepository")
	assertContains(t, got, "List<DataRepository>")
}

func TestClassRename_SealedAndData(t *testing.T) {
	r := &ClassRenamer{}
	src := `sealed class Result
data class Result(val value: String)`
	got, _ := r.Rename(src, "Result", "Outcome")
	assertContains(t, got, "sealed class Outcome")
	assertContains(t, got, "data class Outcome")
}

// ─── Method Rename Tests ───────────────────────────────────────────────────────

func TestMethodRename_Declaration(t *testing.T) {
	r := &MethodRenamer{}
	src := `fun calculateTotal(): Int { return 0 }`
	got, n := r.Rename(src, "calculateTotal", "computeTotal")
	assertContains(t, got, "fun computeTotal()")
	assertCount(t, n, 1)
}

func TestMethodRename_CallSite(t *testing.T) {
	r := &MethodRenamer{}
	src := `val total = cart.calculateTotal()
val x = calculateTotal()`
	got, n := r.Rename(src, "calculateTotal", "computeTotal")
	assertContains(t, got, "cart.computeTotal()")
	assertContains(t, got, "= computeTotal()")
	assertCount(t, n, 2)
}

func TestMethodRename_MethodReference(t *testing.T) {
	r := &MethodRenamer{}
	src := `val fn = ::calculateTotal
val fn2 = cart::calculateTotal`
	got, n := r.Rename(src, "calculateTotal", "computeTotal")
	assertContains(t, got, "::computeTotal")
	assertCount(t, n, 2)
}

func TestMethodRename_DoesNotRenameVariable(t *testing.T) {
	r := &MethodRenamer{}
	// calculateTotal used as a variable name (not followed by ( or ::) should NOT be renamed
	src := `val calculateTotal = 5`
	got, n := r.Rename(src, "calculateTotal", "computeTotal")
	// val calculateTotal is a property, not a function call — should not be renamed by MethodRenamer
	assertNotContains(t, got, "computeTotal")
	assertCount(t, n, 0)
}

// ─── Property Rename Tests ─────────────────────────────────────────────────────

func TestPropertyRename_Declaration(t *testing.T) {
	r := &PropertyRenamer{}
	src := `val userId: String = "abc"
var userId: Int = 0`
	got, n := r.Rename(src, "userId", "accountId")
	assertContains(t, got, "val accountId")
	assertContains(t, got, "var accountId")
	assertCount(t, n, 2)
}

func TestPropertyRename_MemberAccess(t *testing.T) {
	r := &PropertyRenamer{}
	src := `println(user.userId)
this.userId = newId`
	got, n := r.Rename(src, "userId", "accountId")
	assertContains(t, got, "user.accountId")
	assertContains(t, got, "this.accountId")
	assertCount(t, n, 2)
}

func TestPropertyRename_Assignment(t *testing.T) {
	r := &PropertyRenamer{}
	src := `userId = "new-id"`
	got, n := r.Rename(src, "userId", "accountId")
	assertContains(t, got, "accountId =")
	assertCount(t, n, 1)
}

func TestPropertyRename_BareRead(t *testing.T) {
	r := &PropertyRenamer{}
	src := `fun apply() {
    val d = comboDiscount
    this.comboDiscount = d * 2
    println(comboDiscount)
    if (comboDiscount > 0) { }
    comboDiscount = 0.0
}`
	got, n := r.Rename(src, "comboDiscount", "bundle")
	assertContains(t, got, "val d = bundle")
	assertContains(t, got, "this.bundle = d * 2")
	assertContains(t, got, "println(bundle)")
	assertContains(t, got, "if (bundle > 0)")
	assertContains(t, got, "bundle = 0.0")
	assertNotContains(t, got, "comboDiscount")
	if n != 5 {
		t.Errorf("expected 5 replacements, got %d", n)
	}
}

func TestPropertyRename_DoesNotRenameMethodCall(t *testing.T) {
	r := &PropertyRenamer{}
	// comboDiscount() would be a method call — PropertyRenamer must leave it alone
	src := `val x = comboDiscount()`
	got, n := r.Rename(src, "comboDiscount", "bundle")
	assertNotContains(t, got, "bundle()")
	assertCount(t, n, 0)
}

// ─── Parameter Rename Tests ────────────────────────────────────────────────────

func TestParameterRename_SignatureAndBody(t *testing.T) {
	r := &ParameterRenamer{}
	src := `fun greet(userId: String): String {
    return "Hello $userId"
}`
	got, n := r.Rename(src, "userId", "accountId")
	assertContains(t, got, "accountId: String")
	assertContains(t, got, "$accountId")
	if n == 0 {
		t.Error("expected at least 1 replacement")
	}
}

func TestParameterRename_DoesNotTouchOutsideScope(t *testing.T) {
	r := &ParameterRenamer{}
	// userId declared at class level should NOT be renamed if it's not a parameter
	src := `val userId = "global"

fun greet(name: String): String {
    return "Hello $name and $userId"
}`
	// userId is NOT a parameter of greet(), so it should NOT be renamed
	got, _ := r.Rename(src, "userId", "accountId")
	// The global userId should remain
	if strings.Contains(got, "val accountId") {
		t.Error("should not rename class-level property when using ParameterRenamer")
	}
}

// ─── Package Move Tests ────────────────────────────────────────────────────────

func TestRewriteImport(t *testing.T) {
	src := `import com.example.User
import com.example.UserService
import com.other.Foo`

	got, n := rewriteImport(src, "com.example.User", "com.example.newpkg.User")
	assertContains(t, got, "import com.example.newpkg.User")
	assertContains(t, got, "import com.example.UserService") // NOT changed
	assertContains(t, got, "import com.other.Foo")           // NOT changed
	assertCount(t, n, 1)
}

func TestExtractPackage(t *testing.T) {
	src := `package com.example.foo

import com.example.bar.Baz`
	pkg := extractPackage(src)
	if pkg != " com.example.foo" && pkg != "com.example.foo" {
		t.Errorf("extractPackage returned %q, want com.example.foo", pkg)
	}
}

func TestRewritePackageDeclaration(t *testing.T) {
	src := `package com.example.old

class Foo`
	got := rewritePackageDeclaration(src, "com.example.new")
	assertContains(t, got, "package com.example.new")
	assertNotContains(t, got, "package com.example.old")
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q\ngot:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", want, got)
	}
}

func assertCount(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("replacement count: got %d, want %d", got, want)
	}
}
