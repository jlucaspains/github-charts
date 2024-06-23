-- name: GetWorkItemsForIteration :many
SELECT id, change_date, gh_id, name, status, priority, remaining_hours, effort, iteration_id, project_id
FROM work_item_history
JOIN iteration on work_item.iteration_id = iteration.id
WHERE iteration.name = $1;

-- name: UpsertWorkItem :one
INSERT INTO work_item_history (change_date, gh_id, name, status, priority, remaining_hours, effort, iteration_id, project_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT(change_date, gh_id) 
DO UPDATE SET
  "name" = EXCLUDED.name,
  "status" = EXCLUDED.status,
  priority = EXCLUDED.priority,
  remaining_hours = EXCLUDED.remaining_hours,
  effort = EXCLUDED.effort,
  iteration_id = EXCLUDED.iteration_id,
  project_id = EXCLUDED.project_id
RETURNING *;

-- name: UpsertProject :one
INSERT INTO project (gh_id, name)
VALUES ($1, $2)
ON CONFLICT(gh_id) 
DO UPDATE SET
  "name" = EXCLUDED.name
RETURNING *;

-- name: UpsertIteration :one
INSERT INTO iteration (gh_id, name, start_date, end_date, project_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(gh_id) 
DO UPDATE SET
  "name" = EXCLUDED.name,
  start_date = EXCLUDED.start_date,
  end_date = EXCLUDED.end_date
RETURNING *;

-- name: UpsertWorkItemStatus :one
INSERT INTO work_item_status (name)
VALUES ($1)
ON CONFLICT(name) 
DO UPDATE SET
  "name" = EXCLUDED.name
RETURNING *;

-- name: GetIterations :many
SELECT id, gh_id, name, start_date, end_date, project_id 
FROM iteration WHERE project_id = $1;

-- name: GetProjects :many
SELECT id, gh_id, name
FROM project;

-- name: GetIterationBurndown :many
WITH starting_effort AS (
 SELECT sum(effort) AS effort
 FROM work_item_history
 WHERE change_date = (SELECT min(change_date) FROM work_item_history WHERE iteration_id = 1)
   AND iteration_id = $1)

SELECT iteration_day
     , cast(sum(case when status <> 'Done' then work_item_history.effort else 0 end) as decimal) as remaining
     , cast(seffort.effort::decimal - (seffort.effort::decimal / total_days.total * row_number() over (order by iteration_day)) as decimal) as ideal
  FROM iteration
       JOIN lateral (SELECT date_trunc('day', dd):: date as iteration_day
                       FROM generate_series
                               ( iteration.start_date::timestamp 
                               , iteration.end_date::timestamp
                               , '1 day'::interval) dd
                      WHERE EXTRACT(ISODOW FROM dd) not IN (6, 7)) dates on true
       JOIN lateral (SELECT count(*)::decimal as total
                       FROM generate_series
                               ( iteration.start_date::timestamp 
                               , iteration.end_date::timestamp
                               , '1 day'::interval) dd
                      WHERE EXTRACT(ISODOW FROM dd) not IN (6, 7)) total_days on true
        JOIN lateral (SELECT effort from starting_effort) seffort on true
       LEFT JOIN work_item_history on work_item_history.change_date = dates.iteration_day
 WHERE iteration.id = $1
 GROUP BY iteration_day, total_days.total, seffort.effort
ORDER BY iteration_day;


-- name: GetProjectBurnup :many
SELECT statuses.name as status
     , project_day
     , sum(work_item_history.effort)::decimal as qty
  FROM work_item_status statuses
       JOIN lateral (SELECT date_trunc('day', dd):: date as project_day
                       FROM generate_series
                               ( $2::timestamp 
                               , now()::timestamp
                               , '1 day'::interval) dd) dates on true
        LEFT JOIN work_item_history on work_item_history.change_date = dates.project_day and work_item_history.status = statuses.name
        LEFT JOIN iteration on work_item_history.iteration_id = iteration.id
 WHERE (iteration.project_id = $1 or iteration.project_id is null)
 GROUP BY statuses.name, dates.project_day
ORDER BY statuses.name, dates.project_day;