// +build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func init() {
	err := exec.Command("go", "build", "-o", "testdata/stub-agent", "testdata/stub-agent.go").Run()
	if err != nil {
		panic(err)
	}
}

func TestSupervise(t *testing.T) {
	sv := &supervisor{
		prog: "testdata/stub-agent",
	}
	sv.start()
	sv.stop(os.Interrupt)
	err := sv.wait()

	if err == nil {
		t.Errorf("something went wrong")
	}
}

func TestSuperviseReload(t *testing.T) {
	sv := &supervisor{
		prog: "testdata/stub-agent",
	}
	sv.start()
	ch := make(chan os.Signal)
	go sv.handleSignal(ch)
	done := make(chan error)
	go func() {
		done <- sv.wait()
	}()
	oldPid := sv.cmd.Process.Pid
	if !existsPid(oldPid) {
		t.Errorf("process doesn't exist")
	}
	ch <- syscall.SIGHUP
	time.Sleep(time.Second)
	newPid := sv.cmd.Process.Pid
	if oldPid == newPid {
		t.Errorf("reload failed")
	}
	if existsPid(oldPid) {
		t.Errorf("old process isn't terminated")
	}
	if !existsPid(newPid) {
		t.Errorf("new process doesn't exist")
	}

	ch <- syscall.SIGTERM
	err := <-done
	if err == nil {
		t.Errorf("something went wrong")
	}
	if newPid != sv.cmd.Process.Pid {
		t.Errorf("something went wrong")
	}
	if existsPid(newPid) {
		t.Errorf("child process isn't terminated")
	}
}
