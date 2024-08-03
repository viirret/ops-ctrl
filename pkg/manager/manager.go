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

func (m *Manager) AddService(id string, command string, args []string, env []string, workingDir string, serviceMode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, err := service.NewService(id, command, args, env, workingDir, serviceMode)

	if err != nil {
		return fmt.Errorf("NewService returns error: %v", err)
	}
	m.services[id] = service
	return nil
}

func (m *Manager) StartService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	service, exists := m.services[name]
	if !exists {
		return nil
	}
	return service.Start()
}

func (m *Manager) StopService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	service, exists := m.services[name]
	if !exists {
		return nil
	}
	return service.Stop()
}

func (m *Manager) ServiceStatus(name string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	service, exists := m.services[name]
	if !exists {
		return "unknown"
	}
	return service.CheckStatus()
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
