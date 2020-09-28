package main

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/jnormington/clubhouse-go"
	"golang.org/x/oauth2"
)

type Mapping struct {
	ClubhouseState clubhouse.State
	GitHubColumn github.ProjectColumn
}

type ConfigMapping struct {
	Maps []Mapping `json:"maps"`
}

type Migration struct {
	Maps []Mapping
	ClubhouseToken string
	GithubToken string

	Clubhouse *clubhouse.Clubhouse
	Github *github.Client

	ctx context.Context
}

func NewClient(file string) Migration {
	return Migration {
		ctx: context.Background(),
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
func (migration *Migration) clubhouseClient() *clubhouse.Clubhouse {
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
	return migration.githubClient().Projects.ListProjectCards(migration.ctx, *column.ID, nil)
}

func (migration *Migration) ClubhouseProjects() ([]clubhouse.Project, error){
	return migration.clubhouseClient().ListProjects()
}

func (migration *Migration) ClubhouseWorkflow() ([]clubhouse.Workflow, error){
	return migration.clubhouseClient().ListWorkflow()
}

func (migration *Migration) ClubhouseEpics() ([]clubhouse.Epic, error){
	return migration.clubhouseClient().EpicList()
}

const (
	ClubhouseBug = "bug"
	ClubhouseChore = "chore"
	ClubhouseFeature = "feature"
)

func (migration *Migration) ClubhouseProjectStates(story clubhouse.CreateStory) (clubhouse.Story, error){
	return migration.clubhouseClient().CreateStory(story)
}

func (migration *Migration) GithubCardToClubhouseStory(state clubhouse.State, project *clubhouse.Project, card *github.ProjectCard) (clubhouse.Story, error){
	return migration.clubhouseClient().CreateStory(clubhouse.CreateStory{
		CreatedAt: &card.CreatedAt.Time,
		Name: *card.Note,
		ProjectID: project.ID,
		ExternalID: *card.ContentURL,
		WorkflowStateID: state.ID,
	})
}
