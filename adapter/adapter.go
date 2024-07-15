package adapter

import "github.com/hashicorp/go-tfe"

type Adapter interface {
	Listen(runHandler)
}

type runHandler func(run *tfe.Run) error
