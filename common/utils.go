package common

import (
	"os/exec"
	"os/user"
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
	cmd.Env = append(cmd.Env, []string{
		"HOME=" + u.HomeDir,
		"USER=" + u.Username,
	}...)
	return cmd, nil
}
