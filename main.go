package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/wcharczuk/go-chart/v2"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

type Iteration struct {
	Title     string    `json:"title"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type Issue struct {
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	ClosedAt  time.Time `json:"closedAt"`
	Status    string    `json:"status"`
	Effort    float64   `json:"effort"`
	Labels    []string  `json:"labels"`
	Iteration Iteration `json:"iteration"`
}

type Project struct {
	Title  string  `json:"title"`
	Issues []Issue `json:"issues"`
}

func main() {
	println("Hello, World!")

	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	key := os.Getenv("GITHUB_TOKEN")
	if key == "" {
		err = fmt.Errorf("must set GITHUB_TOKEN=<github token>")
		return
	}

	httpClient := http.Client{
		Transport: &authedTransport{
			key:     key,
			wrapped: http.DefaultTransport,
		},
	}
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", &httpClient)

	orgProject, err := getOrganizationProject(context.Background(), graphqlClient, "hitachisolutionsamerica", 12, 5, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(orgProject.Organization.ProjectV2.Title)

	parsedProject := parseProjectInformation(orgProject)
	total, stats := createStatsForProject(parsedProject, &parsedProject.Issues[0].Iteration)
	renderChart(total, stats)

}

func parseProjectInformation(orgProject *getOrganizationProjectResponse) *Project {
	//parse project information
	project := &Project{}
	project.Title = orgProject.Organization.ProjectV2.Title

	for _, item := range orgProject.Organization.ProjectV2.Items.Nodes {
		content := item.Content.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue)
		issue := Issue{}
		issue.Title = content.Title
		issue.CreatedAt = content.CreatedAt
		issue.ClosedAt = content.ClosedAt
		issue.Status = item.Status.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue).Name

		if item.Effort != nil {
			issue.Effort = item.Effort.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue).Number
		}

		if item.Iteration != nil {
			iteration := item.Iteration.(*getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue)

			startDate, _ := time.Parse("2006-01-02", iteration.StartDate)
			issue.Iteration = Iteration{
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

func createStatsForProject(project *Project, iteration *Iteration) (float64, map[time.Time]float64) {
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

func renderChart(total float64, data map[time.Time]float64) {
	keys := make([]time.Time, 0, len(data))
	closed := make([]float64, 0, len(data))
	ideal := make([]float64, 0, len(data))
	idealPerDay := total / float64(len(data))

	for t := range data {
		keys = append(keys, t)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	for idx, key := range keys {
		closed = append(closed, data[key])
		ideal = append(ideal, total-(idealPerDay*float64(idx+1)))
	}

	graph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				Style: chart.Style{
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
				},
				XValues: keys,
				YValues: closed,
			},
			chart.TimeSeries{
				XValues: keys,
				YValues: ideal,
			},
			chart.TimeSeries{
				XValues: []time.Time{time.Now(), time.Now(), time.Now(), time.Now()},
				YValues: []float64{total, total - 1, total - 2, total - 3},
			},
		},
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	graph.Render(chart.PNG, f)
}
