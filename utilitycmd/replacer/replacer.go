package replacer

import (
	"path/filepath"
	"strings"

	"github.com/lcpu-club/hpcjudge/common/consts"
)

type Replacer struct {
	solutionID  string
	problemID   string
	user        string
	storagePath map[string]string
	strReplacer *strings.Replacer
}

func NewReplacer(solutionID string, problemID string, user string, storagePath map[string]string) *Replacer {
	rep := &Replacer{
		solutionID:  solutionID,
		problemID:   problemID,
		user:        user,
		storagePath: storagePath,
	}
	problemPath := filepath.Join(storagePath["problem"], problemID)
	solutionPath := filepath.Join(storagePath["solution"], solutionID, consts.SolutionFileName)
	rep.strReplacer = strings.NewReplacer(
		"${solution_id}", solutionID,
		"${problem_id}", problemID,
		"${solution_path}", solutionPath,
		"${problem_path}", problemPath,
		"${system_user}", user,
	)
	return rep
}

func (r *Replacer) Replace(input string) string {
	return r.strReplacer.Replace(input)
}
