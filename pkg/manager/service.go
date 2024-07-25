package manager

import (
    "log"
    "os/exec"
)

type Service struct {
    Name string
    Command string
    Cmd *exec.Cmd
}

func (s *Service) Start() error {
    cmd := exec.Command("sh", "-c", s.Command)
    err := cmd.Start()
    if err != nil {
        return err
    }
    s.Cmd = cmd
    log.Printf("Service %s started with PID %d", s.Name, cmd.Process.Pid)
    return nil
}

func (s *Service) Stop() error {
    if s.Cmd == nil || s.Cmd.Process == nil {
        return nil
    }
    err := s.Cmd.Process.Kill()
    if err != nil {
        return err
    }
    log.Printf("Service %s stopped", s.Name)
    return nil
}

func (s *Service) Status() string {
    if s.Cmd != nil && s.Cmd.Process != nil && s.Cmd.ProcessState == nil {
        return "running"
    }
    return "stopped"
}
