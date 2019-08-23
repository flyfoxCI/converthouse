package storage

import (
	"fmt"
	"os"
	"testing"
	"time"
	"github.com/infinivision/prophet"
	"github.com/stretchr/testify/assert"
)

func create(t *testing.T) prophet.LocalStorage {
	dir := fmt.Sprintf("%s/converthouse-%d", os.TempDir(), time.Now().Nanosecond())
	s, err := NewBadgerStorage(dir)
	assert.Nilf(t, err, "check badger failed with %+v", err)

	return s
}

func TestGetAndSet(t *testing.T) {
	s := create(t)
	value, err := s.Get([]byte("test-key"))
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, 0, len(value), "check storage failed")

	err = s.Set([]byte("test-key"), []byte("converthouse"))
	assert.Nilf(t, err, "check badger storage failed with %+v", err)

	value, err = s.Get([]byte("test-key"))
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, "converthouse", string(value), "check storage failed")

	s.Set([]byte("test-key"), []byte("converthouse2"))
	value, err = s.Get([]byte("test-key"))
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, "converthouse2", string(value), "check storage failed")
}

func TestRemove(t *testing.T) {
	s := create(t)
	s.Set([]byte("test-key"), []byte("converthouse"))

	value, err := s.Get([]byte("test-key"))
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, "converthouse", string(value), "check storage failed")

	err = s.Remove([]byte("test-key"))
	assert.Nilf(t, err, "check badger storage failed with %+v", err)

	value, err = s.Get([]byte("test-key"))
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, 0, len(value), "check storage failed")
}

func TestRange(t *testing.T) {
	s := create(t)
	s.Set([]byte("test-"), []byte("converthouse"))
	s.Set([]byte("test-02"), []byte("converthouse"))
	s.Set([]byte("test-03"), []byte("converthouse"))
	s.Set([]byte("test-04"), []byte("converthouse"))
	s.Set([]byte("test-05"), []byte("converthouse"))

	c := 0
	fn := func(key, value []byte) bool {
		c++
		return true
	}
	err := s.Range([]byte("test-"), 1, fn)
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, 1, c, "check storage failed")

	c = 0
	err = s.Range([]byte("test-"), 5, fn)
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, 5, c, "check storage failed")

	c = 0
	err = s.Range([]byte("test-"), 6, fn)
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, 5, c, "check storage failed")

	c = 0
	err = s.Range([]byte("test-"), 0, fn)
	assert.Nilf(t, err, "check badger storage failed with %+v", err)
	assert.Equal(t, 5, c, "check storage failed")
}
