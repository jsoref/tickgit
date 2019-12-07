package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/augmentable-dev/tickgit/pkg/comments"
	"github.com/augmentable-dev/tickgit/pkg/todos"
	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func init() {
	rootCmd.AddCommand(todosCmd)
}

var todosCmd = &cobra.Command{
	Use:   "todos",
	Short: "Print a report of current TODOs",
	Long:  `Scans a given git repository looking for any code comments with TODOs. Displays a report of all the TODO items found.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Suffix = " finding TODOs"
		s.Writer = os.Stderr
		s.Start()

		cwd, err := os.Getwd()
		handleError(err)

		dir := cwd
		if len(args) == 1 {
			dir, err = filepath.Rel(cwd, args[0])
			handleError(err)
		}

		validateDir(dir)

		r, err := git.PlainOpen(dir)
		handleError(err)

		ref, err := r.Head()
		handleError(err)

		commit, err := r.CommitObject(ref.Hash())
		handleError(err)

		foundToDos := make(todos.ToDos, 0)
		err = comments.SearchDir(dir, func(comment *comments.Comment) {
			todo := todos.NewToDo(*comment)
			if todo != nil {
				foundToDos = append(foundToDos, todo)
				s.Suffix = fmt.Sprintf(" %d TODOs found", len(foundToDos))
			}
		})
		handleError(err)

		ctx := context.Background()
		// timeout after 30 seconds
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		err = foundToDos.FindBlame(ctx, r, commit, func(commit *object.Commit, remaining int) {
			total := len(foundToDos)
			s.Suffix = fmt.Sprintf(" (%d/%d) %s: %s", total-remaining, total, commit.Hash, humanize.Time(commit.Author.When))
			if total-remaining > 5 {
				cancel()
			}
		})
		sort.Sort(&foundToDos)

		handleError(err)

		s.Stop()
		todos.WriteTodos(foundToDos, os.Stdout)
	},
}
