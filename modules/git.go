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
	"log"
	"path/filepath"
	"strings"
	"time"
)

// Init initialize git repo with git-dir outside of the work-tree
func Init(workTree, dtsDir, instance string) (w *git.Worktree, err error) {
	wt := osfs.New(workTree)
	gd := osfs.New(dtsDir)
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

// Commit add all accessible files and commit them
func Commit(workTree string, files []string, w *git.Worktree) (err error) {
	for i := 0; i < len(files); i++ {
		relPath, _ := filepath.Rel(workTree, files[i])
		if relPath == ".git" {
			continue
		}

		_, err := w.Add(relPath)
		if err != nil {
			return err
		}

		log.Printf("added: %s\n", relPath)
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

// Open call git.Open and return repository structure
func Open(workTree, gitDir string) (*git.Repository, error) {
	wt := osfs.New(workTree)
	gd := osfs.New(gitDir)
	s := filesystem.NewStorage(gd, cache.NewObjectLRUDefault())

	repo, err := git.Open(s, wt)

	return repo, err
}

// Diff retrieve list of modified files(call git.St
func Diff(repo *git.Repository) (diffs MFiles, err error) {
	var modified []string
	//replacer := strings.NewReplacer("\n", "\\n", "\t", "\\t", "\r\n", "\\r\\n")
	diffs = make(MFiles, 0)
	wt, err := repo.Worktree()
	if err != nil {
		return
	}

	status, err := wt.Status()
	if err != nil {
		return
	}

	for k := range status {
		modified = append(modified, k)
	}

	ref, err := repo.Head()
	if err != nil {
		return
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return
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

		var diff []Difference
		var start, end int
		var count int
		for _, v := range fDiff {
			end += len(v.Text)
			diff = append(diff, Difference{Diff: v, Start: start, End: end})

			if v.Type != diffmatchpatch.DiffEqual {
				count += 1
			}
			//log.Printf("%d) %s (%d:%d)", idx+1, v.Type, start, end)
			//log.Printf("[%s]", replac\er.Replace(v.Text))

			start = end
		}

		diffs[modified[i]] = Diffs{Diffs: diff, Count: count}
	}

	return
}

func (mf *MFiles) Telegraf(applName string) string {
	var strs []string
	for key, value := range *mf {
		strs = append(strs, fmt.Sprintf("data-tracking-system,appl_name=%s,filename=%s count=%d", applName, key, value.Count))
	}
	return strings.Join(strs, "\n")
}

type MFiles map[string]Diffs

type Diffs struct {
	Diffs []Difference
	Count int
}

type Difference struct {
	Diff       diffmatchpatch.Diff
	Start, End int
}
