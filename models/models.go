package models

import (
	"fmt"
	"strings"
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

type BurndownItem struct {
	IterationDay time.Time `json:"iterationDay"`
	Remaining    float64   `json:"remaining"`
	Ideal        float64   `json:"ideal"`
}

type BurnupItem struct {
	Status     string    `json:"status"`
	ProjectDay time.Time `json:"projectDay"`
	Qty        float64   `json:"qty"`
}

type JobConfigItem struct {
	OrgName   string
	RepoOwner string
	RepoName  string
	Project   string
	Token     string
}

func (j *JobConfigItem) GetUniqueName() string {
	if j.OrgName != "" {
		return fmt.Sprintf("%s/%s", j.OrgName, j.Project)
	} else {
		return fmt.Sprintf("%s/%s/%s", j.RepoOwner, j.RepoName, j.Project)
	}
}

func (j *JobConfigItem) Validate() error {
	errors := []string{}

	if j.OrgName == "" && j.RepoOwner == "" && j.RepoName == "" {
		errors = append(errors, "org or repo information is required")
	}

	if j.Project == "" {
		errors = append(errors, "project is required")
	}

	if j.Token == "" {
		errors = append(errors, "token is required")
	}

	if len(errors) == 0 {
		return nil
	}

	return fmt.Errorf("invalid configuration: %v", strings.Join(errors, ", "))
}
