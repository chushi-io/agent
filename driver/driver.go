package driver

import (
	"errors"
	"github.com/hashicorp/go-tfe"
)

type Job struct {
	Spec   *JobSpec
	Status *JobStatus
}

type JobSpec struct {
	OrganizationId string
	Image          string
	Run            *tfe.Run
	Workspace      *tfe.Workspace
	Token          string
	Credentials    string
	Variables      []*tfe.Variable
	ConfigVersion  *tfe.ConfigurationVersion
}

func NewJob(spec *JobSpec) *Job {
	return &Job{
		Spec: spec,
		Status: &JobStatus{
			Metadata: map[string]string{},
		},
	}
}

type JobStatus struct {
	Metadata map[string]string
	State    JobState
}

type JobState string

const (
	JobStateRunning  JobState = "running"
	JobStatePending           = "pending"
	JobStateComplete          = "complete"
	JobStateFailed            = "failed"
)

func (job *Job) GetMetadata(key string) (string, error) {
	if val, ok := job.Status.Metadata[key]; ok {
		return val, nil
	}
	return "", errors.New("metadata key not found")
}

type Driver interface {
	Start(job *Job) (*Job, error)
	Wait(job *Job) (*Job, error)
	Cleanup(job *Job) error
}
