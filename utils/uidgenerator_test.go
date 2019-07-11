package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUidGenerator(t *testing.T) {
	req := require.New(t)

	cfg := DefaultUidGeneratorConfig
	cfg.SrvID = 2

	gen, err := NewUidGenerator(cfg, 0)
	req.NotNil(err)

	cfg.SrvID = 1
	gen, err = NewUidGenerator(cfg, 0)
	req.Nil(err)

	var prevID UniqueID

	for i := 0; i < (1<<cfg.CntLen)*3; i++ {
		id := gen.NextID()
		req.True(id > prevID)
		req.Equal(int64(1), gen.ServerID(id))

		str := gen.ToBase32(id)
		decID, err := gen.FromBase32(str)
		req.Nil(err)
		req.Equal(id, decID)

		prevID = id
	}

	gen, err = NewUidGenerator(cfg, prevID)
	req.Nil(err)

	id := gen.NextID()
	req.True(id > prevID)
	req.Equal(int64(1), gen.ServerID(id))

	cfg.EpochLen = 33
	gen, err = NewUidGenerator(cfg, 0)
	req.Nil(err)

	id = gen.NextID()
	str := gen.ToBase32(id)
	decID, err := gen.FromBase32(str)
	req.Nil(err)
	req.Equal(id, decID)
}

func BenchmarkUidGenerator(b *testing.B) {
	cfg := DefaultUidGeneratorConfig
	cfg.EpochLen = 40
	cfg.SrvLen = 0
	cfg.CntLen = 23

	gen, _ := NewUidGenerator(cfg, 0)

	for i := 0; i < b.N; i++ {
		gen.NextID()
	}
}
