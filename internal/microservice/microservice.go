package microservice

import (
	"bytes"
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

	// Add stdout/stderr capture for better diagnostics
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

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
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
			close(done)
		}()

		timer := time.NewTimer(time.Second * 2)
		defer timer.Stop()

		select {
		case <-timer.C:
			// Service ran for at least 2 seconds, consider it stable
			m.status = "running"
			serviceStatus <- nil
		case err := <-done:
			// Service exited quickly, that's an error
			m.status = "stopped"
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					// Get the actual exit code
					serviceStatus <- fmt.Errorf("bad .exe exit status %d: stdout: %s, stderr: %s",
						exitErr.ExitCode(), stdout.String(), stderr.String())
				} else {
					serviceStatus <- fmt.Errorf("bad .exe: %v, stdout: %s, stderr: %s",
						err, stdout.String(), stderr.String())
				}
			} else {
				serviceStatus <- fmt.Errorf("service exited unexpectedly with success code, stdout: %s, stderr: %s",
					stdout.String(), stderr.String())
			}
			return
		}

		// Wait for eventual termination
		err := <-done
		if err != nil {
			slog.Error("Process exited with error", "service", m.exeFileName, "error", err)
		} else {
			slog.Info("Process exited", "service", m.exeFileName)
		}

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
