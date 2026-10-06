package projects

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// GitStatus contains the working tree state of a git repository.
type GitStatus struct {
	HasUncommitted   bool
	UncommittedCount int
	OK               bool
}

// CheckGitStatus runs `git status --porcelain` with a timeout inside repoPath.
// If repoPath is not a git repository or git fails, it returns GitStatus{OK: false}.
func CheckGitStatus(repoPath string) GitStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoPath

	out, err := cmd.Output()
	if err != nil {
		return GitStatus{OK: false}
	}

	trimmed := bytes.TrimSpace(out)
	if len(trimmed) == 0 {
		return GitStatus{
			HasUncommitted:   false,
			UncommittedCount: 0,
			OK:               true,
		}
	}

	lines := strings.Split(string(trimmed), "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}

	return GitStatus{
		HasUncommitted:   count > 0,
		UncommittedCount: count,
		OK:               true,
	}
}
