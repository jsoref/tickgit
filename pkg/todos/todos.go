package todos

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/augmentable-dev/tickgit/pkg/comments"
	"github.com/dustin/go-humanize"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// ToDo represents a ToDo item
type ToDo struct {
	comments.Comment
	String string
	Commit *object.Commit
}

// Commit represents the commit a todo originated in
type Commit struct {
	Hash      string
	Author    Actor
	Committer Actor
}

// Actor represents the action of a commit a todo originated in (the author or committer)
type Actor struct {
	Name  string
	Email string
	When  time.Time
}

func (a *Actor) String() string {
	return fmt.Sprintf("%s <%s>", a.Name, a.Email)
}

// ToDos represents a list of ToDo items
type ToDos []*ToDo

// TimeAgo returns a human readable string indicating the time since the todo was added
func (t *ToDo) TimeAgo() string {
	if t.Commit == nil {
		return "<unknown>"
	}
	return humanize.Time(t.Commit.Author.When)
}

// NewToDo produces a pointer to a ToDo from a comment
func NewToDo(comment comments.Comment) *ToDo {
	s := comment.String()
	if !strings.Contains(s, "TODO") {
		return nil
	}
	re := regexp.MustCompile(`TODO(:|,)?`)
	s = re.ReplaceAllLiteralString(comment.String(), "")
	s = strings.Trim(s, " ")

	todo := ToDo{Comment: comment, String: s}
	return &todo
}

// NewToDos produces a list of ToDos from a list of comments
func NewToDos(comments comments.Comments) ToDos {
	todos := make(ToDos, 0)
	for _, comment := range comments {
		todo := NewToDo(*comment)
		if todo != nil {
			todos = append(todos, todo)
		}
	}
	return todos
}

// Len returns the number of todos
func (t ToDos) Len() int {
	return len(t)
}

// Less compares two todos by their creation time
func (t ToDos) Less(i, j int) bool {
	first := t[i]
	second := t[j]
	if first.Commit == nil {
		return true
	}
	if second.Commit == nil {
		return false
	}
	return first.Commit.Author.When.Before(second.Commit.Author.When)
}

// Swap swaps two todoss
func (t ToDos) Swap(i, j int) {
	temp := t[i]
	t[i] = t[j]
	t[j] = temp
}

// CountWithCommits returns the number of todos with an associated commit (in which that todo was added)
func (t ToDos) CountWithCommits() (count int) {
	for _, todo := range t {
		if todo.Commit != nil {
			count++
		}
	}
	return count
}
