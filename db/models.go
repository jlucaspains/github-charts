// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Iteration struct {
	ID        int64
	GhID      string
	Name      string
	StartDate pgtype.Date
	EndDate   pgtype.Date
}

type Project struct {
	ID   int64
	GhID string
	Name string
}

type WorkItemHistory struct {
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
}

type WorkItemStatus struct {
	ID   int64
	Name string
}
