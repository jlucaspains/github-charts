package models

import (
	"time"
)

type Iteration struct {
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type Issue struct {
	Id             string    `json:"id"`
	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"createdAt"`
	ClosedAt       time.Time `json:"closedAt"`
	Status         string    `json:"status"`
	Effort         float64   `json:"effort"`
	RemainingHours float64   `json:"remainingHours"`
	Labels         []string  `json:"labels"`
	IterationId    string    `json:"iterationId"`
}

type Project struct {
	Id         string      `json:"id"`
	Title      string      `json:"title"`
	Issues     []Issue     `json:"issues"`
	Statuses   []string    `json:"statuses"`
	Iterations []Iteration `json:"iterations"`
}

type ErrorResult struct {
	Errors []string `json:"errors"`
}

type HealthResult struct {
	Healthy      bool               `json:"healthy"`
	Dependencies []HealthResultItem `json:"dependencies"`
}

type HealthResultItem struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Error   string `json:"error"`
}
