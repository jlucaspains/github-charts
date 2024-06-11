CREATE TABLE project (
  id                BIGSERIAL PRIMARY KEY,
  gh_id             varchar(255)    NOT NULL,
  name              varchar(255)    NOT NULL,
  UNIQUE(gh_id)
);

CREATE TABLE iteration (
  id                BIGSERIAL PRIMARY KEY,
  gh_id             varchar(255)    NOT NULL,
  name              varchar(255)    NOT NULL,
  start_date        date,
  end_date          date,
  UNIQUE(gh_id)
);

CREATE TABLE work_item_history (
  id                BIGSERIAL PRIMARY KEY,
  change_date       date            NOT NULL,
  gh_id             varchar(255)    NOT NULL,
  project_id        BIGINT          NOT NULL REFERENCES project (id),
  name              varchar(255)   NOT NULL,
  status            varchar(255),
  priority          integer NULL,
  remaining_hours   integer NULL,
  effort            integer NULL,
  iteration_id      BIGINT  NULL REFERENCES iteration (id),
  UNIQUE(change_date, gh_id)
);