package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

// Process encapsulates the execution logic
type Process struct {
	command      string
	args         []string
	env          []string
	workingDir   string
	cmd          *exec.Cmd
	outputBuffer *bytes.Buffer
	mu           sync.Mutex
}

// NewProcess initializes a new process
func NewProcess(command string, args []string, env []string, dir string) *Process {
	return &Process{
		command:    command,
		args:       args,
		env:        env,
		workingDir: dir,
	}
}

// Start initiates the process
func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Initialize the command
	p.cmd = exec.Command(p.command, p.args...)
	p.cmd.Dir = p.workingDir
	p.cmd.Env = append(p.cmd.Env, p.env...)

	p.outputBuffer = &bytes.Buffer{}
	p.cmd.Stdout = p.outputBuffer
	p.cmd.Stderr = p.outputBuffer

	err := p.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}

	return nil
}

// Stop terminates the process
func (p *Process) SignalProcess(signal os.Signal) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	err := p.cmd.Process.Signal(signal)
	if err != nil {
		return fmt.Errorf("failed to stop process: %v", err)
	}

	switch signal {
	case syscall.SIGTERM, syscall.SIGKILL:
		// Wait for process to exit
		err = p.cmd.Wait()
		if err != nil {
			return fmt.Errorf("failed to wait for process termination: %v", err)
		}
	default:
		// Handle other signals if needed
		fmt.Printf("Process received signal: %v\n", signal)
	}

	return nil
}

// Status returns the current status of the process
func (p *Process) Status() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		return "exited"
	}

	return "running"
}

// Output returns the output buffer contents
func (p *Process) Output() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.outputBuffer.String()
}
