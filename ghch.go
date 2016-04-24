package ghch

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
)

type Ghch struct {
	RepoPath string
}

func (gh *Ghch) Cmd(argv ...string) (string, error) {
	arg := []string{"-C", gh.RepoPath}
	arg = append(arg, argv...)
	cmd := exec.Command("git", arg...)
	cmd.Env = append(os.Environ(), "LANG=C")

	var b bytes.Buffer
	cmd.Stdout = &b
	err := cmd.Run()
	return b.String(), err
}

var verReg = regexp.MustCompile(`^v?[0-9]+(?:\.[0-9]+){0,2}$`)

func (gh *Ghch) Versions() []string {
	out, _ := gh.Cmd("tag")
	rawTags := strings.Split(out, "\n")
	var versions []*semver.Version
	for _, tag := range rawTags {
		t := strings.TrimSpace(tag)
		if verReg.MatchString(t) {
			v, _ := semver.NewVersion(t)
			versions = append(versions, v)
		}
	}
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	var vers = make([]string, len(versions))
	for i, v := range versions {
		vers[i] = v.Original()
	}
	return vers
}

var prMergeReg = regexp.MustCompile(`^[a-f0-9]{7} Merge pull request #([0-9]+) from`)

func (gh *Ghch) MergedPRNums(argv ...string) (nums []string) {
	var from, to string
	if len(argv) > 0 {
		from = argv[0]
	}
	if len(argv) > 1 {
		from = argv[1]
	}
	if from == "" {
		vers := gh.Versions()
		if len(vers) < 1 {
			return
		}
		from = vers[0]
	}
	revisionRange := fmt.Sprintf("%s...%s", from, to)
	out, err := gh.Cmd("log", revisionRange, "--merges", "--oneline")
	if err != nil {
		return
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if matches := prMergeReg.FindStringSubmatch(line); len(matches) > 1 {
			nums = append(nums, matches[1])
		}
	}
	return
}
