package microservice

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Microservices struct {
	entries map[string]*Microservice
}

func NewMicroservices() *Microservices {
	return &Microservices{
		entries: make(map[string]*Microservice),
	}
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
	microservice.id = generateID()
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

func (s *Microservices) StopMicroservice(idToStop string) error {
	service, has := s.entries[idToStop]
	if !has {
		return fmt.Errorf("service not available to stop")
	}

	return service.stop()
}

func (s *Microservices) StartMicroservice(idToStart string) error {
	service, has := s.entries[idToStart]
	if !has {
		return fmt.Errorf("service not available to start")
	}

	return service.start()
}

func (s *Microservices) addMicroservice(info *Microservice) {
	s.entries[info.id] = info
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

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return "svc-" + string(b)
}
