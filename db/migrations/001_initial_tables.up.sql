CREATE TABLE project (
  id                SERIAL PRIMARY KEY,
  gh_id             varchar(255)    NOT NULL,
  name              varchar(255)    NOT NULL,
  UNIQUE(gh_id)
);

CREATE TABLE iteration (
  id                SERIAL PRIMARY KEY,
  gh_id             varchar(255)    NOT NULL,
  name              varchar(255)    NOT NULL,
  start_date        date,
  end_date          date,
  project_id        INT  NOT NULL REFERENCES project (id),
  UNIQUE(gh_id)
);

CREATE TABLE work_item_history (
  id                SERIAL PRIMARY KEY,
  change_date       date            NOT NULL,
  gh_id             varchar(255)    NOT NULL,
  name              varchar(255)    NOT NULL,
  status            varchar(255),
  priority          integer NULL,
  remaining_hours   integer NULL,
  effort            integer NULL,
  iteration_id      INT  NULL REFERENCES iteration (id),
  UNIQUE(change_date, gh_id)
);

CREATE TABLE work_item_status (
  id            SMALLSERIAL PRIMARY KEY,
  name          varchar(255)            NOT NULL,
  UNIQUE(name)
);