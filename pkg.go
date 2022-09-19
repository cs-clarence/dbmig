package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cs-clarence/dbmig/defaults"
	"github.com/go-yaml/yaml"
)

var ErrMigrationNameUsed = errors.New("migration name is already used")

type Config struct {
	DBMig struct {
		Versioning     string `yaml:"versioning"`
		MigrationFiles string `yaml:"migration-files"`
	} `yaml:"dbmig"`
}

type MigrationFilesSummary struct {
	Summary struct {
		LatestVersion uint64      `yaml:"latest-version"`
		Migrations    []Migration `yaml:"migrations"`
	} `yaml:"summary"`
}

type Migration struct {
	Name      string `yaml:"name"`
	Version   uint64 `yaml:"version"`
	CreatedAt string `yaml:"created_at"`
}

func PathExists(f string) bool {
	if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func FileReadToEnd(file fs.File) ([]byte, error) {
	buffer := new(bytes.Buffer)
	_, err := io.Copy(buffer, file)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func InitDBMigProject(path string) error {
	fp := filepath.Join(path, "/dbmig.yaml")
	// Does file already exist?
	if PathExists(fp) {
		// If yes, then the project is already initialized, so exit early
		fmt.Println("Project is already initialized")
		return nil
	}

	file, err := os.Create("dbmig.yaml")
	if err != nil {
		return err
	}
	defer file.Close()
	defaultDBMig, _ := defaults.FS.Open("default-dbmig.yaml")

	buff, err := FileReadToEnd(defaultDBMig)
	if err != nil {
		return err
	}

	_, err = file.Write(buff)

	if err != nil {
		return err
	}
	return nil
}

func CreateNewMigration(name string, c Config) (*Migration, *MigrationFilesSummary, error) {
	if !PathExists(c.DBMig.MigrationFiles) {
		err := os.Mkdir(c.DBMig.MigrationFiles, os.ModePerm)
		if err != nil {
			return nil, nil, err
		}
	}

	yamlFP := filepath.Join(c.DBMig.MigrationFiles, "/summary.yaml")
	sumFile, err := os.Open(yamlFP)
	if err != nil {
		sumFile, err = os.Create(yamlFP)
		if err != nil {
			return nil, nil, err
		}

		defSum, err := defaults.FS.Open("default-summary.yaml")
		if err != nil {
			return nil, nil, err
		}
		defer defSum.Close()

		_, err = io.Copy(sumFile, defSum)
		if err != nil {
			return nil, nil, err
		}
	}
	defer sumFile.Close()

	buff, err := FileReadToEnd(sumFile)
	if err != nil {
		return nil, nil, err
	}

	sum := &MigrationFilesSummary{}
	err = yaml.Unmarshal(buff, sum)
	if err != nil {
		return nil, nil, err
	}

	// check if migration name is already used
	for _, mig := range sum.Summary.Migrations {
		if mig.Name == name {
			return nil, nil, ErrMigrationNameUsed
		}
	}

	var version uint64

	utcNow := time.Now().UTC()
	switch c.DBMig.Versioning {
	case "serialint":
		version = sum.Summary.LatestVersion
		version++
	case "timestamp":
		timeStr := utcNow.Format("200601021504")
		us := fmt.Sprintf("%-6d", utcNow.Nanosecond()/1000)
		timeStr += us
		version, _ = strconv.ParseUint(timeStr, 10, 64)
	default:
		return nil, nil, fmt.Errorf(
			"dbmig.yaml: Invalid versioning value, acceptable are \"timestamp\" or \"serialint\"",
		)
	}

	m := Migration{
		Name:      name,
		Version:   version,
		CreatedAt: utcNow.Format(time.RFC3339Nano),
	}

	sum.Summary.LatestVersion = version
	sum.Summary.Migrations = append(sum.Summary.Migrations, m)

	return &m, sum, nil
}
