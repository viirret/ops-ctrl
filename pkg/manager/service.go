package manager

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type ServiceStatus struct {
	State string // Current state. States:
	// running, stopped, error
	Details []string  // Additional details or logs
	Updated time.Time // Last update time
}

// NewServiceStatus creates a new ServiceStatus with the given state and details
func NewServiceStatus(state string, details ...string) ServiceStatus {
	return ServiceStatus{
		State:   state,
		Details: details,
		Updated: time.Now(),
	}
}

type Mode string

const (
    BinaryArgument Mode = "binary_argument"
    Name Mode = "name"
    PID = "pid"
    Other = "other"
)

func (m Mode) IsValid() bool {
    switch m {
    case BinaryArgument, Name, PID, Other:
        return true
    }
    return false
}

// NewMode creates a Mode from a string, validating it against known modes
func NewMode(modeStr string) (Mode, error) {
    mode := Mode(modeStr)
    if !mode.IsValid() {
        return "", fmt.Errorf("invalid mode: %s", modeStr)
    }
    return mode, nil
}

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
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	err := p.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to stop process: %v", err)
	}

	// Wait for process to exit
	err = p.cmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for process termination: %v", err)
	}

	return nil
}

// Status returns the current status of the process
func (p *Process) Status() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		return "stopped"
	}

	return "running"
}

// Output returns the output buffer contents
func (p *Process) Output() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.outputBuffer.String()
}

type Service struct {
	Name         string        // Name of the service
    Process      *Process      // Encapsulated process
	Status       ServiceStatus // Detailed status of the service
    Mode        Mode           // Mode of operation
}

// NewService initializes a new service
func NewService(name, command string, args []string, env []string, dir string, modeStr string) (*Service, error) {
    mode, err := NewMode(modeStr)
    if err != nil {
        return nil, err
    }

    return &Service{
        Name:    name,
        Process: NewProcess(command, args, env, dir),
        Status:  NewServiceStatus("initialized"),
        Mode:    mode,
    }, nil
}

func (s *Service) Start() error {
	if s.Status.State == "running" {
		return fmt.Errorf("service %s is already running", s.Name)
	}
    err := s.Process.Start()

	if err != nil {
		s.Status = NewServiceStatus("error", fmt.Sprintf("failed to start: %v", err))
		return fmt.Errorf("failed to start service %s: %v", s.Name, err)
	}

	s.Status = NewServiceStatus("running", fmt.Sprintf("started with PID %d", s.Process.cmd.Process.Pid))
	log.Printf("Service %s started with PID %d", s.Name, s.Process.cmd.Process.Pid)
	return nil
}

func (s *Service) Stop() error {
    err := s.Process.Stop()
    if err != nil {
        s.Status = NewServiceStatus("error", fmt.Sprintf("Failed to stop: %v", err))
        return err
    }

	s.Status = NewServiceStatus("stopped", "process terminated successfully")
	log.Printf("Service %s stopped", s.Name)
	return nil
}

func (s *Service) CheckStatus() string {
    return s.Process.Status()
}
