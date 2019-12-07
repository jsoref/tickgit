package todos

import (
	"context"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/format/diff"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// addedInPatch checks if a todo item exists in the tree of a given commit
func (t ToDo) addedInPatch(firstCommit, secondCommit *object.Commit) (bool, error) {
	if isAncestor, err := firstCommit.IsAncestor(secondCommit); err != nil {
		return false, err
	} else if !isAncestor {
		return false, nil
	}
	patch, err := firstCommit.Patch(secondCommit)
	if err != nil {
		return false, err
	}
	patches := patch.FilePatches()
	for _, filePatch := range patches {
		if filePatch.IsBinary() {
			continue
		}
		_, to := filePatch.Files()
		if to == nil {
			continue
		}
		if to.Path() == t.FilePath {
			chunks := filePatch.Chunks()
			for _, chunk := range chunks {
				chunkType := chunk.Type()
				if chunkType == diff.Add {
					if strings.Contains(chunk.Content(), t.Comment.String()) {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

// FindBlame sets the blame information on each todo in a set of todos
func (t ToDos) FindBlame(ctx context.Context, repo *git.Repository, from *object.Commit, cb func(commit *object.Commit, remainingToDos int)) error {
	commitIter, err := repo.Log(&git.LogOptions{
		From: from.Hash,
	})
	if err != nil {
		return err
	}
	defer commitIter.Close()

	remainingTodos := t
	prevCommit := from
	err = commitIter.ForEach(func(commit *object.Commit) error {
		if len(remainingTodos) == 0 {
			return storer.ErrStop
		}
		// if commit.NumParents() > 1 {
		// 	return nil
		// }
		if prevCommit.Hash.String() == commit.Hash.String() {
			return nil
		}
		select {
		case <-ctx.Done():
			return storer.ErrStop
		default:
			newRemainingTodos := make(ToDos, 0)
			// TODO, if the todo item was added in the initial commit, we don't handle that correctly
			for _, todo := range remainingTodos {
				exists, err := todo.addedInPatch(commit, prevCommit)
				if err != nil {
					return err
				}
				if exists { // if the todo doesn't exist in this commit, it was added in the previous commit (previous wrt the iterator, more recent in time)
					todo.Commit = prevCommit
				} else { // if the todo does exist in this commit, add it to the new list of remaining todos
					newRemainingTodos = append(newRemainingTodos, todo)
				}
			}
			if cb != nil {
				cb(commit, len(newRemainingTodos))
			}
			prevCommit = commit
			remainingTodos = newRemainingTodos
			return nil
		}
	})
	if err != nil {
		return err
	}
	// if len(remainingTodos) != 0 {
	// 	for _, todo := range remainingTodos {
	// 		todo.Commit = prevCommit
	// 	}
	// }
	return nil
}
