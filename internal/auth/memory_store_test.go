package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_MemoryStoreSet(t *testing.T) {
	store := NewMemoryStore()
	err := store.Set("runId", "test-token")
	assert.Nil(t, err)
	check, err := store.Check("runId", "test-token")
	assert.Nil(t, err)
	assert.True(t, check)
}
