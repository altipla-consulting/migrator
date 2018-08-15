package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

var (
	user      = flag.String("user", "", "database username")
	password  = flag.String("password", "", "database password")
	address   = flag.String("address", "", "database address")
	directory = flag.String("directory", "", "directory with the migration files")

	dbcache = map[string]*sql.DB{}
)

func main() {
	if err := run(); err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
}

func run() error {
	flag.Parse()

	if *user == "" || *address == "" {
		return errors.NotValidf("database credentials required")
	}
	if *directory == "" {
		return errors.NotValidf("migrations directory required")
	}

	migrations, err := fetchAppliedMigrations()
	if err != nil {
		return errors.Trace(err)
	}
	if len(migrations) > 0 {
		log.WithField("count", len(migrations)).Info("Found previously applied migrations")
	}

	files, err := ioutil.ReadDir(*directory)
	if err != nil {
		return errors.Trace(err)
	}
	var migrationFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".sql" {
			fullname := filepath.Join(*directory, file.Name())
			return errors.NotValidf("all migration files should have SQL extension: %s", fullname)
		}

		migrationFiles = append(migrationFiles, file.Name())
	}

	for i, name := range migrationFiles {
		if len(migrations) > i {
			if migrations[i] != name {
				return errors.NotValidf("inconsistent applied state found: %s != %s", migrations[i], name)
			}
		} else {
			if err := applyMigration(name); err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}

func openConnection(dbname string) (*sql.DB, error) {
	if dbcache[dbname] == nil {
		credentials := *user
		if *password != "" {
			credentials = fmt.Sprintf("%s:%s", *user, *password)
		}
		dsn := fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_bin", credentials, *address, dbname)
		log.WithField("dsn", dsn).Info("Connect to remote database")

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, errors.Trace(err)
		}

		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(0)

		if err := db.Ping(); err != nil {
			return nil, errors.Trace(err)
		}

		dbcache[dbname] = db
	}

	return dbcache[dbname], nil
}

func fetchAppliedMigrations() ([]string, error) {
	db, err := openConnection("migrator")
	if err != nil {
		return nil, errors.Trace(err)
	}

	var names []string

	rows, err := db.Query("SELECT name FROM migrations ORDER BY applied")
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, errors.Trace(err)
		}

		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Trace(err)
	}

	return names, nil
}

func flagAppliedMigration(name string) error {
	db, err := openConnection("migrator")
	if err != nil {
		return errors.Trace(err)
	}

	sql := `INSERT INTO migrations(name) VALUES (?)`
	if _, err := db.Exec(sql, name); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func applyMigration(name string) error {
	content, err := ioutil.ReadFile(filepath.Join(*directory, name))
	if err != nil {
		return errors.Trace(err)
	}
	lines := strings.Split(string(content), ";\n")

	var dbname string
	var logged bool
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matched, err := regexp.MatchString("^USE [a-z0-9_]+$", line)
		if err != nil {
			return errors.Trace(err)
		} else if matched {
			if dbname == "" {
				dbname = line[len("USE "):]
				continue
			} else {
				return errors.NotValidf("only one database per migration file allowed: %s", name)
			}
		}

		if dbname == "" {
			return errors.NotValidf("database not selected before running statement: %s: %s", name, line)
		}

		if !logged {
			log.WithFields(log.Fields{
				"name":     name,
				"database": dbname,
			}).Info("Apply migration")
			logged = true
		}

		db, err := openConnection(dbname)
		if err != nil {
			return errors.Trace(err)
		}

		if _, err := db.Exec(line); err != nil {
			log.WithField("statement", line).Info("Migration SQL statement failed")
			return errors.Trace(err)
		}
	}

	if err := flagAppliedMigration(name); err != nil {
		return errors.Trace(err)
	}

	log.Info("Migration applied successfully!")
	return nil
}
