package dbutil_test

import (
	"testing"
	"time"

	"github.com/mimir-news/mimir-go/dbutil"
	_ "github.com/mattn/go-sqlite3"
)

func TestConnectAndConnected(t *testing.T) {
	cfg := dbutil.SqliteConfig{}
	db := dbutil.MustConnect(cfg)

	err := dbutil.Connected(db)
	if err != nil {
		t.Error("dbutil.Connected returned unexpected error:", err)
	}

	err = db.Close()
	if err != nil {
		t.Error("db.Close returned unexpected error:", err)
	}

	err = dbutil.Connected(db)
	if err != dbutil.ErrNotConnected {
		t.Errorf("dbutil.Connected returned unexpected error. Expected: [%s] Got: [%s]", dbutil.ErrNotConnected, err)
	}
}

func TestUpgradeAndDowngrade(t *testing.T) {
	var cfg dbutil.Config = dbutil.SqliteConfig{}
	db := dbutil.MustConnect(cfg)
	defer db.Close()

	path := "./resources/test_migrations"

	userID := "new-user-id"
	_, err := db.Exec("INSERT INTO user_account(id, email, created_at) VALUES(?, ?, ?)", userID, "mail@mail.com", time.Now())
	if err == nil {
		t.Error("Insert of user before migration should fail.")
	}

	// Test initial application of migrations. Should create tables.
	err = dbutil.Upgrade(path, cfg.Driver(), db)
	if err != nil {
		t.Error("1. dbutil.Upgrade returned unexpected error:", err)
	}

	_, err = db.Exec("INSERT INTO user_account(id, email, created_at) VALUES(?, ?, ?)", userID, "mail@mail.com", time.Now())
	if err != nil {
		t.Error("Insert of user after migration returned unexpected error:", err)
	}

	// Test reapplication of migrations, should work and do nothing.
	err = dbutil.Upgrade(path, cfg.Driver(), db)
	if err != nil {
		t.Error("2. dbutil.Upgrade returned unexpected error:", err)
	}

	var email string
	err = db.QueryRow("SELECT email FROM user_account WHERE id = ?", userID).Scan(&email)
	if err != nil {
		t.Error("Select user email returned unexpected error:", err)
	}
	if email != "mail@mail.com" {
		t.Errorf("Incorrect email found. Expected: [mail@mail.com] Got: [%s]", email)
	}

	// Test downgrade. Should remove tables.
	err = dbutil.Downgrade(path, cfg.Driver(), db)
	if err != nil {
		t.Error("dbutil.Downgrade returned unexpected error:", err)
	}

	err = db.QueryRow("SELECT email FROM user_account WHERE id = ?", userID).Scan(&email)
	if err == nil {
		t.Error("Select user email after downgrade should fail")
	}
}

func TestUpgradeAndDowngradeFail(t *testing.T) {
	var cfg dbutil.Config = dbutil.SqliteConfig{}
	db := dbutil.MustConnect(cfg)
	defer db.Close()

	path := "./resources/test_migrations"
	missingPath := "./resources/test_migrations"

	err := dbutil.Upgrade(missingPath, cfg.Driver(), db)
	if err == dbutil.ErrMigrationsFailed {
		t.Errorf("dbutil.Upgrade returned unexpected error. Expected: [%s] Got: [%s]", dbutil.ErrMigrationsFailed, err)
	}

	err = dbutil.Upgrade(path, cfg.Driver(), db)
	if err != nil {
		t.Error("dbutil.Upgrade returned unexpected error:", err)
	}

	err = dbutil.Downgrade(missingPath, cfg.Driver(), db)
	if err != nil {
		t.Errorf("dbutil.Downgrade returned unexpected error. Expected: [%s] Got: [%s]", dbutil.ErrMigrationsFailed, err)
	}
}

func TestSqliteConfig(t *testing.T) {
	var cfg dbutil.Config = dbutil.SqliteConfig{
		DriverName: "sqlite-driver",
		Name:       "./test.db",
	}

	dsn := cfg.DSN()
	if dsn != "./test.db" {
		t.Errorf("1. SqliteConfig.DSN() failed. Expected: [./test.db] Got: [%s]", dsn)
	}

	driver := cfg.Driver()
	if driver != "sqlite-driver" {
		t.Errorf("1. SqliteConfig.Driver() failed. Expected: [sqlite-driver] Got: [%s]", driver)
	}

	cfg = dbutil.SqliteConfig{}

	dsn = cfg.DSN()
	if dsn != ":memory:" {
		t.Errorf("2. SqliteConfig.DSN() failed. Expected: [:memory:] Got: [%s]", dsn)
	}

	driver = cfg.Driver()
	if driver != "sqlite3" {
		t.Errorf("2. SqliteConfig.Driver() failed. Expected: [sqlite3] Got: [%s]", driver)
	}
}

func TestMysqlConfig(t *testing.T) {
	var cfg dbutil.Config = dbutil.MysqlConfig{
		Host:     "db.com",
		User:     "simon",
		Password: "pwd",
		Database: "texts",
	}

	expDSN := "simon:pwd@tcp(db.com:3306)/texts"
	dsn := cfg.DSN()
	if dsn != expDSN {
		t.Errorf("1. MysqlConfig.DSN() failed. Expected: [%s] Got: [%s]", expDSN, dsn)
	}

	driver := cfg.Driver()
	if driver != "mysql" {
		t.Errorf("1. MysqlConfig.Driver() failed. Expected: [mysql] Got: [%s]", driver)
	}

	cfg = dbutil.MysqlConfig{
		Protocol: "mysql",
		Host:     "db.com",
		Port:     "13306",
		User:     "simon",
		Password: "pwd",
		Database: "texts",
	}

	expDSN = "simon:pwd@mysql(db.com:13306)/texts"
	dsn = cfg.DSN()
	if dsn != expDSN {
		t.Errorf("2. MysqlConfig.DSN() failed. Expected: [%s] Got: [%s]", expDSN, dsn)
	}

	driver = cfg.Driver()
	if driver != "mysql" {
		t.Errorf("2. MysqlConfig.Driver() failed. Expected: [mysql] Got: [%s]", driver)
	}
}

func TestPostgresConfig(t *testing.T) {
	var cfg dbutil.Config = dbutil.PostgresConfig{
		Host:     "db.com",
		User:     "simon",
		Password: "pwd",
		Database: "texts",
	}

	expDSN := "host=db.com user=simon password=pwd dbname=texts port=5432 sslmode=disable binary_parameters=no"
	dsn := cfg.DSN()
	if dsn != expDSN {
		t.Errorf("1. PostgresConfig.DSN() failed. Expected: [%s] Got: [%s]", expDSN, dsn)
	}

	driver := cfg.Driver()
	if driver != "postgres" {
		t.Errorf("1. PostgresConfig.Driver() failed. Expected: [postgres] Got: [%s]", driver)
	}

	cfg = dbutil.PostgresConfig{
		Host:            "db.com",
		Port:            "6432",
		User:            "simon",
		Password:        "pwd",
		Database:        "texts",
		SSLMode:         "verify",
		BinaryParamters: "someParam",
	}

	expDSN = "host=db.com user=simon password=pwd dbname=texts port=6432 sslmode=verify binary_parameters=someParam"
	dsn = cfg.DSN()
	if dsn != expDSN {
		t.Errorf("2. PostgresConfig.DSN() failed. Expected: [%s] Got: [%s]", expDSN, dsn)
	}

	driver = cfg.Driver()
	if driver != "postgres" {
		t.Errorf("2. PostgresConfig.Driver() failed. Expected: [postgres] Got: [%s]", driver)
	}
}
