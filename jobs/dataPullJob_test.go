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
	dataPullJob := &DataPullJob{
		cron:    "* * * * *",
		ticker:  nil,
		gron:    nil,
		running: false,
		queries: nil,
		ghToken: "token",
	}

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
