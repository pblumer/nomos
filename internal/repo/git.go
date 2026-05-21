package repo

// git.go adds local git lifecycle operations (status, commit, branches, tags)
// on top of the read-only probing in repo.go. All operations shell out to the
// git binary against a repository's working directory; remote/push handling is
// intentionally out of scope (local git-first only, ADR-0022 §3).

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileChange is one entry of a porcelain status: a two-letter code and the path
// it applies to (e.g. {Code: "??", Path: "README.md"} for an untracked file).
type FileChange struct {
	Code string
	Path string
}

// GitStatus is the working-tree state used by the commit UI.
type GitStatus struct {
	Initialized bool
	Branch      string
	Head        string
	Dirty       bool
	Files       []FileChange
}

// Tag is an annotated or lightweight git tag.
type Tag struct {
	Name    string
	Message string
}

// hasGit reports whether loc contains a git database.
func hasGit(loc string) bool {
	_, err := os.Stat(filepath.Join(loc, ".git"))
	return err == nil
}

// Status returns the working-tree state of the repository at loc. A directory
// without a git database reports Initialized=false.
func Status(loc string) (GitStatus, error) {
	st := GitStatus{}
	if !hasGit(loc) {
		return st, nil
	}
	st.Initialized = true
	st.Head = gitOutput(loc, "rev-parse", "--short", "HEAD")
	st.Branch = gitOutput(loc, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := runGit(loc, "status", "--porcelain")
	if err != nil {
		return st, err
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Porcelain v1: "XY <path>"; codes occupy the first two columns.
		code := strings.TrimSpace(line[:min(2, len(line))])
		path := strings.TrimSpace(line[min(3, len(line)):])
		st.Files = append(st.Files, FileChange{Code: code, Path: path})
	}
	st.Dirty = len(st.Files) > 0
	return st, nil
}

// Commit stages every change and records a commit. It initializes a git
// database first when the location has none, so a freshly created filesystem
// repository can take its first commit. A missing git identity falls back to a
// neutral Nomos author so commits never fail in a bare container.
func Commit(loc, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("commit message is required")
	}
	if !hasGit(loc) {
		if err := run(loc, "init", "-q"); err != nil {
			return err
		}
	}
	if err := run(loc, "add", "-A"); err != nil {
		return err
	}
	if gitOutput(loc, "status", "--porcelain") == "" {
		return fmt.Errorf("nothing to commit: working tree is clean")
	}
	args := []string{"commit", "-m", message}
	if gitOutput(loc, "config", "user.email") == "" {
		// A synthetic identity cannot satisfy commit signing, so disable it for
		// this commit only; a configured user keeps their signing settings.
		args = append([]string{
			"-c", "user.name=Nomos",
			"-c", "user.email=nomos@localhost",
			"-c", "commit.gpgsign=false",
		}, args...)
	}
	return run(loc, args...)
}

// Branches returns the current branch and all local branch names.
func Branches(loc string) (current string, all []string, err error) {
	if !hasGit(loc) {
		return "", nil, nil
	}
	current = gitOutput(loc, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := runGit(loc, "branch", "--format=%(refname:short)")
	if err != nil {
		return current, nil, err
	}
	for _, line := range strings.Split(out, "\n") {
		if b := strings.TrimSpace(line); b != "" {
			all = append(all, b)
		}
	}
	return current, all, nil
}

// CreateBranch creates a new branch. When checkout is true it also switches to
// it. A repository with no commits yet cannot branch, so the caller should
// commit first.
func CreateBranch(loc, name string, checkout bool) error {
	if err := validRef(name); err != nil {
		return err
	}
	if !hasGit(loc) {
		return fmt.Errorf("repository is not initialized")
	}
	if checkout {
		return run(loc, "checkout", "-b", name)
	}
	return run(loc, "branch", name)
}

// Checkout switches the working tree to an existing branch.
func Checkout(loc, name string) error {
	if err := validRef(name); err != nil {
		return err
	}
	if !hasGit(loc) {
		return fmt.Errorf("repository is not initialized")
	}
	return run(loc, "checkout", name)
}

// Tags returns the repository's tags, newest first, with their annotation
// message (empty for lightweight tags).
func Tags(loc string) ([]Tag, error) {
	if !hasGit(loc) {
		return nil, nil
	}
	out, err := runGit(loc, "tag", "--sort=-creatordate", "--format=%(refname:short)%09%(contents:subject)")
	if err != nil {
		return nil, err
	}
	var tags []Tag
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		name, msg, _ := strings.Cut(line, "\t")
		tags = append(tags, Tag{Name: strings.TrimSpace(name), Message: strings.TrimSpace(msg)})
	}
	return tags, nil
}

// CreateTag creates a tag at HEAD. A non-empty message produces an annotated
// tag; otherwise a lightweight tag is created.
func CreateTag(loc, name, message string) error {
	if err := validRef(name); err != nil {
		return err
	}
	if !hasGit(loc) {
		return fmt.Errorf("repository is not initialized")
	}
	if gitOutput(loc, "rev-parse", "--verify", "-q", "HEAD") == "" {
		return fmt.Errorf("cannot tag a repository with no commits yet")
	}
	if msg := strings.TrimSpace(message); msg != "" {
		args := []string{"tag", "-a", name, "-m", msg}
		if gitOutput(loc, "config", "user.email") == "" {
			args = append([]string{"-c", "user.name=Nomos", "-c", "user.email=nomos@localhost", "-c", "tag.gpgsign=false"}, args...)
		}
		return run(loc, args...)
	}
	return run(loc, "tag", name)
}

// validRef rejects names git would refuse or that could inject flags.
func validRef(name string) error {
	name = strings.TrimSpace(name)
	switch {
	case name == "":
		return fmt.Errorf("name is required")
	case strings.HasPrefix(name, "-"):
		return fmt.Errorf("invalid name: %q", name)
	case strings.ContainsAny(name, " ~^:?*[\\\x7f") || strings.Contains(name, ".."):
		return fmt.Errorf("invalid name: %q", name)
	}
	return nil
}

// run executes a git command, surfacing stderr on failure.
func run(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return nil
}

// runGit returns trimmed stdout, surfacing errors (unlike gitOutput which
// swallows them for best-effort probing).
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
