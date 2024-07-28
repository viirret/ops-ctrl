package manager

import (
	"bytes"
    "log"
    "os/exec"
	"time"
	"sync"
    "fmt"
    "os"
    "syscall"
)

type ServiceStatus struct {
	State string // Current state. States:
	// running, stopped, error
	Details []string 	// Additional details or logs
	Updated time.Time 	// Last update time
}

// NewServiceStatus creates a new ServiceStatus with the given state and details
func NewServiceStatus(state string, details ...string) ServiceStatus {
	return ServiceStatus{
		State:   state,
		Details: details,
		Updated: time.Now(),
	}
}

type Service struct {
    Name string 		// Name of the service
    Command string  	// Command to execute
    Cmd *exec.Cmd 		// Exec command
	WorkingDir string 	// Working directory for the command
	Env	[]string 		// Environment variables
	OutputBuffer *bytes.Buffer	// Buffer to store command output
	Status ServiceStatus// Detailed status of the service
	mu sync.Mutex 		// Mutex to ensure thread safe operation
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status.State == "running" {
		return fmt.Errorf("service %s is already running", s.Name)
	}

    // Initialize the command with the specified working directory and environment variables
	s.Cmd = exec.Command("sh", "-c", s.Command)
	s.Cmd.Dir = s.WorkingDir
	s.Cmd.Env = append(s.Cmd.Env, s.Env...)

	// Capture standard output and error
	s.OutputBuffer = &bytes.Buffer{}
	s.Cmd.Stdout = s.OutputBuffer
	s.Cmd.Stderr = s.OutputBuffer

    err := s.Cmd.Start()

	if err != nil {
		s.Status = NewServiceStatus("error", fmt.Sprintf("failed to start: %v", err))
		return fmt.Errorf("failed to start service %s: %v", s.Name, err)
	}

	s.Status = NewServiceStatus("running", fmt.Sprintf("started with PID %d", s.Cmd.Process.Pid))
	log.Printf("Service %s started with PID %d", s.Name, s.Cmd.Process.Pid)
	return nil
}

func (s *Service) Stop() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.Cmd == nil || s.Cmd.Process == nil {
        s.Status = NewServiceStatus("unknown", "process not found")
        return nil
    }

    err := s.Cmd.Process.Signal(syscall.SIGTERM) // Send SIGTERM first for graceful termination
	if err != nil {
		if err.Error() == "os: process already finished" {
			s.Status = NewServiceStatus("stopped", "process already finished")
			return nil
		}
		return err
	}

    // Wait for the process to exit
	exitErr := s.Cmd.Wait()
	if exitErr != nil {
		if _, ok := exitErr.(*os.SyscallError); ok {
			s.Status = NewServiceStatus("stopped", "error waiting for process termination")
			return exitErr
		}
	}

    s.Status = NewServiceStatus("stopped", "process terminated successfully")
	log.Printf("Service %s stopped", s.Name)
	return nil
}

func (s *Service) CheckStatus() string {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.Cmd != nil && s.Cmd.Process != nil && s.Cmd.ProcessState == nil {
        return "running"
    }
    return "stopped"
}
