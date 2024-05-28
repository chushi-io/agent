package auth

type MemoryStore struct {
	tokens map[string]string
}

func NewMemoryStore() MemoryStore {
	return MemoryStore{tokens: map[string]string{}}
}

func (ms MemoryStore) Set(runId string, jwt string) error {
	ms.tokens[runId] = jwt
	return nil
}

func (ms MemoryStore) Check(runId string, jwt string) (bool, error) {
	if _, ok := ms.tokens[runId]; !ok {
		return false, nil
	}
	if jwt != ms.tokens[runId] {
		return false, nil
	}
	return true, nil
}

func (ms MemoryStore) Delete(runId string) error {
	delete(ms.tokens, runId)
	return nil
}
