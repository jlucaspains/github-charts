// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getIterationBurndown = `-- name: GetIterationBurndown :many

with starting_effort as (
 select sum(effort) as effort
 from work_item_history
 where change_date = (select min(change_date) from work_item_history where iteration_id = 1)
   and iteration_id = 1)

select iteration_day
     , cast(sum(case when status <> 'Done' then work_item_history.effort else 0 end) as decimal) as remaining
     , cast(seffort.effort::decimal - (seffort.effort::decimal / total_days.total * row_number() over (order by iteration_day)) as decimal) as ideal
  from iteration
       join lateral (SELECT date_trunc('day', dd):: date as iteration_day
                       FROM generate_series
                               ( iteration.start_date::timestamp 
                               , iteration.end_date::timestamp
                               , '1 day'::interval) dd
                      where EXTRACT(ISODOW FROM dd) not IN (6, 7)) dates on true
       join lateral (SELECT count(*)::decimal as total
                       FROM generate_series
                               ( iteration.start_date::timestamp 
                               , iteration.end_date::timestamp
                               , '1 day'::interval) dd
                      where EXTRACT(ISODOW FROM dd) not IN (6, 7)) total_days on true
        join lateral (select effort from starting_effort) seffort on true
       left join work_item_history on work_item_history.change_date = dates.iteration_day
 where iteration.id = 1
 group by iteration_day, total_days.total, seffort.effort
order by iteration_day
`

type GetIterationBurndownRow struct {
	IterationDay pgtype.Date
	Remaining    pgtype.Numeric
	Ideal        pgtype.Numeric
}

func (q *Queries) GetIterationBurndown(ctx context.Context) ([]GetIterationBurndownRow, error) {
	rows, err := q.db.Query(ctx, getIterationBurndown)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetIterationBurndownRow
	for rows.Next() {
		var i GetIterationBurndownRow
		if err := rows.Scan(&i.IterationDay, &i.Remaining, &i.Ideal); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getWorkItemsForIteration = `-- name: GetWorkItemsForIteration :many
SELECT work_item_history.id, change_date, work_item_history.gh_id, project_id, work_item_history.name, status, priority, remaining_hours, effort, iteration_id, iteration.id, iteration.gh_id, iteration.name, start_date, end_date FROM work_item_history
join iteration on work_item.iteration_id = iteration.id
WHERE iteration.name = $1
`

type GetWorkItemsForIterationRow struct {
	ID             int64
	ChangeDate     pgtype.Date
	GhID           string
	ProjectID      int64
	Name           string
	Status         pgtype.Text
	Priority       pgtype.Int4
	RemainingHours pgtype.Int4
	Effort         pgtype.Int4
	IterationID    pgtype.Int8
	ID_2           int64
	GhID_2         string
	Name_2         string
	StartDate      pgtype.Date
	EndDate        pgtype.Date
}

func (q *Queries) GetWorkItemsForIteration(ctx context.Context, name string) ([]GetWorkItemsForIterationRow, error) {
	rows, err := q.db.Query(ctx, getWorkItemsForIteration, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetWorkItemsForIterationRow
	for rows.Next() {
		var i GetWorkItemsForIterationRow
		if err := rows.Scan(
			&i.ID,
			&i.ChangeDate,
			&i.GhID,
			&i.ProjectID,
			&i.Name,
			&i.Status,
			&i.Priority,
			&i.RemainingHours,
			&i.Effort,
			&i.IterationID,
			&i.ID_2,
			&i.GhID_2,
			&i.Name_2,
			&i.StartDate,
			&i.EndDate,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertIteration = `-- name: UpsertIteration :one
INSERT INTO iteration (gh_id, name, start_date, end_date)
VALUES ($1, $2, $3, $4)
ON CONFLICT(gh_id) 
DO UPDATE SET
  "name" = EXCLUDED.name,
  start_date = EXCLUDED.start_date,
  end_date = EXCLUDED.end_date
RETURNING id, gh_id, name, start_date, end_date
`

type UpsertIterationParams struct {
	GhID      string
	Name      string
	StartDate pgtype.Date
	EndDate   pgtype.Date
}

func (q *Queries) UpsertIteration(ctx context.Context, arg UpsertIterationParams) (Iteration, error) {
	row := q.db.QueryRow(ctx, upsertIteration,
		arg.GhID,
		arg.Name,
		arg.StartDate,
		arg.EndDate,
	)
	var i Iteration
	err := row.Scan(
		&i.ID,
		&i.GhID,
		&i.Name,
		&i.StartDate,
		&i.EndDate,
	)
	return i, err
}

const upsertProject = `-- name: UpsertProject :one
INSERT INTO project (gh_id, name)
VALUES ($1, $2)
ON CONFLICT(gh_id) 
DO UPDATE SET
  "name" = EXCLUDED.name
RETURNING id, gh_id, name
`

type UpsertProjectParams struct {
	GhID string
	Name string
}

func (q *Queries) UpsertProject(ctx context.Context, arg UpsertProjectParams) (Project, error) {
	row := q.db.QueryRow(ctx, upsertProject, arg.GhID, arg.Name)
	var i Project
	err := row.Scan(&i.ID, &i.GhID, &i.Name)
	return i, err
}

const upsertWorkItem = `-- name: UpsertWorkItem :one
INSERT INTO work_item_history (change_date, gh_id, project_id, name, status, priority, remaining_hours, effort, iteration_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT(change_date, gh_id) 
DO UPDATE SET
  project_id = EXCLUDED.project_id,
  "name" = EXCLUDED.name,
  "status" = EXCLUDED.status,
  priority = EXCLUDED.priority,
  remaining_hours = EXCLUDED.remaining_hours,
  effort = EXCLUDED.effort,
  iteration_id = EXCLUDED.iteration_id
RETURNING id, change_date, gh_id, project_id, name, status, priority, remaining_hours, effort, iteration_id
`

type UpsertWorkItemParams struct {
	ChangeDate     pgtype.Date
	GhID           string
	ProjectID      int64
	Name           string
	Status         pgtype.Text
	Priority       pgtype.Int4
	RemainingHours pgtype.Int4
	Effort         pgtype.Int4
	IterationID    pgtype.Int8
}

func (q *Queries) UpsertWorkItem(ctx context.Context, arg UpsertWorkItemParams) (WorkItemHistory, error) {
	row := q.db.QueryRow(ctx, upsertWorkItem,
		arg.ChangeDate,
		arg.GhID,
		arg.ProjectID,
		arg.Name,
		arg.Status,
		arg.Priority,
		arg.RemainingHours,
		arg.Effort,
		arg.IterationID,
	)
	var i WorkItemHistory
	err := row.Scan(
		&i.ID,
		&i.ChangeDate,
		&i.GhID,
		&i.ProjectID,
		&i.Name,
		&i.Status,
		&i.Priority,
		&i.RemainingHours,
		&i.Effort,
		&i.IterationID,
	)
	return i, err
}
