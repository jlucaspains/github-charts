package jobs

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/adhocore/gronx"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jlucaspains/github-charts/db"
	"github.com/jlucaspains/github-charts/models"
)

type DataPullJob struct {
	cron         string
	ticker       *time.Ticker
	gron         *gronx.Gronx
	running      bool
	queries      *db.Queries
	ghToken      string
	projectId    int
	organization string
}

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

func (c *DataPullJob) Init(schedule string, queries *db.Queries, projectId int, ghToken string, organization string) error {
	c.gron = gronx.New()

	if schedule == "" || !c.gron.IsValid(schedule) {
		log.Printf("A valid cron schedule is required in the format e.g.: * * * * *")
		return fmt.Errorf("a valid cron schedule is required")
	}

	c.cron = schedule
	c.projectId = projectId
	c.ticker = time.NewTicker(time.Minute)
	c.ghToken = ghToken
	c.organization = organization
	c.queries = queries

	return nil
}

func (c *DataPullJob) Start() {
	c.running = true
	log.Printf("Started DataPullJob job with cron %s", c.cron)

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

	log.Printf("tryExecute job, isDue: %t", due)

	if due {
		c.execute()
	}
}

func (c *DataPullJob) execute() {
	log.Printf("execute job")

	httpClient := http.Client{
		Transport: &authedTransport{
			key:     c.ghToken,
			wrapped: http.DefaultTransport,
		},
	}
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", &httpClient)

	orgProject, err := getOrgProject(graphqlClient, c.organization, c.projectId)

	if err != nil {
		return
	}

	project := parseProjectInformation(orgProject)
	saveProjectInformation(project, c.queries)
}

func saveProjectInformation(project *models.Project, queries *db.Queries) error {
	ctx := context.Background()
	dbProject, _ := queries.UpsertProject(ctx, db.UpsertProjectParams{
		GhID: project.Id,
		Name: project.Title,
	})

	for _, issue := range project.Issues {
		dbIteration, err := queries.UpsertIteration(ctx, db.UpsertIterationParams{
			GhID:      project.Id,
			Name:      project.Title,
			StartDate: pgtype.Date{Time: issue.Iteration.StartDate, Valid: true},
			EndDate:   pgtype.Date{Time: issue.Iteration.EndDate, Valid: true},
		})

		if err != nil {
			return err
		}

		now := time.Now().Truncate(24 * time.Hour)
		today := now.UTC()

		_, err = queries.UpsertWorkItem(ctx, db.UpsertWorkItemParams{
			GhID:        issue.Id,
			ChangeDate:  pgtype.Date{Time: today, Valid: true},
			Name:        issue.Title,
			Effort:      pgtype.Int4{Int32: int32(issue.Effort), Valid: true},
			Status:      pgtype.Text{String: issue.Status, Valid: true},
			ProjectID:   dbProject.ID,
			IterationID: pgtype.Int8{Int64: dbIteration.ID, Valid: true},
		})

		if err != nil {
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

func parseProjectInformation(orgProject *getOrganizationProjectResponse) *models.Project {
	project := &models.Project{
		Id:    orgProject.Organization.ProjectV2.Id,
		Title: orgProject.Organization.ProjectV2.Title,
	}

	for _, item := range orgProject.Organization.ProjectV2.Items.Nodes {
		content := item.Content.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue)
		issue := models.Issue{
			Id:        item.Id,
			Title:     content.Title,
			CreatedAt: content.CreatedAt,
			ClosedAt:  content.ClosedAt,
			Status:    item.Status.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue).Name,
		}
		if item.Effort != nil {
			issue.Effort = item.Effort.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue).Number
		}

		if item.Iteration != nil {
			iteration := item.Iteration.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue)

			startDate, _ := time.Parse("2006-01-02", iteration.StartDate)
			issue.Iteration = models.Iteration{
				Id:        iteration.Id,
				Title:     iteration.Title,
				StartDate: startDate,
				EndDate:   startDate.Add(time.Duration(iteration.Duration) * 24 * time.Hour),
			}
		}

		// extract labels from isues
		for _, label := range content.Labels.Nodes {
			issue.Labels = append(issue.Labels, label.Name)
		}

		project.Issues = append(project.Issues, issue)
	}

	return project
}

func createStatsForProject(project *models.Project, iteration *models.Iteration) (float64, map[time.Time]float64) {
	closedItemsStats := make(map[time.Time]float64)
	result := make(map[time.Time]float64)

	remainingEffort := 0.0

	for _, issue := range project.Issues {
		if issue.Iteration.Title == iteration.Title {
			remainingEffort += issue.Effort

			if issue.Status == "Done" {
				closedItemsStats[getDatePart(issue.ClosedAt)] += issue.Effort
			}
		}
	}

	totalEffort := remainingEffort

	for _, day := range daysBetween(iteration.StartDate, iteration.EndDate) {
		remainingEffort -= closedItemsStats[day]
		result[day] = remainingEffort
	}

	return totalEffort, result
}

func daysBetween(time1, time2 time.Time) []time.Time {
	// return an array with each day in between time1 and time2
	var days []time.Time
	for d := time1; d.Before(time2) || d == time2; d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}

	return days
}

// gets date only part of the time
func getDatePart(value time.Time) time.Time {
	return value.Truncate(24 * time.Hour)
}
