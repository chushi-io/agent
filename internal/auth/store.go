package auth

type Store interface {
	Set(runId string, jwt string) error
	Check(runId string, jwt string) (bool, error)
	Delete(runId string) error
}
