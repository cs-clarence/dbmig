package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed default-dbmig.yaml default-summary.yaml
var DefaultsFS embed.FS

type Config struct {
	DBMig struct {
		Versioning     string `yaml:"versioning"`
		MigrationFiles string `yaml:"migration-files"`
	} `yaml:"dbmig"`
}

func InitDBMigProject(c Config) {
	file, err := os.Create("dbmig.yaml")
	if err != nil {
		log.Fatalf("Error when trying to create dbmig.yaml file: %v", err)
	}
	defaultDBMig, _ := DefaultsFS.Open("default-dbmig.yaml")

	buff := make([]byte, 1000)
	defaultDBMig.Read(buff)
	_, err = file.Write(buff)

	if err != nil {
		log.Fatalf("Error when writing to dbmig.yaml: %v", err)
	}
	file.Close()
}

type Migration struct {
	Name    string `yaml:"name"`
	Version uint   `yaml:"version"`
}

type MigrationFilesSummary struct {
	Summary struct {
		LatestVersion uint        `yaml:"latest-version"`
		Migrations    []Migration `yaml:"migrations"`
	} `yaml:"summary"`
}

func CreateNewMigration(name string, c Config) {
	yamlFP := filepath.Join(c.DBMig.MigrationFiles, "/summary.yaml")
	sumFile, err := os.Open(yamlFP)
	if err != nil {
		sumFile, err = os.Create(yamlFP)
	}

	buff := make([]byte, 1000)

	sumFile.Read(buff)

	sb := strings.Builder{}

	sum := MigrationFilesSummary{}

	switch c.DBMig.Versioning {
	case "serialint":
		sb.WriteString(strconv.Itoa(int(sum.Summary.LatestVersion)))
	case "timestamp":
		now := time.Now().UTC()
		sb.WriteString(now.Format("200601021504"))
		sb.WriteString(fmt.Sprintf("%-9d", now.Nanosecond()))
	default:
		log.Fatalf("Invalid versioning value, acceptable are timestamp or serialint")
	}
	sb.WriteString("_" + name + "_")
}

func main() {
}
