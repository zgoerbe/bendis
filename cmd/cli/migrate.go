package main

func doMigrate(arg2, arg3 string) error {
	dsn := getDSN()
	// run the migration command
	switch arg2 {
	case "up":
		err := bend.MigrateUp(dsn)
		if err != nil {
			return err
		}
	case "down":
		if arg3 == "all" {
			err := bend.MigrateDownAll(dsn)
			if err != nil {
				return err
			}
		} else {
			err := bend.Steps(-1, dsn)
			if err != nil {
				return err
			}
		}
	case "reset":
		err := bend.MigrateDownAll(dsn)
		if err != nil {
			return err
		}
		err = bend.MigrateUp(dsn)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}
	return nil
}
