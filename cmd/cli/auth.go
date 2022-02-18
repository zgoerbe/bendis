package main

import (
	"github.com/fatih/color"
)

func doAuth() error {
	checkForDB()

	// migrations
	dbType := bend.DB.DatabaseType

	tx, err := bend.PopConnect()
	if err != nil {
		exitGracefully(err)
	}
	defer tx.Close()

	upBytes, err := templateFS.ReadFile("templates/migrations/auth_tables." + dbType + ".sql")
	if err != nil {
		exitGracefully(err)
	}

	downBytes := []byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;")

	err = bend.CreatePopMigration(upBytes, downBytes, "auth", "sql")
	if err != nil {
		exitGracefully(err)
	}

	// run migrations
	err = bend.RunPopMigrations(tx)
	if err != nil {
		exitGracefully(err)
	}

	// copy files over
	err = copyFileFromTemplate("templates/data/user.go.txt", bend.RootPath+"/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/token.go.txt", bend.RootPath+"/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/remember_token.go.txt", bend.RootPath+"/data/remember_token.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy from middleware
	err = copyFileFromTemplate("templates/middleware/auth.go.txt", bend.RootPath+"/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/middleware/auth-token.go.txt", bend.RootPath+"/middleware/auth-token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/middleware/remember.go.txt", bend.RootPath+"/middleware/remember.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/handlers/auth-handlers.go.txt", bend.RootPath+"/handlers/auth-handlers.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/mailer/password-reset.html.tmpl", bend.RootPath+"/mail/password-reset.html.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/mailer/password-reset.plain.tmpl", bend.RootPath+"/mail/password-reset.plain.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/views/login.jet", bend.RootPath+"/views/login.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/views/forgot.jet", bend.RootPath+"/views/forgot.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/views/reset-password.jet", bend.RootPath+"/views/reset-password.jet")
	if err != nil {
		exitGracefully(err)
	}

	color.Green("  - users, tokens and remember_tokens migrations created and executed")
	color.Green("  - user and token models created")
	color.Green("  - auth middleware created")
	color.Green("")
	color.Yellow("Don't forget to add user and token models in data/models.go and to add appropriate middleware to your routes!")

	return nil
}
