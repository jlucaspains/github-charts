package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/jlucaspains/github-charts/models"
	"github.com/stretchr/testify/assert"
)

func TestInitOrg(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, err := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			OrgName: "org",
			Project: "1",
			Token:   "token",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "* * * * *", dataPullJob.cron)
	assert.Equal(t, "token", dataPullJob.projects[0].Token)
	assert.Equal(t, "org", dataPullJob.projects[0].OrgName)
	assert.Equal(t, "1", dataPullJob.projects[0].Project)
	assert.Empty(t, dataPullJob.projects[0].RepoName)
	assert.Empty(t, dataPullJob.projects[0].RepoOwner)
	assert.False(t, dataPullJob.running)
}

func TestInitRepo(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, err := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			RepoOwner: "org",
			RepoName:  "repo",
			Project:   "1",
			Token:     "token",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "* * * * *", dataPullJob.cron)
	assert.Equal(t, "token", dataPullJob.projects[0].Token)
	assert.Equal(t, "org", dataPullJob.projects[0].RepoOwner)
	assert.Equal(t, "repo", dataPullJob.projects[0].RepoName)
	assert.Equal(t, "1", dataPullJob.projects[0].Project)
	assert.Empty(t, dataPullJob.projects[0].OrgName)
	assert.False(t, dataPullJob.running)
}

func TestInitInvalidCron(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, err := NewDataPullJob("*", querier, []models.JobConfigItem{
		{
			RepoOwner: "org",
			RepoName:  "repo",
			Project:   "1",
			Token:     "token",
		},
	})

	assert.Error(t, err, "invalid cron expression")
	assert.Nil(t, dataPullJob)
}

func TestStart(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			OrgName: "org",
			Project: "1",
			Token:   "token",
		},
	})
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

type mockGraphqlOrgClient struct {
	result getOrganizationProjectResponse
	err    error
}

func (m mockGraphqlOrgClient) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	resp.Data.(*getOrganizationProjectResponse).Organization.ProjectV2 = m.result.Organization.ProjectV2
	return m.err
}

type mockGraphqlRepoClient struct {
	result getRepositoryProjectResponse
	err    error
}

func (m mockGraphqlRepoClient) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	resp.Data.(*getRepositoryProjectResponse).Repository.ProjectV2 = m.result.Repository.ProjectV2
	return m.err
}

func TestExecuteWillInsertOrgProject(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			OrgName: "org",
			Project: "1",
			Token:   "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlOrgClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					ProjectFields: ProjectFields{
						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name:    "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations:          []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{},
						},
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

func TestExecuteWillInsertRepoProject(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			RepoOwner: "org",
			RepoName:  "repo",
			Project:   "1",
			Token:     "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlRepoClient{
		result: getRepositoryProjectResponse{
			Repository: getRepositoryProjectRepository{
				ProjectV2: getRepositoryProjectRepositoryProjectV2{
					ProjectFields: ProjectFields{

						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name:    "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations:          []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{},
						},
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

func TestExecuteWillInsertOrgCategories(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			OrgName: "org",
			Project: "1",
			Token:   "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlOrgClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					ProjectFields: ProjectFields{
						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name: "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{
								{
									Name: "New",
								},
								{
									Name: "Done",
								},
							},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations:          []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{},
						},
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

func TestExecuteWillInsertRepoCategories(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			RepoOwner: "org",
			RepoName:  "repo",
			Project:   "1",
			Token:     "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlRepoClient{
		result: getRepositoryProjectResponse{
			Repository: getRepositoryProjectRepository{
				ProjectV2: getRepositoryProjectRepositoryProjectV2{
					ProjectFields: ProjectFields{
						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name: "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{
								{
									Name: "New",
								},
								{
									Name: "Done",
								},
							},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations:          []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{},
						},
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

func TestExecuteWillInsertOrgIterations(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			OrgName: "org",
			Project: "1",
			Token:   "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlOrgClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					ProjectFields: ProjectFields{
						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name:    "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{
									{
										Id:        "2",
										Title:     "Iteration 2",
										StartDate: "2024-01-08",
										Duration:  7,
									},
								},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{
									{
										Id:        "1",
										Title:     "Iteration 1",
										StartDate: "2024-01-01",
										Duration:  7,
									},
								},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{},
						},
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

func TestExecuteWillInsertRepoIterations(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			RepoOwner: "org",
			RepoName:  "repo",
			Project:   "1",
			Token:     "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlRepoClient{
		result: getRepositoryProjectResponse{
			Repository: getRepositoryProjectRepository{
				ProjectV2: getRepositoryProjectRepositoryProjectV2{
					ProjectFields: ProjectFields{
						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name:    "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{
									{
										Id:        "2",
										Title:     "Iteration 2",
										StartDate: "2024-01-08",
										Duration:  7,
									},
								},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{
									{
										Id:        "1",
										Title:     "Iteration 1",
										StartDate: "2024-01-01",
										Duration:  7,
									},
								},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{},
						},
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

func TestExecuteWillInsertOrgWorkItems(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			OrgName: "org",
			Project: "1",
			Token:   "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlOrgClient{
		result: getOrganizationProjectResponse{
			Organization: getOrganizationProjectOrganization{
				ProjectV2: getOrganizationProjectOrganizationProjectV2{
					ProjectFields: ProjectFields{
						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name: "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{
								{
									Name: "New",
								},
							},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{
									{
										Id:        "2",
										Title:     "Iteration 2",
										StartDate: "2024-01-08",
										Duration:  7,
									},
								},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{
								{
									Id: "1",
									Status: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue{
										Name: "New",
									},
									Effort: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue{
										Number: 5,
									},
									Remaining: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemRemainingProjectV2ItemFieldNumberValue{
										Number: 16,
									},
									Iteration: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue{
										IterationId: "2",
									},
									Content: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue{
										Typename:  "Issue",
										Title:     "Issue 1",
										CreatedAt: time.Now().AddDate(0, 0, -1),
										Labels: ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnection{
											Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnectionNodesLabel{
												{
													Name: "Label 1",
												},
											},
										},
									},
								},
								{
									Id: "2",
									Status: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue{
										Name: "New",
									},
									Effort: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue{
										Number: 3,
									},
									Remaining: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemRemainingProjectV2ItemFieldNumberValue{
										Number: 8,
									},
									Iteration: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue{
										IterationId: "2",
									},
									Content: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue{
										Typename:  "Issue",
										Title:     "Issue 2",
										CreatedAt: time.Now().AddDate(0, 0, -2),
										Labels: ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnection{
											Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnectionNodesLabel{
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
	assert.Equal(t, time.Now().UTC().Format("2006-01-02"), querier.UpsertWorkItemsValue[0].ChangeDate.Time.Format("2006-01-02"))
	assert.Equal(t, "2", querier.UpsertWorkItemsValue[1].GhID)
	assert.Equal(t, "Issue 2", querier.UpsertWorkItemsValue[1].Name)
	assert.Equal(t, int64(3), effort2.Int64)
	assert.Equal(t, int64(8), remaining2.Int64)
	assert.Equal(t, time.Now().UTC().Format("2006-01-02"), querier.UpsertWorkItemsValue[0].ChangeDate.Time.Format("2006-01-02"))
}

func TestExecuteWillInsertRepoWorkItems(t *testing.T) {
	querier := &MockQuerier{}
	dataPullJob, _ := NewDataPullJob("* * * * *", querier, []models.JobConfigItem{
		{
			RepoOwner: "org",
			RepoName:  "repo",
			Project:   "1",
			Token:     "token",
		},
	})
	dataPullJob.graphqlClients[dataPullJob.projects[0].GetUniqueName()] = mockGraphqlRepoClient{
		result: getRepositoryProjectResponse{
			Repository: getRepositoryProjectRepository{
				ProjectV2: getRepositoryProjectRepositoryProjectV2{
					ProjectFields: ProjectFields{
						Id:    "1",
						Title: "Project 1",
						Status: &ProjectFieldsStatusProjectV2SingleSelectField{
							Name: "Status",
							Options: []ProjectFieldsStatusProjectV2SingleSelectFieldOptionsProjectV2SingleSelectFieldOption{
								{
									Name: "New",
								},
							},
						},
						Iteration: &ProjectFieldsIterationProjectV2IterationField{
							Name: "Iteration",
							Configuration: ProjectFieldsIterationProjectV2IterationFieldConfiguration{
								Iterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationIterationsProjectV2IterationFieldIteration{
									{
										Id:        "2",
										Title:     "Iteration 2",
										StartDate: "2024-01-08",
										Duration:  7,
									},
								},
								CompletedIterations: []ProjectFieldsIterationProjectV2IterationFieldConfigurationCompletedIterationsProjectV2IterationFieldIteration{},
							},
						},
						Items: ProjectFieldsItemsProjectV2ItemConnection{
							Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2Item{
								{
									Id: "1",
									Status: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue{
										Name: "New",
									},
									Effort: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue{
										Number: 5,
									},
									Remaining: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemRemainingProjectV2ItemFieldNumberValue{
										Number: 16,
									},
									Iteration: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue{
										IterationId: "2",
									},
									Content: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue{
										Typename:  "Issue",
										Title:     "Issue 1",
										CreatedAt: time.Now().AddDate(0, 0, -1),
										Labels: ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnection{
											Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnectionNodesLabel{
												{
													Name: "Label 1",
												},
											},
										},
									},
								},
								{
									Id: "2",
									Status: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemStatusProjectV2ItemFieldSingleSelectValue{
										Name: "New",
									},
									Effort: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemEffortProjectV2ItemFieldNumberValue{
										Number: 3,
									},
									Remaining: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemRemainingProjectV2ItemFieldNumberValue{
										Number: 8,
									},
									Iteration: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemIterationProjectV2ItemFieldIterationValue{
										IterationId: "2",
									},
									Content: &ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssue{
										Typename:  "Issue",
										Title:     "Issue 2",
										CreatedAt: time.Now().AddDate(0, 0, -2),
										Labels: ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnection{
											Nodes: []ProjectFieldsItemsProjectV2ItemConnectionNodesProjectV2ItemContentIssueLabelsLabelConnectionNodesLabel{
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
	assert.Equal(t, time.Now().UTC().Format("2006-01-02"), querier.UpsertWorkItemsValue[0].ChangeDate.Time.Format("2006-01-02"))
	assert.Equal(t, "2", querier.UpsertWorkItemsValue[1].GhID)
	assert.Equal(t, "Issue 2", querier.UpsertWorkItemsValue[1].Name)
	assert.Equal(t, int64(3), effort2.Int64)
	assert.Equal(t, int64(8), remaining2.Int64)
	assert.Equal(t, time.Now().UTC().Format("2006-01-02"), querier.UpsertWorkItemsValue[0].ChangeDate.Time.Format("2006-01-02"))
}
