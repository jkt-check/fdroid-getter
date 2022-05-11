package main

import (
	"fmt"
	"os"
)

func init() {
	commands = append(commands, &Command{
		UsageLine: "test",
		Short:     "just for test",
		Run: func(args []string) error {

			apps, err := loadIndexes()
			if err != nil {
				fmt.Fprintf(os.Stderr, "loadIndexes error %v\n", err)
				return err
			}
			fmt.Fprintf(os.Stderr, "print app's info now\n")

			for _, app := range apps {
				/*
					if len(app.ScreenshotURLs()) > 0 {
						fmt.Fprintf(os.Stderr, "packageName: %s, name: %s, Summary: %s, Icon: %s, Description: %s, License: %s, Categories: %v, website: %s, sourceCode: %s, ScreenShotURLs: %v\n", app.PackageName, app.Name, app.Summary, app.Icon, app.Description, app.License, app.Categories, app.Website, app.SourceCode, app.ScreenshotURLs())
				*/
				if app.Icon == "" {

					fmt.Println("icon empty in app definition", app.PackageName, "so we should get the icon in the localization ", app.IconURL2())
					break
				}
			}
			return nil

		},
	})
}
