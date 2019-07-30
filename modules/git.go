package modules

import (
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"io/ioutil"
	"path/filepath"
	"time"
)

//Init with gitdir outside of worktree
func Init(worktree, gitdir, instance string) (w *git.Worktree, err error) {
	wt := osfs.New(worktree)
	gd := osfs.New(gitdir)
	gd, err = gd.Chroot(instance)
	if err != nil {
		return nil, err
	}

	s := filesystem.NewStorage(gd, cache.NewObjectLRUDefault())

	r, err := git.Init(s, wt)
	if err != nil {
		return nil, err
	}

	w, err = r.Worktree()
	if err != nil {
		return nil, err
	}
	return w, nil
}

// Add all accessible files and commit
func Commit(worktree string, files []string, w *git.Worktree) (err error) {
	for i := 0; i < len(files); i++ {
		relPath, _ := filepath.Rel(worktree, files[i])
		if relPath == ".git" {
			continue
		}
		_, err := w.Add(relPath)
		if err != nil {
			return err
		}
		fmt.Printf("added: %s\n", relPath)
	}
	_, err = w.Commit(time.Now().Format(time.RFC822), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "data-tracking-system",
			Email: "bss-devautotools@megafon.ru",
			When:  time.Now(),
		},
	})

	return err
}

func Open(worktree, gitdir, instance string) (*git.Repository, error) {
	wt := osfs.New(worktree)
	gd := osfs.New(gitdir)
	gd, err := gd.Chroot(instance)
	if err != nil {
		return nil, err
	}
	s := filesystem.NewStorage(gd, cache.NewObjectLRUDefault())

	repo, err := git.Open(s, wt)

	return repo, err
}

func Diff(repo *git.Repository) (diffs map[string]int, err error) {
	var modified []string
	diffs = make(map[string]int, 0)
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	for k, _ := range status {
		modified = append(modified, k)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commits, _ := repo.Log(&git.LogOptions{From: ref.Hash()})
	commit, err := commits.Next()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(modified); i++ {
		prevFile, err := commit.File(modified[i])
		if err != nil {
			return nil, err
		}

		prevFileContent, err := prevFile.Contents()

		curFile, err := wt.Filesystem.Open(modified[i])
		if err != nil {
			return nil, err
		}

		curFileBuffer, err := ioutil.ReadAll(curFile)
		curFileContent := string(curFileBuffer)

		dmp := diffmatchpatch.New()
		fDiff := dmp.DiffMain(prevFileContent, curFileContent, false)
		for _, v := range fDiff {
			if v.Type != diffmatchpatch.DiffEqual {
				diffs[modified[i]] += 1
				//fmt.Println(v.Type.String(), v.Text)
			}
		}
	}
	if len(diffs) == 0 {
		return nil, nil
	}
	return diffs, nil
}
