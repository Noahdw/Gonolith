package microservice

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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

	serviceStatus := make(chan error)
	go func() {
		done := make(chan error)
		go func() {
			done <- cmd.Wait()
		}()

		timer := time.NewTimer(time.Second * 2)
		defer timer.Stop()

		select {
		case <-timer.C:
			m.status = "running"
			serviceStatus <- nil
		case err := <-done:
			m.status = "stopped"
			serviceStatus <- fmt.Errorf("bad .exe %v", err)
			return
		}
		<-done
		slog.Info("Process exited", "service", m.exeFileName)
		m.status = "stopped"
	}()

	return <-serviceStatus
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
