package repository_test

import (
	"testing"

	"github.com/rtcheap/turn-server/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestKeyMap(t *testing.T) {
	assert := assert.New(t)
	repo := repository.NewKeyRepository()

	val, ok := repo.Find("user-1")
	assert.False(ok)
	assert.Nil(val)

	key := "my-key"
	err := repo.Save("user-1", []byte(key))
	assert.NoError(err)

	val, ok = repo.Find("user-1")
	assert.True(ok)
	assert.Equal(key, string(val))

	val, ok = repo.Find("user-2")
	assert.False(ok)
	assert.Nil(val)
}
