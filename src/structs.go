package main

import (
	"context"
	"fmt"
	"github.com/Masterminds/goutils"
	"github.com/google/go-github/github"
	"github.com/nhalstead/clubhouse"
	"golang.org/x/oauth2"
	"regexp"
	"strconv"
)

type Mapping struct {
	GitHubColumn github.ProjectColumn
	ClubhouseState clubhouse.WorkflowState
}

type ConfigMapping struct {
	Maps []Mapping `json:"maps"`
}

type Migration struct {
	Maps []Mapping
	ClubhouseToken string
	GithubToken string
	PerColumnCardLimit int

	Clubhouse *clubhouse.Client
	Github *github.Client

	ctx context.Context
}

func NewClient(file string) Migration {
	return Migration {
		ctx: context.Background(),
		PerColumnCardLimit: 1200,
	}
}

// Return the github Client
func (migration *Migration) githubClient() *github.Client {

	if migration.Github == nil {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: migration.GithubToken},
		)
		tc := oauth2.NewClient(migration.ctx, ts)

		migration.Github = github.NewClient(tc)
	}
	return migration.Github
}

// Return the clubhouse Client
func (migration *Migration) clubhouseClient() *clubhouse.Client {
	if migration.Clubhouse == nil {
		migration.Clubhouse = clubhouse.New(migration.ClubhouseToken)
	}

	return migration.Clubhouse
}

func (migration *Migration) GithubRepos() ([]*github.Repository, *github.Response, error){
	opt := &github.RepositoryListOptions{
		Type: "private",
		Visibility: "private",
		ListOptions: github.ListOptions{PerPage: 250},
	}
	return migration.githubClient().Repositories.List(migration.ctx, "", opt)
}

func (migration *Migration) GithubProjectsInRepo(owner string, repo string) ([]*github.Project, *github.Response, error){
	return migration.githubClient().Repositories.ListProjects(migration.ctx, owner, repo, nil)
}

func (migration *Migration) GitHubProjectColumns(project github.Project) ([]*github.ProjectColumn, *github.Response, error){
	return migration.githubClient().Projects.ListProjectColumns(migration.ctx, *project.ID, nil)
}

func (migration *Migration) GitHubProjectCards(column github.ProjectColumn) ([]*github.ProjectCard, *github.Response, error){
	// Not going down the path of pagination the entire request yet.
	options := &github.ProjectCardListOptions {}
	options.PerPage = migration.PerColumnCardLimit
	return migration.githubClient().Projects.ListProjectCards(migration.ctx, *column.ID, options)
}

func (migration *Migration) ListClubhouseProjects() ([]*clubhouse.Project, error){
	return migration.clubhouseClient().ListProjects()
}

func (migration *Migration) ListClubhouseWorkflow() ([]clubhouse.Workflow, error){
	return migration.clubhouseClient().ListWorkflows()
}

func (migration *Migration) ListClubhouseEpics() ([]clubhouse.Epic, error){
	return migration.clubhouseClient().ListEpics()
}

func (migration *Migration) GithubCardToClubhouseStory(state clubhouse.WorkflowState, project clubhouse.Project, card *github.ProjectCard) (*clubhouse.Story, error){
	note := notNullString(card.Note, "- Github Card Issue -")
	name, _ := goutils.Abbreviate(note, 64)
	var payload clubhouse.CreateStoryParams

	re := regexp.MustCompile(`^https:\/\/api\.github\.com\/repos\/(.*)\/(.*)\/issues/(.*)$`)

	// Handle if the github project card is a github issue.
	// If it is the card won't contain anything, its all in the issue.
	if card.ContentURL != nil && len(re.FindStringIndex(*card.ContentURL)) > 0 {
		var user string
		var repo string
		var numberString string
		var number int64
		isClosed := false

		matches := re.FindStringSubmatch(*card.ContentURL)
		user = matches[1]
		repo = matches[2]
		numberString = matches[3]
		number, _= strconv.ParseInt(numberString, 10, 64)

		// If the Card has a content URL then its a Github Issue
		issue, _, err := migration.Github.Issues.Get(migration.ctx, user, repo, int(number))

		if err != nil {
			fmt.Println(err)
			fmt.Println("Cant migrate card:", card.ID)
			return nil, nil
		}

		if issue.ClosedAt != nil {
			isClosed = true
		}

		htmlUrl := fmt.Sprintf("https://github.com/%s/%s/issues/%d", user, repo, issue.ID)
		payload = clubhouse.CreateStoryParams{
			Name: *issue.Title,
			Description: *issue.Body,
			ProjectID: project.ID,
			StoryType: clubhouse.StoryTypeFeature,
			Archived: isClosed,
			WorkflowStateID: &state.ID,

			CreatedAt: &card.CreatedAt.Time,
			ExternalTickets: []clubhouse.CreateExternalTicketParams{
				{
					ExternalURL: &htmlUrl,
					ExternalID: &numberString,
				},
			},
		}
	} else {
		payload = clubhouse.CreateStoryParams{
			Name: name,
			Description: note,
			ProjectID: project.ID,
			StoryType: clubhouse.StoryTypeFeature,
			Archived: false,
			WorkflowStateID: &state.ID,
			CreatedAt: &card.CreatedAt.Time,
		}
	}

	return migration.clubhouseClient().StoryCreate(payload)
}
