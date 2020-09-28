package main

import (
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/gookit/color"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"path/filepath"
)

var migrateCommand *flag.FlagSet

func main() {
	migrateCommand = flag.NewFlagSet("migrate", flag.ExitOnError)
	ghApiToken := migrateCommand.String("gh-token", "", "Github API Token")
	chApiToken := migrateCommand.String("ch-token", "", "Clubhouse API Token")
	mapFile := migrateCommand.String("map", "", "Mapping Json File")

	if len(os.Args) < 2 {
		showCommandHelp()
		os.Exit(1)
	}

	// Switch for the Sub Commands
	switch os.Args[1] {
		case "migrate":
			migrateCommand.Parse(os.Args[2:])
			fmt.Println(*mapFile)

			migrate:
			ex, err := os.Executable()
			if err != nil {
				panic(err)
			}
			exPath := filepath.Dir(ex)

			migration := NewClient(exPath + "/mapping.json")
			migration.GithubToken = *ghApiToken
			migration.ClubhouseToken = *chApiToken

			repo := promptForRepoSelect(migration)
			project := promptForRepoProjectSelect(migration, repo)
			chProject := promptForProjectSelect(migration)

			fmt.Println("Is this configuration correct?")
			fmt.Println("    GitHub Project:\t", *project.Name, "(" + *project.HTMLURL + ")")
			fmt.Println(" Clubhouse Project:\t", chProject.Name)

			if yesNo("Correct Configuration") == false {
				goto migrate
			}


			mapppings := promptForMapping(migration, project, chProject)

			fmt.Println(mapppings)

		default:
			fmt.Println("Command Not Found.")
			showCommandHelp()
			os.Exit(1)
	}

}

func showCommandHelp() {
	fmt.Println("Commands:")

	color.LightCyan.Println(" migrate")
	color.Gray.Println("  Move all Github project cards to Clubhouse")
	migrateCommand.PrintDefaults()
}

func yesNo(message string) bool {
	prompt := promptui.Select{
		Label: message,
		Items: []string{"Yes", "No"},

	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result == "Yes"
}

func inRepoList(repos []github.Repository, name string) int {

	for i, repo := range repos {
		if *repo.FullName == name {
			return i
		}
	}
	return -1
}
