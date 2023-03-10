package runner

import (
	"encoding/json"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

func CommandUseUser(cmd *exec.Cmd, username string) (*exec.Cmd, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, err
	}
	if cmd.SysProcAttr.Credential == nil {
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		}
	}
	_, err = os.Stat(u.HomeDir)
	if os.IsNotExist(err) {
		os.Mkdir(u.HomeDir, os.FileMode(0700))
		os.Chown(u.HomeDir, int(uid), int(gid))
	}
	cmd.Dir = u.HomeDir
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, []string{
		"HOME=" + u.HomeDir,
		"USER=" + u.Username,
	}...)
	return cmd, nil
}

func GetHomeDirectory() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}

func GetCurrentUsername() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

type Status struct {
	ProblemID       string `json:"problem-id"`
	SolutionID      string `json:"solution-id"`
	EntrancePID     int    `json:"entrance-pid"`
	ProblemStoredTo string `json:"problem-stored-to"`
}

func getStatusFileName(storagePath map[string]string, user string) (string, error) {
	u := user
	var err error
	if u == "" {
		u, err = GetCurrentUsername()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(storagePath["status"], u+".judge.json"), nil
}

func GetStatus(storagePath map[string]string) (*Status, error) {
	path, err := getStatusFileName(storagePath, "")
	if err != nil {
		return nil, err
	}
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	status := new(Status)
	err = json.Unmarshal(f, status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func WriteStatus(
	storagePath map[string]string,
	problemID string, solutionID string, entrancePID int, problemStoredTo string,
	user string,
) error {
	path, err := getStatusFileName(storagePath, user)
	if err != nil {
		return err
	}
	status := &Status{
		SolutionID:      solutionID,
		ProblemID:       problemID,
		EntrancePID:     entrancePID,
		ProblemStoredTo: problemStoredTo,
	}
	f, err := json.Marshal(status)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, f, os.FileMode(0600))
	if err != nil {
		return err
	}
	return os.Chown(path, 0, 0)
}

func ClearStatus(storagePath map[string]string, user string) error {
	path, err := getStatusFileName(storagePath, user)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
