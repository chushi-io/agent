package driver

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-tfe"
	"io"
	"os"
	"path/filepath"
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
	ProxyToken     string
	Operation      string
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

func downloadConfigurationVersion(client *tfe.Client, run *tfe.Run) (string, error) {
	data, err := client.ConfigurationVersions.Download(context.TODO(), run.ConfigurationVersion.ID)
	if err != nil {
		return "", err
	}
	r := bytes.NewReader(data)
	uncompressedStream, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}

	tarReader := tar.NewReader(uncompressedStream)
	dname, err := os.MkdirTemp(os.TempDir(), "chushi")

	if err != nil {
		return "", err
	}

	var topLevelDir string
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch header.Typeflag {

		// For some reason, downloading the tar from github returns this header
		// I don't know why it does, but it causes unexpected failures
		// TODO: Write a better reason for doing this
		case tar.TypeXGlobalHeader:
			continue
		case tar.TypeDir:
			if topLevelDir == "" {
				topLevelDir = header.Name
			}
			if err := os.Mkdir(filepath.Join(dname, header.Name), 0755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			archiveFilePath := filepath.Join(dname, header.Name)
			if err := os.MkdirAll(filepath.Dir(archiveFilePath), 0770); err != nil {
				return "", err
			}
			outFile, err := os.Create(archiveFilePath)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return "", err
			}
			outFile.Close()

		default:
			return "", errors.New(fmt.Sprintf("ExtractTarGz: uknown type: %s in %s", header.Typeflag, header.Name))
		}
	}

	return filepath.Join(dname, topLevelDir), nil
}
