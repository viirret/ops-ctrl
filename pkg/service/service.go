package service

import (
	"fmt"
	"log"
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
	Name           Mode = "name"
	PID            Mode = "pid"
	Other          Mode = "other"
)

func (m Mode) IsValid() bool {
	switch m {
	case BinaryArgument, Name, PID, Other:
		return true
	}
	return false
}

func CheckArguments(args []string) map[Mode]string {
	validArgs := make(map[Mode]string)

	binaryArgument := map[string]bool{
		"-b":    true,
		"--bin": true,
	}

	for i, arg := range args {
		if binaryArgument[arg] {
			validArgs[BinaryArgument] = string(arg[i+1])
		}
	}
	return validArgs
}

// NewMode creates a Mode from a string, validating it against known modes
func NewMode(modeStr string) (Mode, error) {
	mode := Mode(modeStr)
	if !mode.IsValid() {
		return "", fmt.Errorf("invalid mode: %s", modeStr)
	}
	return mode, nil
}

type Service struct {
	ID      string        // ID of the service
	Process *Process      // Encapsulated process
	Status  ServiceStatus // Detailed status of the service
	Mode    Mode          // Mode of operation
}

// NewService initializes a new service
func NewService(id string, command string, args []string, env []string, dir string, modeStr string) (*Service, error) {
	mode, err := NewMode(modeStr)
	if err != nil {
		return nil, err
	}

	return &Service{
		ID:      id,
		Process: NewProcess(command, args, env, dir),
		Status:  NewServiceStatus("initialized"),
		Mode:    mode,
	}, nil
}

func (s *Service) Start() error {
	if s.Status.State == "running" {
		return fmt.Errorf("service is already running")
	}
	err := s.Process.Start()

	if err != nil {
		s.Status = NewServiceStatus("error", fmt.Sprintf("failed to start: %v", err))
		return fmt.Errorf("failed to start service: %v", err)
	}

	s.Status = NewServiceStatus("running", fmt.Sprintf("started with PID %d", s.Process.cmd.Process.Pid))
	log.Printf("Service started with PID %d", s.Process.cmd.Process.Pid)
	return nil
}

func (s *Service) Stop() error {
	err := s.Process.Stop()
	if err != nil {
		s.Status = NewServiceStatus("error", fmt.Sprintf("Failed to stop: %v", err))
		return err
	}

	s.Status = NewServiceStatus("stopped", "process terminated successfully")
	log.Printf("Service stopped")
	return nil
}

func (s *Service) CheckStatus() string {
	return s.Process.Status()
}

func (s *Service) GetPID() int {
	return s.Process.cmd.Process.Pid
}
