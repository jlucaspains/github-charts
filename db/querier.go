package db

import "context"

type Querier interface {
	GetIterationBurndown(ctx context.Context, id int32) ([]GetIterationBurndownRow, error)
	GetIterations(ctx context.Context, projectID int32) ([]Iteration, error)
	GetProjectBurnup(ctx context.Context, arg GetProjectBurnupParams) ([]GetProjectBurnupRow, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetWorkItemsForIteration(ctx context.Context, name string) ([]GetWorkItemsForIterationRow, error)
	UpsertIteration(ctx context.Context, arg UpsertIterationParams) (Iteration, error)
	UpsertProject(ctx context.Context, arg UpsertProjectParams) (Project, error)
	UpsertWorkItem(ctx context.Context, arg UpsertWorkItemParams) (WorkItemHistory, error)
	UpsertWorkItemStatus(ctx context.Context, name string) (WorkItemStatus, error)
}
