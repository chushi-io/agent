package listener

type Listener interface {
	Listen(runHandler)
}

type Event struct {
	OrganizationId string
	RunId          string
	WorkspaceId    string
}

type runHandler func(event *Event) error
