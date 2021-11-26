package bendis

import (
	"github.com/golang-migrate/migrate/v4"
	"log"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (b *Bendis) MigrateUp(dsn string) error {
	rootPath := filepath.ToSlash(b.RootPath)
	m, err := migrate.New("file://" + rootPath + "/migrations", dsn)
	if err != nil {
		log.Println("migrate.New error:", err)
		return err
	}
	//defer m.Close()
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {

		}
	}(m)

	if err = m.Up(); err != nil {
		log.Println("Error running migration:", err)
		return err
	}
	return nil
}

func (b *Bendis) MigrateDownAll(dsn string) error {
	rootPath := filepath.ToSlash(b.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	//defer m.Close()
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {

		}
	}(m)

	if err := m.Down(); err != nil {
		return err
	}
	return nil
}

func (b *Bendis) Steps(n int, dsn string) error {
	rootPath := filepath.ToSlash(b.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	//defer m.Close()
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {

		}
	}(m)

	if err := m.Steps(n); err != nil {
		return err
	}
	return nil
}

func (b *Bendis) MigrateForce(dsn string) error {
	rootPath := filepath.ToSlash(b.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	//defer m.Close()
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {

		}
	}(m)

	if err := m.Force(-1); err != nil {
		return err
	}
	return nil
}
