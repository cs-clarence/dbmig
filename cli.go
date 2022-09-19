package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-yaml/yaml"
)

type CLI struct {
	Init InitCmd `kong:"cmd,help='Initialize the project'"`
	New  NewCmd  `kong:"cmd,help='Create a new migration'"`
}

type InitCmd struct {
	Path string `kong:"arg,optional,default='.',name='path',help='The directory of the project',type='path'"`
}

type NewCmd struct {
	MigrationName string `kong:"arg,name='migration-name',help='The name of the new migration'"`
}

func (l *InitCmd) Run() error {
	return InitDBMigProject(l.Path)
}

func (n *NewCmd) Run() error {
	config := new(Config)

	configFile, err := os.Open("dbmig.yaml")
	if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(
			os.Stderr,
			"dbmig.yaml file not found in this directory. Did you forget to 'dbmig init'?",
		)
		return nil
	}
	defer configFile.Close()

	buff, err := FileReadToEnd(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buff, config)
	if err != nil {
		return err
	}

	mig, sum, err := CreateNewMigration(n.MigrationName, *config)
	if errors.Is(err, ErrMigrationNameUsed) {
		fmt.Fprintf(os.Stderr, "%s: %v\n", n.MigrationName, err)
		return nil
	}

	upFP := filepath.Join(
		config.DBMig.MigrationFiles,
		fmt.Sprintf("%d_%s.up.sql", mig.Version, mig.Name),
	)

	downFP := filepath.Join(
		config.DBMig.MigrationFiles,
		fmt.Sprintf("%d_%s.down.sql", mig.Version, mig.Name),
	)

	sumFP := filepath.Join(config.DBMig.MigrationFiles, "summary.yaml")

	if file, err := os.Create(upFP); err != nil {
		return err
	} else {
		_, err = file.Write([]byte(fmt.Sprintf(`-- Name: %s
-- Version: %d
-- Type: Upgrade
-- Created At (UTC): %s
`, mig.Name, mig.Version, mig.CreatedAt,
		)))
		if err != nil {
			return err
		}
	}

	if file, err := os.Create(downFP); err != nil {
		return err
	} else {
		_, err = file.Write([]byte(fmt.Sprintf(`-- Name: %s
-- Version: %d
-- Type: Downgrade
-- Created At (UTC): %s
`, mig.Name, mig.Version, mig.CreatedAt,
		)))
		if err != nil {
			return err
		}
	}

	if sumFile, err := os.Create(sumFP); err == nil {
		s, err := yaml.Marshal(sum)
		if err != nil {
			return err
		}
		_, err = sumFile.Write(s)

		if err != nil {
			return err
		}
	} else {
		return nil
	}

	return nil
}
