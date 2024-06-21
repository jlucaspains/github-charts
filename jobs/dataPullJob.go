package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/adhocore/gronx"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jlucaspains/github-charts/db"
	"github.com/jlucaspains/github-charts/models"
)

type DataPullJob struct {
	cron           string
	ticker         *time.Ticker
	gron           *gronx.Gronx
	running        bool
	queries        db.Querier
	projects       []models.JobConfigItem
	graphqlClients map[string]graphql.Client
}

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

func NewDataPullJob(schedule string, queries db.Querier, projects []models.JobConfigItem) (*DataPullJob, error) {
	c := &DataPullJob{}
	c.gron = gronx.New()

	if schedule == "" || !c.gron.IsValid(schedule) {
		slog.Error("A valid cron schedule is required in the format e.g.: * * * * *", "cron", schedule)
		return nil, fmt.Errorf("a valid cron schedule is required")
	}

	c.cron = schedule
	c.projects = projects
	c.ticker = time.NewTicker(time.Minute)
	c.queries = queries

	slog.Info("Init DataPullJob job")

	for index, item := range c.projects {
		slog.Info("Project configuration", "index", index, "repoOwner", item.RepoOwner, "repoName", item.RepoName, "orgName", item.OrgName, "projectId", item.Project)
	}

	c.graphqlClients = make(map[string]graphql.Client)
	for _, project := range projects {
		httpClient := http.Client{
			Transport: &authedTransport{
				key:     project.Token,
				wrapped: http.DefaultTransport,
			},
		}
		graphqlClient := graphql.NewClient("https://api.github.com/graphql", &httpClient)
		c.graphqlClients[project.GetUniqueName()] = graphqlClient
	}

	return c, nil
}

func (c *DataPullJob) Start() {
	c.running = true
	slog.Info("Started DataPullJob job", "cron", c.cron)

	go func() {
		for range c.ticker.C {
			c.tryExecute()
		}
	}()
}

func (c *DataPullJob) Stop() {
	c.running = false

	if c.ticker != nil {
		c.ticker.Stop()
	}
}

func (c *DataPullJob) tryExecute() {
	due, _ := c.gron.IsDue(c.cron, time.Now().Truncate(time.Minute))

	slog.Info("tryExecute job", "isDue", due)

	if due {
		c.execute()
	}
}

func (c *DataPullJob) execute() {
	for _, project := range c.projects {
		projectId, _ := strconv.Atoi(project.Project)
		var projectFields *ProjectFields
		var err error
		if project.OrgName != "" {
			slog.Info("Data pull job started", "orgName", project.OrgName, "projectId", projectId)
			var orgProject *getOrganizationProjectResponse
			orgProject, err = getOrgProject(c.graphqlClients[project.GetUniqueName()], project.OrgName, projectId)

			if err == nil {
				projectFields = &orgProject.Organization.ProjectV2.ProjectFields
			}
		} else {
			slog.Info("Data pull job started", "repoOwner", project.RepoOwner, "repoName", project.RepoName, "projectId", projectId)
			var repoProject *getRepositoryProjectResponse
			repoProject, err = getRepoProject(c.graphqlClients[project.GetUniqueName()], project.RepoOwner, project.RepoName, projectId)

			if err == nil {
				projectFields = &repoProject.Repository.ProjectV2.ProjectFields
			}
		}

		if err == nil {
			project := parseProjectInformation(projectFields)
			saveProjectInformation(project, c.queries)
		} else {
			slog.Error("Error fetching project information", "error", err)
		}
	}
}

func saveProjectInformation(project *models.Project, queries db.Querier) error {
	ctx := context.Background()
	dbProject, err := queries.UpsertProject(ctx, db.UpsertProjectParams{
		GhID: project.Id,
		Name: project.Title,
	})

	if err != nil {
		slog.Error("Error on UpsertProject", "error", err)
		return err
	}

	iterationsMap := make(map[string]int32)
	for _, iteration := range project.Iterations {
		dbIteration, err := queries.UpsertIteration(ctx, db.UpsertIterationParams{
			GhID:      iteration.Id,
			Name:      iteration.Title,
			ProjectID: dbProject.ID,
			StartDate: pgtype.Date{Time: iteration.StartDate, Valid: true},
			EndDate:   pgtype.Date{Time: iteration.EndDate, Valid: true},
		})

		iterationsMap[iteration.Id] = dbIteration.ID

		if err != nil {
			slog.Error("Error on UpsertIteration", "error", err)
			return err
		}
	}

	for _, status := range project.Statuses {
		_, err := queries.UpsertWorkItemStatus(ctx, status)

		if err != nil {
			slog.Error("Error on UpserWorkItemStatus", "error", err)
			return err
		}
	}

	for _, issue := range project.Issues {
		now := time.Now().Truncate(24 * time.Hour)
		today := now.UTC()
		iterationId, iterationIdOk := iterationsMap[issue.IterationId]

		_, err := queries.UpsertWorkItem(ctx, db.UpsertWorkItemParams{
			GhID:           issue.Id,
			ChangeDate:     pgtype.Date{Time: today, Valid: true},
			Name:           issue.Title,
			Effort:         pgtype.Int4{Int32: int32(issue.Effort), Valid: true},
			RemainingHours: pgtype.Int4{Int32: int32(issue.RemainingHours), Valid: true},
			Status:         pgtype.Text{String: issue.Status, Valid: true},
			IterationID:    pgtype.Int4{Int32: iterationId, Valid: iterationIdOk},
			ProjectID:      dbProject.ID,
		})

		if err != nil {
			slog.Error("Error on UpserWorkItem", "error", err)
			return err
		}
	}

	return nil
}

func getOrgProject(graphqlClient graphql.Client, orgName string, projectId int) (*getOrganizationProjectResponse, error) {
	hasNextPage := true
	isFirstPage := true
	cursor := ""
	orgProject := &getOrganizationProjectResponse{}

	for hasNextPage {
		result, err := getOrganizationProject(context.Background(), graphqlClient, orgName, projectId, 5, cursor)
		if err != nil {
			return nil, err
		}

		if isFirstPage {
			orgProject = result
		} else {
			orgProject.Organization.ProjectV2.Items.Nodes = append(orgProject.Organization.ProjectV2.Items.Nodes, result.Organization.ProjectV2.Items.Nodes...)
		}

		isFirstPage = false
		hasNextPage = orgProject.Organization.ProjectV2.Items.PageInfo.HasNextPage
		cursor = orgProject.Organization.ProjectV2.Items.PageInfo.EndCursor
	}

	return orgProject, nil
}

func getRepoProject(graphqlClient graphql.Client, repoOwner string, repoName string, projectId int) (*getRepositoryProjectResponse, error) {
	hasNextPage := true
	isFirstPage := true
	cursor := ""
	repoProject := &getRepositoryProjectResponse{}

	for hasNextPage {
		result, err := getRepositoryProject(context.Background(), graphqlClient, repoOwner, repoName, projectId, 5, cursor)
		if err != nil {
			return nil, err
		}

		if isFirstPage {
			repoProject = result
		} else {
			repoProject.Repository.ProjectV2.Items.Nodes = append(repoProject.Repository.ProjectV2.Items.Nodes, result.Repository.ProjectV2.Items.Nodes...)
		}

		isFirstPage = false
		hasNextPage = repoProject.Repository.ProjectV2.Items.PageInfo.HasNextPage
		cursor = repoProject.Repository.ProjectV2.Items.PageInfo.EndCursor
	}

	return repoProject, nil
}

func parseProjectInformation(projectFields *ProjectFields) *models.Project {
	project := &models.Project{
		Id:         projectFields.Id,
		Title:      projectFields.Title,
		Iterations: []models.Iteration{},
		Statuses:   []string{},
	}

	status := projectFields.Status.(*ProjectFieldsStatusProjectV2SingleSelectField)

	for _, option := range status.Options {
		project.Statuses = append(project.Statuses, option.Name)
	}

	iteration := projectFields.Iteration.(*ProjectFieldsIterationProjectV2IterationField)
	for _, completedIteration := range iteration.Configuration.CompletedIterations {
		startDate, _ := time.Parse("2006-01-02", completedIteration.StartDate)
		project.Iterations = append(project.Iterations, models.Iteration{
			Id:        completedIteration.Id,
			Title:     completedIteration.Title,
			StartDate: startDate,
			EndDate:   startDate.Add(time.Duration(completedIteration.Duration) * 24 * time.Hour),
		})
	}
	for _, futureIteration := range iteration.Configuration.Iterations {
		startDate, _ := time.Parse("2006-01-02", futureIteration.StartDate)
		project.Iterations = append(project.Iterations, models.Iteration{
			Id:        futureIteration.Id,
			Title:     futureIteration.Title,
			StartDate: startDate,
			EndDate:   startDate.Add(time.Duration(futureIteration.Duration) * 24 * time.Hour),
		})
	}

	for _, item := range projectFields.Items.Nodes {
		if item.Content.GetTypename() != "Issue" {
			continue
		}

		content := item.Content.(*ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue)
		issue := models.Issue{
			Id:        item.Id,
			Title:     content.Title,
			CreatedAt: content.CreatedAt,
			ClosedAt:  content.ClosedAt,
			Status:    item.Status.(*ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue).Name,
		}
		if item.Effort != nil {
			issue.Effort = item.Effort.(*ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue).Number
		}

		if item.Remaining != nil {
			issue.RemainingHours = item.Remaining.(*ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemRemainingProjectV2ItemFieldNumberValue).Number
		}

		if item.Iteration != nil {
			iteration := item.Iteration.(*ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue)

			issue.IterationId = iteration.IterationId
		}

		// extract labels from isues
		for _, label := range content.Labels.Nodes {
			issue.Labels = append(issue.Labels, label.Name)
		}

		project.Issues = append(project.Issues, issue)
	}

	return project
}
