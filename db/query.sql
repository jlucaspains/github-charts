-- name: GetWorkItemsForIteration :many
SELECT * FROM work_item_history
join iteration on work_item.iteration_id = iteration.id
WHERE iteration.name = $1;

-- name: UpsertWorkItem :one
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
RETURNING *;

-- name: UpsertProject :one
INSERT INTO project (gh_id, name)
VALUES ($1, $2)
ON CONFLICT(gh_id) 
DO UPDATE SET
  "name" = EXCLUDED.name
RETURNING *;

-- name: UpsertIteration :one
INSERT INTO iteration (gh_id, name, start_date, end_date)
VALUES ($1, $2, $3, $4)
ON CONFLICT(gh_id) 
DO UPDATE SET
  "name" = EXCLUDED.name,
  start_date = EXCLUDED.start_date,
  end_date = EXCLUDED.end_date
RETURNING *;

-- name: GetIterationBurndown :many
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
order by iteration_day;