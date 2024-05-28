package agent

import "github.com/hashicorp/go-tfe"

type PlanStartedEvent struct {
	Plan *tfe.Plan
}

type PlanCompletedEvent struct {
	Plan *tfe.Plan
}

type PlanFailedEvent struct {
	Plan  *tfe.Plan
	Error error
}

type ApplyStartedEvent struct {
	Apply *tfe.Apply
}

type ApplyCompletedEvent struct {
	Apply *tfe.Apply
}

type ApplyFailedEvent struct {
	Apply *tfe.Apply
	Error error
}
