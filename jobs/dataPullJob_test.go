package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	dataPullJob := &DataPullJob{}
	querier := &MockQuerier{}
	err := dataPullJob.Init("* * * * *", querier, 1, "token", "org")

	assert.Nil(t, err)
	assert.Equal(t, "* * * * *", dataPullJob.cron)
	assert.Equal(t, "token", dataPullJob.ghToken)
	assert.Equal(t, "org", dataPullJob.organization)
	assert.Equal(t, 1, dataPullJob.projectId)
	assert.False(t, dataPullJob.running)
}

func TestInitInvalidCron(t *testing.T) {
	dataPullJob := &DataPullJob{}
	querier := &MockQuerier{}
	err := dataPullJob.Init("*", querier, 1, "token", "org")

	assert.Error(t, err, "invalid cron expression")
	assert.False(t, dataPullJob.running)
}

func TestStart(t *testing.T) {
	dataPullJob := &DataPullJob{}
	querier := &MockQuerier{}

	dataPullJob.Init("* * * * *", querier, 1, "token", "org")
	dataPullJob.Start()

	assert.True(t, dataPullJob.running)
}

func TestStop(t *testing.T) {
	dataPullJob := &DataPullJob{
		running: true,
		ticker:  time.NewTicker(time.Minute),
	}

	dataPullJob.Start()
	dataPullJob.Stop()

	assert.False(t, dataPullJob.running)
}

type mockGraphqlClient struct {
	result getOrganizationProjectResponse
	err    error
}

func (m mockGraphqlClient) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	resp.Data.(*getOrganizationProjectResponse).Organization.ProjectV2 = m.result.Organization.ProjectV2
	return m.err
}

func TestExecuteWillInsertProject(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob := &DataPullJob{}
	dataPullJob.Init("* * * * *", querier, 1, "token", "org")
	dataPullJob.graphqlClient = mockGraphqlClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					Id:    "1",
					Title: "Project 1",
					Status: &getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectField{
						Name:    "Status",
						Options: []getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{},
					},
					Iteration: &getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationField{
						Name: "Iteration",
						Configuration: getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfiguration{
							Iterations:          []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{},
							CompletedIterations: []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
						},
					},
					Items: getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnection{
						Nodes: []getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2Item{},
					},
				},
			},
		},
	}
	dataPullJob.Start()

	dataPullJob.tryExecute()

	assert.NotNil(t, querier.UpsertProjectValue)
	assert.Equal(t, "1", querier.UpsertProjectValue.GhID)
	assert.Equal(t, "Project 1", querier.UpsertProjectValue.Name)
}

func TestExecuteWillInsertCategories(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob := &DataPullJob{}
	dataPullJob.Init("* * * * *", querier, 1, "token", "org")
	dataPullJob.graphqlClient = mockGraphqlClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					Id:    "1",
					Title: "Project 1",
					Status: &getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectField{
						Name: "Status",
						Options: []getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{
							{
								Name: "New",
							},
							{
								Name: "Done",
							},
						},
					},
					Iteration: &getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationField{
						Name: "Iteration",
						Configuration: getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfiguration{
							Iterations:          []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{},
							CompletedIterations: []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
						},
					},
					Items: getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnection{
						Nodes: []getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2Item{},
					},
				},
			},
		},
	}
	dataPullJob.Start()

	dataPullJob.tryExecute()

	assert.NotNil(t, querier.UpsertWorkItemStatusValue)
	assert.Equal(t, "New", querier.UpsertWorkItemStatusValue[0])
	assert.Equal(t, "Done", querier.UpsertWorkItemStatusValue[1])
}

func TestExecuteWillInsertIterations(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob := &DataPullJob{}
	dataPullJob.Init("* * * * *", querier, 1, "token", "org")
	dataPullJob.graphqlClient = mockGraphqlClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					Id:    "1",
					Title: "Project 1",
					Status: &getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectField{
						Name:    "Status",
						Options: []getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{},
					},
					Iteration: &getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationField{
						Name: "Iteration",
						Configuration: getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfiguration{
							Iterations: []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{
								{
									Id:        "2",
									Title:     "Iteration 2",
									StartDate: "2024-01-08",
									Duration:  7,
								},
							},
							CompletedIterations: []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{
								{
									Id:        "1",
									Title:     "Iteration 1",
									StartDate: "2024-01-01",
									Duration:  7,
								},
							},
						},
					},
					Items: getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnection{
						Nodes: []getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2Item{},
					},
				},
			},
		},
	}
	dataPullJob.Start()

	dataPullJob.tryExecute()

	assert.NotNil(t, querier.UpsertWorkItemIterationsValue)
	assert.Nil(t, querier.UpsertWorkItemIterationsError)
	assert.Equal(t, "1", querier.UpsertWorkItemIterationsValue[0].GhID)
	assert.Equal(t, "Iteration 1", querier.UpsertWorkItemIterationsValue[0].Name)
	assert.Equal(t, "2024-01-01", querier.UpsertWorkItemIterationsValue[0].StartDate.Time.Format("2006-01-02"))
	assert.Equal(t, "2024-01-08", querier.UpsertWorkItemIterationsValue[0].EndDate.Time.Format("2006-01-02"))
	assert.Equal(t, "2", querier.UpsertWorkItemIterationsValue[1].GhID)
	assert.Equal(t, "Iteration 2", querier.UpsertWorkItemIterationsValue[1].Name)
	assert.Equal(t, "2024-01-08", querier.UpsertWorkItemIterationsValue[1].StartDate.Time.Format("2006-01-02"))
	assert.Equal(t, "2024-01-15", querier.UpsertWorkItemIterationsValue[1].EndDate.Time.Format("2006-01-02"))
}

func TestExecuteWillInsertWorkItems(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob := &DataPullJob{}
	dataPullJob.Init("* * * * *", querier, 1, "token", "org")
	dataPullJob.graphqlClient = mockGraphqlClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					Id:    "1",
					Title: "Project 1",
					Status: &getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectField{
						Name: "Status",
						Options: []getOrganizationProjectOrganizationProjectV2StatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{
							{
								Name: "New",
							},
						},
					},
					Iteration: &getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationField{
						Name: "Iteration",
						Configuration: getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfiguration{
							Iterations: []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{
								{
									Id:        "2",
									Title:     "Iteration 2",
									StartDate: "2024-01-08",
									Duration:  7,
								},
							},
							CompletedIterations: []getOrganizationProjectOrganizationProjectV2IterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
						},
					},
					Items: getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnection{
						Nodes: []getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2Item{
							{
								Id: "1",
								Status: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue{
									Name: "New",
								},
								Effort: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue{
									Number: 5,
								},
								Remaining: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemRemainingProjectV2ItemFieldNumberValue{
									Number: 16,
								},
								Iteration: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue{
									IterationId: "2",
								},
								Content: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue{
									Title:     "Issue 1",
									CreatedAt: time.Now().AddDate(0, 0, -1),
									Labels: getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnection{
										Nodes: []getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnectionNodesLabel{
											{
												Name: "Label 1",
											},
										},
									},
								},
							},
							{
								Id: "2",
								Status: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue{
									Name: "New",
								},
								Effort: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue{
									Number: 3,
								},
								Remaining: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemRemainingProjectV2ItemFieldNumberValue{
									Number: 8,
								},
								Iteration: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue{
									IterationId: "2",
								},
								Content: &getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue{
									Title:     "Issue 2",
									CreatedAt: time.Now().AddDate(0, 0, -2),
									Labels: getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnection{
										Nodes: []getOrganizationProjectOrganizationProjectV2ItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnectionNodesLabel{
											{
												Name: "Label 2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	dataPullJob.Start()

	dataPullJob.tryExecute()

	effort1, _ := querier.UpsertWorkItemsValue[0].Effort.Int64Value()
	remaining1, _ := querier.UpsertWorkItemsValue[0].RemainingHours.Int64Value()
	effort2, _ := querier.UpsertWorkItemsValue[1].Effort.Int64Value()
	remaining2, _ := querier.UpsertWorkItemsValue[1].RemainingHours.Int64Value()
	assert.NotNil(t, querier.UpsertWorkItemsValue)
	assert.Nil(t, querier.UpsertWorkItemsError)
	assert.Equal(t, "1", querier.UpsertWorkItemsValue[0].GhID)
	assert.Equal(t, "Issue 1", querier.UpsertWorkItemsValue[0].Name)
	assert.Equal(t, int64(5), effort1.Int64)
	assert.Equal(t, int64(16), remaining1.Int64)
	assert.Equal(t, time.Now().Format("2006-01-02"), querier.UpsertWorkItemsValue[0].ChangeDate.Time.Format("2006-01-02"))
	assert.Equal(t, "2", querier.UpsertWorkItemsValue[1].GhID)
	assert.Equal(t, "Issue 2", querier.UpsertWorkItemsValue[1].Name)
	assert.Equal(t, int64(3), effort2.Int64)
	assert.Equal(t, int64(8), remaining2.Int64)
	assert.Equal(t, time.Now().Format("2006-01-02"), querier.UpsertWorkItemsValue[0].ChangeDate.Time.Format("2006-01-02"))
}
