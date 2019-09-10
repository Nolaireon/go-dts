package dts

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//type MFiles map[string]int

type MFiles struct {
	Changes  map[string]int `json:"changes,omitempty"`
	Binaries []string       `json:"binaries,omitempty"`
}

func Init(workTree, gitDir string) (output []byte, err error) {
	var b []byte
	// init repository
	args := []string{"git", "--work-tree", workTree, "--git-dir", gitDir, "init"}
	b, err = execCmd(args)
	if err != nil {
		return
	}
	output = append(output, b...)

	// write repository config name
	args = append(args[:len(args)-1], "config", "user.name", "\"Go-DTS\"")
	b, err = execCmd(args)
	if err != nil {
		return
	}
	output = append(output, b...)

	// write repository config email
	args = append(args[:len(args)-2], "user.email", "\"bss-devautotools@megafon.ru\"")
	b, err = execCmd(args)
	if err != nil {
		return
	}
	output = append(output, b...)

	return
}

func AddNCommit(workTree, gitDir string, files []string) (output []byte, err error) {
	var b []byte
	args := []string{"git", "--work-tree", workTree, "--git-dir", gitDir, "add"}
	for i := 0; i < len(files); i++ {
		if b, err = execCmd(append(args, files[i])); err != nil {
			return
		}

		output = append(output, b...)
	}

	t := time.Now().Format(time.RFC3339)
	args = append(args[:len(args)-1], "commit", "-m", t)
	if b, err = execCmd(args); err != nil {
		return
	}

	output = append(output, b...)
	return
}

func Numstat(workTree, gitDir string) (mFiles *MFiles, err error) {
	args := []string{"git", "--work-tree", workTree, "--git-dir", gitDir, "diff", "--numstat"}
	b, err := execCmd(args)
	if err != nil {
		return
	}

	mFiles = &MFiles{
		Changes:  make(map[string]int),
		Binaries: make([]string, 0),
	}

	r := strings.NewReplacer("\\t", "\t", "\\n", "\n", "\\r\\n", "\r\n")
	output := r.Replace(string(b))
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		parts := strings.Split(lines[i], "\t")
		if len(parts) == 3 {
			deletions, err := strconv.Atoi(parts[0])
			if err != nil {
				mFiles.Binaries = append(mFiles.Binaries, parts[2])
				continue
			}

			insertions, err := strconv.Atoi(parts[1])
			if err != nil {
				mFiles.Binaries = append(mFiles.Binaries, parts[2])
				continue
			}

			mFiles.Changes[parts[2]] = deletions + insertions
		}
	}

	return
}

func execCmd(args []string) ([]byte, error) {
	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}
	return exec.Command(command, args...).Output()
}

func (mf *MFiles) Telegraf(appName string) (s []string) {
	for k, v := range mf.Changes {
		s = append(s, fmt.Sprintf("data-tracking-system,appl_name=%s,filename=%s count=%d", appName, k, v))
	}

	return
}
