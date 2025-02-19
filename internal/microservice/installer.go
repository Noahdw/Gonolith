package microservice

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Microservices struct {
	entries []Microservice
}

type Microservice struct {
	config      MicroserviceConfig
	exeFileName string
	status      string
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

// Given a zip file, extract its contents and execute the exe
func (s *Microservices) InstallMicroservice(rawzip []byte) error {
	slog.Info("Begin installing microservice...")
	tmpdir, err := os.MkdirTemp("", "microservicedir")
	if err != nil {
		return err
	}

	err = os.Chdir(tmpdir)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp(tmpdir, "microservicefile")
	if err != nil {
		return err
	}
	defer file.Close()

	binaryReader := bytes.NewReader(rawzip)
	_, err = io.Copy(file, binaryReader)
	if err != nil {
		return err
	}

	archive, err := zip.OpenReader(file.Name())
	if err != nil {
		return err
	}

	microservice := Microservice{}
	for _, f := range archive.File {
		unzippedfile, err := f.Open()
		if err != nil {
			return err
		}

		newFile, err := os.Create(f.Name)
		if err != nil {
			return err
		}

		io.Copy(newFile, unzippedfile)
		newFile.Close()

		if strings.Contains(f.Name, ".exe") {
			microservice.exeFileName = f.Name

		} else if strings.Contains(f.Name, ".toml") {
			config := parseConfig(f.Name)
			if config != nil {
				microservice.config = *config
			}
		}
	}
	microservice.status = "installed"
	slog.Info("Microservice package OK. Begin service")
	s.addMicroservice(&microservice)
	microservice.start()

	return nil
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
	m.status = "running"
	_ = cmd.Wait()
	return nil
}

func (s *Microservices) addMicroservice(info *Microservice) {
	s.entries = append(s.entries, *info)
	fmt.Printf("%#v\n", *info)
}

func parseConfig(fileName string) *MicroserviceConfig {
	var config MicroserviceConfig

	newFile, err := os.Open(fileName)
	if err != nil {
		slog.Error("Cannot open config", "service", fileName)
		return nil
	}

	raw, err := io.ReadAll(newFile)
	if err != nil {
		slog.Error("Cannot read config", "service", fileName)
		return nil
	}

	err = toml.Unmarshal(raw, &config)
	if err != nil {
		slog.Error("Invalid config", "service", fileName)
		return nil
	}

	println(config.Name, config.Version, string(raw))
	return &config
}

func (m *Microservice) GetConfig() MicroserviceConfig {
	return m.config
}
