package main

import (
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/gookit/color"
	"github.com/jnormington/clubhouse-go"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func promptForRepoSelect(migration Migration) github.Repository {

	start:
		if yesNo("Enter manually") {
			// Manual Input
			prompt:
				repoName := promptui.Prompt{
					Label: "Repo (owner/repo)",
				}
				result, _ := repoName.Run()

				if result == "q" || result == "" {
					goto start
				}

				s := strings.Split(result, "/")

				if len(s) != 2 {
					fmt.Println("Format like the following:")
					fmt.Println(" {{owner}}/{{repo}}")
					goto prompt
				}

				repo, _, err := migration.githubClient().Repositories.Get(migration.ctx, s[0], s[1])

				if err != nil {
					fmt.Println(err)
					goto prompt
				}

				return *repo
		}

	repos, _, _ := migration.GithubRepos()

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F449 {{ .FullName }}",
		Inactive: "   {{ .FullName }}",
		Selected: "\U00002714 {{ .FullName }}",
		Details: "\n--------- Repo ----------\n Name:\t{{ .FullName }}\n Description:\t{{ .Description }}",
	}

	prompt := promptui.Select{
		Label:     "Github Repository Selection",
		Items:     repos,
		Templates: templates,
		Size:      15,
	}

	i, _, _ := prompt.Run()

	return *repos[i]
}

func promptForRepoProjectSelect(migration Migration, repo github.Repository) github.Project {
	projects, _, _ := migration.GithubProjectsInRepo(*repo.Owner.Login, *repo.Name)

	if len(projects) == 1 {
		// Auto Select First Project
		fmt.Println("\U00002714 Github Project: " + *projects[0].Name)
		return *projects[0]
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F449 {{ .Name }}",
		Inactive: "   {{ .Name }}",
		Selected: "\U00002714 Github Project: {{ .Name }}",
		Details: "\n--------- Project ----------\n ID:\t{{ .ID }}\n Description:\t{{ .Name }}",
	}

	prompt := promptui.Select{
		Label:     "Github Repository Project Selection",
		Items:     projects,
		Templates: templates,
		Size:      15,
	}

	i, _, _ := prompt.Run()

	return *projects[i]
}

func promptForProjectSelect(migration Migration) clubhouse.Project {
	projects, _ := migration.ClubhouseProjects()

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F449 {{ .Name }}",
		Inactive: "   {{ .Name }}",
		Selected: "\U00002714 Clubhouse Project {{ .Name }}",
		Details: "\n--------- Project ----------\n Name:\t{{ .Name }}\n Description:\t{{ .Description }}",
	}

	prompt := promptui.Select{
		Label:     "Clubhouse Project Selection",
		Items:     projects,
		Templates: templates,
		Size:      15,
	}

	i, _, _ := prompt.Run()

	return projects[i]
}

func promptForMapping(migration Migration, project github.Project, chProject clubhouse.Project) []Mapping {
	var mapping []Mapping
	var toMap []*github.ProjectColumn
	var workflow clubhouse.Workflow
	
	projectColumns, _, _ := migration.GitHubProjectColumns(project)
	projectWorkflows, _ := migration.ClubhouseWorkflow()

	if len(projectWorkflows) == 1 {
		// Auto Select First Workflow
		fmt.Println("\U00002714 Clubhouse Workflow: " + projectWorkflows[0].Name)
		workflow = projectWorkflows[0]
	} else {
		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U000027A1 {{ .Name }}",
			Inactive: "   {{ .Name }}",
			Selected: "\U000027A1 Clubhouse Workflow: {{ .Name }}",
			Details: "\n--------- Project ----------\n Name:\t{{ .Name }}",
		}

		prompt := promptui.Select{
			Label:     "Select Clubhouse Workflow",
			Items:     projectWorkflows,
			Templates: templates,
			Size:      15,
		}

		m, _, _ := prompt.Run()
		workflow = projectWorkflows[m]
	}

	selectCols:
	for _, column := range projectColumns {
		if yesNo("Map? " + *column.Name) {
			toMap = append(toMap, column)
		}
	}

	if len(toMap) == 0 {
		fmt.Println("Nothing to map from Github to Clubhouse!")
		goto selectCols
	}

	// Map the Selected GitHub Columns to the Clubhouse Workflow States
	for i, column := range toMap {
		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U000027A1 {{ .Name }}",
			Inactive: "   {{ .Name }}",
			Selected: "\U000027A1 Github Project Column \"" + *column.Name + "\" maps to: {{ .Name }}",
			Details: "\n--------- Project ----------\n Name:\t{{ .Name }}\n Direct Link:\t{{ .URL }}",
		}

		prompt := promptui.Select{
			Label:     "Github Project Column \"" + *column.Name + "\" maps to what Clubhouse State",
			Items:     workflow.States,
			Templates: templates,
			Size:      15,
		}

		k, _, _ := prompt.Run()
		mapping = append(mapping, Mapping{
			GitHubColumn:   *projectColumns[i],
			ClubhouseState: workflow.States[k],
		})
	}

	return mapping
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
