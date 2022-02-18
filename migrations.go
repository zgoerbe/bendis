package bendis

import (
	"github.com/gobuffalo/pop"
	"github.com/golang-migrate/migrate/v4"
	"log"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (b *Bendis) PopConnect() (*pop.Connection, error) {
	tx, err := pop.Connect("development")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (b *Bendis) CreatePopMigration(up, down []byte, migrationName, migrationType string) error {
	var migrationPath = b.RootPath + "/migrations"
	err := pop.MigrationCreate(migrationPath, migrationName, migrationType, up, down)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bendis) RunPopMigrations(tx *pop.Connection) error {
	var migrationPath = b.RootPath + "/migrations"

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Up()
	if err != nil {
		return err
	}
	return nil
}

func (b *Bendis) PopMigrateDown(tx *pop.Connection, steps ...int) error {
	var migrationPath = b.RootPath + "/migrations"

	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Down(step)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bendis) PopMigrateReset(tx *pop.Connection) error {
	var migrationPath = b.RootPath + "/migrations"

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Reset()
	if err != nil {
		return err
	}

	return nil
}

func (b *Bendis) MigrateUp(dsn string) error {
	rootPath := filepath.ToSlash(b.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
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
