package service

import (
	"fmt"
	"strconv"
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

type Service struct {
	ID      string        // ID of the service
	Process *Process      // Encapsulated process
	Status  ServiceStatus // Detailed status of the service
}

// NewService initializes a new service
func NewService(id string, command string, args []string, env []string, dir string) (*Service, error) {
	return &Service{
		ID:      id,
		Process: NewProcess(command, args, env, dir),
		Status:  NewServiceStatus("initialized"),
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

	s.Status = NewServiceStatus("running", fmt.Sprintf("started with PID:%d and ID:%s", s.Process.cmd.Process.Pid, s.ID))
	fmt.Printf("Service started with PID %d and ID %s", s.Process.cmd.Process.Pid, s.ID)
	return nil
}

func (s *Service) Stop() error {
	pid := strconv.Itoa(s.GetPID())
	err := s.Process.Stop()
	if err != nil {
		s.Status = NewServiceStatus("error", fmt.Sprintf("Failed to stop: %v", err))
		return err
	}

	s.Status = NewServiceStatus("stopped", "process terminated successfully")
	fmt.Printf("Service " + pid + " stopped")
	return nil
}

func (s *Service) CheckStatus() string {
	return s.Process.Status()
}

func (s *Service) GetPID() int {
	return s.Process.cmd.Process.Pid
}
