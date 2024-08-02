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

// NewMode creates a Mode from a string, validating it against known modes
func NewMode(modeStr string) (Mode, error) {
	mode := Mode(modeStr)
	if !mode.IsValid() {
		return "", fmt.Errorf("invalid mode: %s", modeStr)
	}
	return mode, nil
}

type Service struct {
	Name    string        // Name of the service
	Process *Process      // Encapsulated process
	Status  ServiceStatus // Detailed status of the service
	Mode    Mode          // Mode of operation
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
