package db

import "context"

// mock for Queries
type MockQuerier struct {
	GetIterationsResult []Iteration
}

// GetIterationBurndown implements Querier.
func (m *MockQuerier) GetIterationBurndown(ctx context.Context, id int64) ([]GetIterationBurndownRow, error) {
	panic("unimplemented")
}

// GetIterations implements Querier.
func (m *MockQuerier) GetIterations(ctx context.Context) ([]Iteration, error) {
	return m.GetIterationsResult, nil
}

// GetProjectBurnup implements Querier.
func (m *MockQuerier) GetProjectBurnup(ctx context.Context, arg GetProjectBurnupParams) ([]GetProjectBurnupRow, error) {
	panic("unimplemented")
}

// GetProjects implements Querier.
func (m *MockQuerier) GetProjects(ctx context.Context) ([]Project, error) {
	panic("unimplemented")
}

// GetWorkItemsForIteration implements Querier.
func (m *MockQuerier) GetWorkItemsForIteration(ctx context.Context, name string) ([]GetWorkItemsForIterationRow, error) {
	panic("unimplemented")
}

// UpsertIteration implements Querier.
func (m *MockQuerier) UpsertIteration(ctx context.Context, arg UpsertIterationParams) (Iteration, error) {
	panic("unimplemented")
}

// UpsertProject implements Querier.
func (m *MockQuerier) UpsertProject(ctx context.Context, arg UpsertProjectParams) (Project, error) {
	panic("unimplemented")
}

// UpsertWorkItem implements Querier.
func (m *MockQuerier) UpsertWorkItem(ctx context.Context, arg UpsertWorkItemParams) (WorkItemHistory, error) {
	panic("unimplemented")
}

// UpsertWorkItemStatus implements Querier.
func (m *MockQuerier) UpsertWorkItemStatus(ctx context.Context, name string) (WorkItemStatus, error) {
	panic("unimplemented")
}
