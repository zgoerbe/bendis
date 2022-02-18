package main

func doMigrate(arg2, arg3 string) error {
	checkForDB()

	tx, err := bend.PopConnect()
	if err != nil {
		exitGracefully(err)
	}
	defer tx.Close()

	// run the migration command
	switch arg2 {
	case "up":
		err := bend.RunPopMigrations(tx)
		if err != nil {
			return err
		}
	case "down":
		if arg3 == "all" {
			err := bend.PopMigrateDown(tx, -1)
			if err != nil {
				exitGracefully(err)
			}
		} else {
			err := bend.PopMigrateDown(tx, 1)
			if err != nil {
				exitGracefully(err)
			}
		}
	case "reset":
		err := bend.PopMigrateReset(tx)
		if err != nil {
			exitGracefully(err)
		}
	default:
		showHelp()
	}
	return nil
}
