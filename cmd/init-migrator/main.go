package main

import (
	"database/sql"
	"flag"
	"fmt"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

var (
	user      = flag.String("user", "", "database username")
	password  = flag.String("password", "", "database password")
	address   = flag.String("address", "", "database address")
	namespace = flag.String("namespace", "", "optional namespace to maintain different migrations sets")
)

func main() {
	if err := run(); err != nil {
		log.Fatal(errors.Details(err))
	}
}

func run() error {
	flag.Parse()

	if *user == "" || *address == "" {
		return errors.NotValidf("database credentials required")
	}

	if err := createSchema(); err != nil {
		return errors.Trace(err)
	}
	if err := createTable(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func createSchema() error {
	credentials := *user
	if *password != "" {
		credentials = fmt.Sprintf("%s:%s", *user, *password)
	}
	dsn := fmt.Sprintf("%s@tcp(%s)/information_schema?parseTime=true&charset=utf8mb4&collation=utf8mb4_bin", credentials, *address)
	log.WithField("dsn", dsn).Info("Connect to remote database")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return errors.Trace(err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	if err := db.Ping(); err != nil {
		return errors.Trace(err)
	}

	log.Info("Create migrator database")
	sql := `CREATE SCHEMA migrator`
	if *namespace != "" {
		sql = fmt.Sprintf(`CREATE SCHEMA migrator_%s`, *namespace)
	}
	if _, err := db.Exec(sql); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func createTable() error {
	credentials := *user
	if *password != "" {
		credentials = fmt.Sprintf("%s:%s", *user, *password)
	}
	migrator := "migrator"
	if *namespace != "" {
		migrator = fmt.Sprintf("migrator_%s", *namespace)
	}
	dsn := fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_bin", credentials, *address, migrator)
	log.WithField("dsn", dsn).Info("Connect to remote database")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return errors.Trace(err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)

	if err := db.Ping(); err != nil {
		return errors.Trace(err)
	}

	log.Info("Create migrations table")
	sql := `
		CREATE TABLE migrations (
		  name VARCHAR(191) NOT NULL,
		  applied DATETIME NOt NULL DEFAULT CURRENT_TIMESTAMP,
		  
		  PRIMARY KEY (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin
	`
	if _, err := db.Exec(sql); err != nil {
		return errors.Trace(err)
	}

	return nil
}
