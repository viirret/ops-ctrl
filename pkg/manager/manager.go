package manager

import (
	"sync"
	"fmt"
)

type Manager struct {
	services map[string]*Service
	mu       sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		services: make(map[string]*Service),
	}
}

func (m *Manager) AddService(name, command string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, err := NewService(name, command , []string{"exampleargs"}, []string{"exampleenv"}, "/", "binary_argument")

	if err != nil {
		return fmt.Errorf("NewService returns error: %v", err)
	} 
    m.services[name] = service
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
