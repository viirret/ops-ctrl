package manager

import (
	"fmt"
	"sync"

	"ops-ctrl/pkg/service"
)

type Manager struct {
	services map[string]*service.Service
	mu       sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		services: make(map[string]*service.Service),
	}
}

func (m *Manager) AddService(id string, command string, args []string, env []string, workingDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, err := service.NewService(id, command, args, env, workingDir)

	if err != nil {
		return fmt.Errorf("NewService returns error: %v", err)
	}
	m.services[id] = service
	return nil
}

func (m *Manager) StartService(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	service, exists := m.services[id]
	if !exists {
		return nil
	}
	return service.Start()
}

func (m *Manager) StopServiceWithID(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	service, exists := m.services[id]
	if !exists {
		return nil
	}
	return service.Stop()
}

func (m *Manager) StopServiceWithPID(pid int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, service := range m.services {
		if service.GetPID() == pid {
			service.Stop()
		}
	}
	return nil
}

func (m *Manager) ServiceStatusByID(id string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	service, exists := m.services[id]
	if !exists {
		return "unknown"
	}
	return service.CheckStatus()
}

func (m *Manager) ServiceStatusByPID(pid int) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, service := range m.services {
		if service.GetPID() == pid {
			return service.CheckStatus()
		}
	}
	return "unknown"
}

func (m *Manager) GetPID(id string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	service, exists := m.services[id]
	if !exists {
		return -1
	}
	return service.GetPID()
}
