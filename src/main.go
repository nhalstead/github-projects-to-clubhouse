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

	if len(os.Args) < 2 {
		showCommandHelp()
		os.Exit(1)
	}

	// Switch for the Sub Commands
	switch os.Args[1] {
		case "migrate":
			_ = migrateCommand.Parse(os.Args[2:])

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

			mappings := promptForMapping(migration, project)

			// Run the Migration on each of the columns listed.
			// This loop is based on the total columns in GitHub's Project
			for _, mapping := range mappings {
				cards, _, _ := migration.GitHubProjectCards(mapping.GitHubColumn)

				// Reverse the order so we are able to have the old cards created first
				//  then the new cards have the highest number and order (so they will be on top)
				cards = reverseArray(cards)

				for _, card := range cards {
					name := MakeTitle(notNullString(card.Note, "- Github Card Issue -"))
					fmt.Println("Migrating: " + name)
					_, err = migration.GithubCardToClubhouseStory(mapping.ClubhouseState, *chProject, card)

					if err != nil {
						fmt.Println("ERR")
						fmt.Println(err)
					}
				}
			}
			fmt.Println("Migration Complete")

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

func notNullString(val *string, def string) string {
	if val != nil {
		return *val
	}
	return def
}

func reverseArray(arr []*github.ProjectCard) []*github.ProjectCard {
	newArr := make([]*github.ProjectCard, 0, len(arr))
	for i := len(arr)-1; i >= 0; i-- {
		newArr = append(newArr, arr[i])
	}
	return newArr
}