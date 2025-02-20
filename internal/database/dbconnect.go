package database

import (
	"fmt"
	"log"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx"
)

func PgConnect(dbName string, user string, password string, host string, port string) (*sqlx.DB, error) {
	db, err := splunksqlx.Connect("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName))
	if err != nil {
		defaultDB, err := splunksqlx.Connect("postgres",
			fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				host, port, "postgres", password, "postgres"))
		if err != nil {
			return nil, errors.Wrapf(err, "default database connect")
		}
		defer defaultDB.Close()

		userExists, err := checkUserExists(defaultDB, user)
		if err != nil {
			return nil, errors.Wrap(err, "pg connect")
		}
		if !userExists {
			log.Println("Creating database user")
			if err := createUser(defaultDB, user, password); err != nil {
				return nil, errors.Wrap(err, "pg connect")
			}
		}
		dbExists, err := checkDatabaseExists(defaultDB, dbName)
		if err != nil {
			return nil, errors.Wrap(err, "pg connect")
		}
		if !dbExists {
			log.Println("Creating new database")
			if err := createDb(defaultDB, user, dbName); err != nil {
				return nil, errors.Wrap(err, "pg connect")
			}
		}
		db, err = splunksqlx.Connect("postgres",
			fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				host, port, user, password, dbName))
		if err != nil {
			return nil, errors.Wrap(err, "database connect")
		}
	}
	return db, nil
}

func checkUserExists(db *sqlx.DB, user string) (userExists bool, err error) {
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname=$1);`, user).Scan(&userExists)
	if err != nil {
		return false, errors.Wrap(err, "check user exists")
	}
	return userExists, nil
}

func checkDatabaseExists(db *sqlx.DB, dbName string) (dbExists bool, err error) {
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1);`, dbName).Scan(&dbExists)
	if err != nil {
		return false, errors.Wrap(err, "check database exists")
	}
	return dbExists, nil
}

func createUser(db *sqlx.DB, user, password string) (err error) {
	_, err = db.Exec("CREATE USER " + user + " WITH ENCRYPTED PASSWORD '" + password + "';")
	if err != nil {
		return errors.Wrapf(err, "database create user")
	}
	return nil
}

func createDb(db *sqlx.DB, user, dbName string) (err error) {
	_, err = db.Exec("CREATE DATABASE " + dbName + ";")
	if err != nil {
		return errors.Wrapf(err, "database create")
	}
	if _, err = db.Exec("GRANT ALL PRIVILEGES ON DATABASE " + dbName + " TO " + user + ";"); err != nil {
		return errors.Wrapf(err, "database create user privileges")
	}
	return nil
}
