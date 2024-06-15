package jobs

import (
	"context"

	"github.com/jlucaspains/github-charts/db"
)

// mock for Queries
type MockQuerier struct {
	UpsertProjectValue db.UpsertProjectParams
	UpsertProjectError error

	UpsertWorkItemStatusValue []string
	UpsertWorkItemStatusError error

	UpsertWorkItemIterationsValue []db.UpsertIterationParams
	UpsertWorkItemIterationsError error

	UpsertWorkItemsValue []db.UpsertWorkItemParams
	UpsertWorkItemsError error
}

// GetIterationBurndown implements Querier.
func (m *MockQuerier) GetIterationBurndown(ctx context.Context, id int64) ([]db.GetIterationBurndownRow, error) {
	panic("unimplemented")
}

// GetIterations implements Querier.
func (m *MockQuerier) GetIterations(ctx context.Context) ([]db.Iteration, error) {
	panic("unimplemented")
}

// GetProjectBurnup implements Querier.
func (m *MockQuerier) GetProjectBurnup(ctx context.Context, arg db.GetProjectBurnupParams) ([]db.GetProjectBurnupRow, error) {
	panic("unimplemented")
}

// GetProjects implements Querier.
func (m *MockQuerier) GetProjects(ctx context.Context) ([]db.Project, error) {
	panic("unimplemented")
}

// GetWorkItemsForIteration implements Querier.
func (m *MockQuerier) GetWorkItemsForIteration(ctx context.Context, name string) ([]db.GetWorkItemsForIterationRow, error) {
	panic("unimplemented")
}

// UpsertIteration implements Querier.
func (m *MockQuerier) UpsertIteration(ctx context.Context, arg db.UpsertIterationParams) (db.Iteration, error) {
	m.UpsertWorkItemIterationsValue = append(m.UpsertWorkItemIterationsValue, arg)

	return db.Iteration{
		ID:        1,
		GhID:      arg.GhID,
		Name:      arg.Name,
		StartDate: arg.StartDate,
		EndDate:   arg.EndDate,
	}, m.UpsertWorkItemIterationsError
}

// UpsertProject implements Querier.
func (m *MockQuerier) UpsertProject(ctx context.Context, arg db.UpsertProjectParams) (db.Project, error) {
	m.UpsertProjectValue = arg
	return db.Project{}, m.UpsertProjectError
}

// UpsertWorkItem implements Querier.
func (m *MockQuerier) UpsertWorkItem(ctx context.Context, arg db.UpsertWorkItemParams) (db.WorkItemHistory, error) {
	m.UpsertWorkItemsValue = append(m.UpsertWorkItemsValue, arg)

	return db.WorkItemHistory{
		ID:             1,
		GhID:           arg.GhID,
		IterationID:    arg.IterationID,
		ProjectID:      arg.ProjectID,
		Status:         arg.Status,
		Effort:         arg.Effort,
		ChangeDate:     arg.ChangeDate,
		Name:           arg.Name,
		Priority:       arg.Priority,
		RemainingHours: arg.RemainingHours,
	}, m.UpsertWorkItemsError
}

// UpsertWorkItemStatus implements Querier.
func (m *MockQuerier) UpsertWorkItemStatus(ctx context.Context, name string) (db.WorkItemStatus, error) {
	m.UpsertWorkItemStatusValue = append(m.UpsertWorkItemStatusValue, name)
	return db.WorkItemStatus{
		ID:   1,
		Name: name,
	}, m.UpsertWorkItemStatusError
}
