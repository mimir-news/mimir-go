package dbutil

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mimir-news/mimir-go/logger"
	migrate "github.com/rubenv/sql-migrate"
)

var log = logger.GetDefaultLogger("pkg/dbutil").Sugar()

// Common errors
var (
	ErrMigrationsFailed = errors.New("database migrations failed")
	ErrNotConnected     = errors.New("not connected to database")
)

// Config interface for a database config.
type Config interface {
	DSN() string
	Driver() string
}

// Connect establishes and tests a new database connection.
func Connect(cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver(), cfg.DSN())
	if err != nil {
		return db, err
	}

	err = db.Ping()
	return db, err
}

// MustConnect establishes and tests a new database connection, and panics on errors.
func MustConnect(cfg Config) *sql.DB {
	db, err := Connect(cfg)
	if err != nil {
		log.Panicw("Failed to connect database", "driver", cfg.Driver(), "error", err)
	}

	return db
}

// MysqlConfig configuration info for a MySQL databse.
type MysqlConfig struct {
	Protocol         string `json:"protocol"`
	Host             string `json:"host"`
	Port             string `json:"port"`
	User             string `json:"username"`
	Password         string `json:"password"`
	Database         string `json:"database"`
	ConnectionParams string `json:"connectionParams"`
}

// DSN gets the datasource name.
func (cfg MysqlConfig) DSN() string {
	proto := cfg.Protocol
	if proto == "" {
		proto = "tcp"
	}

	port := cfg.Port
	if port == "" {
		port = "3306"
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%s)/%s", cfg.User, cfg.Password, proto, cfg.Host, port, cfg.Database)
	if cfg.ConnectionParams != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, cfg.ConnectionParams)
	}

	return dsn
}

// Driver gets the SQLite driver name.
func (cfg MysqlConfig) Driver() string {
	return "mysql"
}

// PostgresConfig configuration info for a MySQL databse.
type PostgresConfig struct {
	Host            string `json:"host,omitempty"`
	Port            string `json:"port,omitempty"`
	User            string `json:"user,omitempty"`
	Password        string `json:"password,omitempty"`
	Database        string `json:"database,omitempty"`
	SSLMode         string `json:"sslMode,omitempty"`
	BinaryParamters string `json:"binaryParamters,omitempty"`
}

// DSN gets the datasource name.
func (cfg PostgresConfig) DSN() string {
	port := cfg.Port
	if port == "" {
		port = "5432"
	}

	ssl := cfg.SSLMode
	if ssl == "" {
		ssl = "disable"
	}

	binaryParams := cfg.BinaryParamters
	if binaryParams == "" {
		binaryParams = "no"
	}

	connectionTemplate := "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s binary_parameters=%s"
	return fmt.Sprintf(connectionTemplate, cfg.Host, cfg.User, cfg.Password, cfg.Database, port, ssl, binaryParams)
}

// Driver gets the SQLite driver name.
func (cfg PostgresConfig) Driver() string {
	return "postgres"
}

// SqliteConfig configuration info for a SQLite databse.
type SqliteConfig struct {
	DriverName string `json:"driverName"`
	Name       string `json:"name"`
}

// DSN gets the datasource name.
func (cfg SqliteConfig) DSN() string {
	if cfg.Name != "" {
		return cfg.Name
	}

	return ":memory:"
}

// Driver gets the SQLite driver name.
func (cfg SqliteConfig) Driver() string {
	if cfg.DriverName != "" {
		return cfg.DriverName
	}

	return "sqlite3"
}

// Upgrade runs upgrade database mirgrations.
func Upgrade(path, driver string, db *sql.DB) error {
	return runMigrations(path, driver, db, migrate.Up)
}

// Downgrade runs downgrade database mirgrations.
func Downgrade(path, driver string, db *sql.DB) error {
	return runMigrations(path, driver, db, migrate.Down)
}

func runMigrations(path, driver string, db *sql.DB, direction migrate.MigrationDirection) error {
	directionName := migrationDirectionName(direction)
	source := &migrate.FileMigrationSource{Dir: path}
	migrate.SetTable("schema_version")

	migrations, err := source.FindMigrations()
	if err != nil {
		log.Errorw("Missing database migrations", "driver", driver, "path", path, "direction", directionName, "error", err)
		return ErrMigrationsFailed
	}

	if len(migrations) == 0 {
		log.Errorw("Missing database migrations", "driver", driver, "path", path, "direction", directionName)
		return ErrMigrationsFailed
	}

	_, err = migrate.Exec(db, driver, source, direction)
	if err != nil {
		log.Errorw("Error applying database migrations", "driver", driver, "direction", directionName, "error", err)
		return ErrMigrationsFailed
	}

	return nil
}

// Rollback rolls back a transaction and logs any errors that occured.
func Rollback(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		log.Errorw("Failed to rollback transaction", "error", err)
	}
}

// Connected checks that the client is connected to the database.
func Connected(db *sql.DB) error {
	rows, err := db.Query("SELECT 1")
	if err != nil {
		log.Errorw("DB health check failed", "error", err)
		return ErrNotConnected
	}
	defer rows.Close()
	return nil
}

func migrationDirectionName(direction migrate.MigrationDirection) string {
	if direction == migrate.Up {
		return "up"
	}

	return "down"
}
