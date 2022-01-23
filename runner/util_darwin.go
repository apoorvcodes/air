package runner

import (
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
)

func (e *Engine) killCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid

	if e.config.Build.SendInterrupt {
		// Sending a signal to make it clear to the process that it is time to turn off
		if err = syscall.Kill(pid, syscall.SIGINT); err != nil {
			e.mainDebug("trying to send signal failed %v", err)
			return
		}
		time.Sleep(e.config.Build.KillDelay * time.Millisecond)
	}

	// find process by pid and kill it and its children by group id
	proc, err := ps.FindProcess(pid)
	if err != nil {
		return pid, errors.Wrapf(err, "failed to find process %d", pid)
	}
	err = syscall.Kill(-proc.Pid(), syscall.SIGKILL)
	if err != nil {
		return pid, errors.Wrapf(err, "failed to kill process %d", pid)
	}

	e.mainDebug("killed process pid %d successed", pid)
	return pid, nil
}

func (e *Engine) startCmd(cmd string) (*exec.Cmd, io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	c := exec.Command("/bin/sh", "-c", cmd)
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	stdin, err := c.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err = c.Start()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return c, stdin, stdout, stderr, nil
}
