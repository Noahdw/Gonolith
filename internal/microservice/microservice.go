package microservice

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Microservice struct {
	config      MicroserviceConfig
	exeFileName string
	status      string
	id          string
	process     *exec.Cmd
}

func NewMicroservice() *Microservice {
	return &Microservice{
		status: "Not installed",
	}
}

type MicroserviceConfig struct {
	Name    string
	Version string
}

type MicroserviceStatusAPI struct {
	Status  string `json:"status"`
	Id      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (m *Microservice) GetStatus() MicroserviceStatusAPI {
	return MicroserviceStatusAPI{
		Status:  m.status,
		Id:      m.id,
		Name:    m.config.Name,
		Version: m.config.Version,
	}
}

func (m *Microservice) start() error {
	// Make it executable
	err := os.Chmod(m.exeFileName, 0700)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(m.exeFileName)
	if err != nil {
		return err
	}

	// Execute the file
	cmd := exec.Command(absPath)
	slog.Info("Begin executing", "service", m.exeFileName)
	err = cmd.Start()
	if err != nil {
		slog.Error("Failed to execute service", "error", err.Error())
		return err
	}
	m.process = cmd
	m.status = "running"

	go func() {
		_ = cmd.Wait()
		slog.Info("Process exited", "service", m.exeFileName)
		m.status = "stopped"
	}()

	return nil
}

func (m *Microservice) stop() error {
	if m.process == nil {
		slog.Error("No cmd for process", "service", m.exeFileName)
		return fmt.Errorf("no cmd available to stop process")
	}

	err := m.process.Process.Kill()
	if err != nil {
		return fmt.Errorf("could not kill process %v", err)
	}

	slog.Info("successfully stopped process", "service", m.exeFileName)
	return nil
}

func (m *Microservice) GetConfig() MicroserviceConfig {
	return m.config
}
