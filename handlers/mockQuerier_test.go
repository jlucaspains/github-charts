package handlers

import (
	"context"

	"github.com/jlucaspains/github-charts/db"
)

// mock for Queries
type MockQuerier struct {
	GetIterationsResult []db.Iteration
	GetIterationsError  error

	GetIterationBurndownResult []db.GetIterationBurndownRow
	GetIterationBurndownError  error

	GetProjectsResult []db.Project
	GetProjectsError  error

	GetProjectBurnupResult []db.GetProjectBurnupRow
	GetProjectBurnupError  error
}

// GetIterationBurndown implements Querier.
func (m *MockQuerier) GetIterationBurndown(ctx context.Context, id int64) ([]db.GetIterationBurndownRow, error) {
	return m.GetIterationBurndownResult, m.GetIterationBurndownError
}

// GetIterations implements Querier.
func (m *MockQuerier) GetIterations(ctx context.Context) ([]db.Iteration, error) {
	return m.GetIterationsResult, m.GetIterationsError
}

// GetProjectBurnup implements Querier.
func (m *MockQuerier) GetProjectBurnup(ctx context.Context, arg db.GetProjectBurnupParams) ([]db.GetProjectBurnupRow, error) {
	return m.GetProjectBurnupResult, m.GetProjectBurnupError
}

// GetProjects implements Querier.
func (m *MockQuerier) GetProjects(ctx context.Context) ([]db.Project, error) {
	return m.GetProjectsResult, m.GetProjectsError
}

// GetWorkItemsForIteration implements Querier.
func (m *MockQuerier) GetWorkItemsForIteration(ctx context.Context, name string) ([]db.GetWorkItemsForIterationRow, error) {
	panic("unimplemented")
}

// UpsertIteration implements Querier.
func (m *MockQuerier) UpsertIteration(ctx context.Context, arg db.UpsertIterationParams) (db.Iteration, error) {
	panic("unimplemented")
}

// UpsertProject implements Querier.
func (m *MockQuerier) UpsertProject(ctx context.Context, arg db.UpsertProjectParams) (db.Project, error) {
	panic("unimplemented")
}

// UpsertWorkItem implements Querier.
func (m *MockQuerier) UpsertWorkItem(ctx context.Context, arg db.UpsertWorkItemParams) (db.WorkItemHistory, error) {
	panic("unimplemented")
}

// UpsertWorkItemStatus implements Querier.
func (m *MockQuerier) UpsertWorkItemStatus(ctx context.Context, name string) (db.WorkItemStatus, error) {
	panic("unimplemented")
}
