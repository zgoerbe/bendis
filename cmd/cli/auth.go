package main

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

func doAuth() error {
	// migrations
	dbType := bend.DB.DatabaseType
	fileName := fmt.Sprintf("%d_create_auth_table", time.Now().UnixMicro())
	//rootPath := filepath.ToSlash(bend.RootPath)
	upFile := bend.RootPath + "/migrations/" + fileName + ".up.sql"
	downFile := bend.RootPath + "/migrations/" + fileName + ".down.sql"

	err := copyFileFromTemplate("templates/migrations/auth_tables."+dbType+".sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;"), downFile)
	if err != nil {
		exitGracefully(err)
	}

	// run migrations
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	// copy files over
	err = copyFileFromTemplate("templates/data/user.go.txt", bend.RootPath + "/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/data/token.go.txt", bend.RootPath + "/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy from middleware
	err = copyFileFromTemplate("templates/middleware/auth.go.txt", bend.RootPath + "/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFileFromTemplate("templates/middleware/auth-token.go.txt", bend.RootPath + "/middleware/auth-token.go")
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