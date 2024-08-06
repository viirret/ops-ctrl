package manager

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"ops-ctrl/pkg/config"
	"ops-ctrl/pkg/service"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// init seeds the random number generator with the current time.
func init() {
	rand.NewSource(time.Now().UnixNano())
}

type Manager struct {
	services map[string]*service.Service
	mu       sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		services: make(map[string]*service.Service),
	}
}

func (m *Manager) RandomID(length int) string {
	if length <= 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		randomChar := charset[rand.Intn(len(charset))]
		sb.WriteByte(randomChar)
	}

	for _, service := range m.services {
		if service.ID == sb.String() {
			return m.RandomID(length)
		}
	}

	return sb.String()
}

func (m *Manager) RunAutostart() {
	applications := config.GetConfig().Autostart
	for _, app := range applications {
		id := m.RandomID(10)
		m.AddService(id, app, []string{}, []string{}, "/")
		err := m.StartService(id)
		if err != nil {
			log.Fatal("Failed to start service error: ", err)
		}
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
