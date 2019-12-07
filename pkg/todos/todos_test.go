package todos

import (
	"context"
	"flag"
	"sort"
	"testing"
	"time"

	"github.com/augmentable-dev/tickgit/pkg/comments"
	"gopkg.in/src-d/go-git.v4"
)

var dir string

func init() {
	flag.StringVar(&dir, "dir", "", "Location of the git repo directory")
}

func TestLargeRepository(t *testing.T) {
	r, err := git.PlainOpen(dir)
	if err != nil {
		t.Fatal(err)
	}

	ref, err := r.Head()
	if err != nil {
		t.Fatal(err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		t.Fatal(err)
	}

	foundToDos := make(ToDos, 0)
	err = comments.SearchDir(dir, func(comment *comments.Comment) {
		todo := NewToDo(*comment)
		if todo != nil {
			foundToDos = append(foundToDos, todo)
		}
	})

	ctx := context.Background()
	// timeout after 30 seconds
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err = foundToDos.FindBlame(ctx, r, commit, nil)

	if err != nil {
		t.Fatal(err)
	}

	sort.Sort(&foundToDos)
}
