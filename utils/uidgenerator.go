package utils

import (
	"log"
	"math"
	"sync"
	"time"

	"errors"
)

type UidGeneratorConfig struct {
	EpochLen    uint8
	SrvLen      uint8
	CntLen      uint8
	IntervalLen uint8
	TruncStr    bool
	EpochStart  int64
	SrvID       int64

	strLen     int8
	epochShift uint8
	epochMask  UniqueID
	epochIota  UniqueID
	maxSrv     UniqueID
	srvMask    UniqueID
	maxCnt     UniqueID
	cntMask    UniqueID
	interval   int64
	timeMask   int64
}

func (cfg *UidGeneratorConfig) update() error {
	idLen := cfg.EpochLen + cfg.SrvLen + cfg.CntLen
	if idLen > 64 {
		return ErrTooLongID
	}

	cfg.strLen = int8(math.Ceil(float64(idLen) / letterLen))
	cfg.epochShift = cfg.SrvLen + cfg.CntLen
	cfg.epochMask = (1<<cfg.EpochLen - 1) << cfg.epochShift
	cfg.epochIota = 1 << cfg.epochShift
	cfg.maxSrv = 1<<cfg.SrvLen - 1
	cfg.srvMask = cfg.maxSrv << cfg.CntLen
	cfg.maxCnt = 1<<cfg.CntLen - 1
	cfg.cntMask = cfg.maxCnt

	if cfg.IntervalLen == 0 {
		cfg.IntervalLen = 61 - cfg.EpochLen
	}

	cfg.interval = 1 << cfg.IntervalLen
	cfg.timeMask = cfg.interval - 1

	return nil
}

var SnowflakeConfig = UidGeneratorConfig{
	EpochLen:    41,
	SrvLen:      10,
	CntLen:      12,
	EpochStart:  1288834974, // 2010-11-04T01:42:54
	IntervalLen: 20,
}

const (
	letters    = "abcdefghijklmnopqrstuvwxyzABCDEF"
	letterLen  = 5
	letterMask = 1<<letterLen - 1
)

var decodeLetters [256]byte

// ErrInvalidStringUID is returned by UidFromString when given an invalid string
var ErrInvalidStringUID = errors.New("invalid encoded UniqueID")

// ErrTooBigServerID is returned by NewUidGenerator when given a server ID bigger than 2^srvLen-1
var ErrTooBigServerID = errors.New("server ID is too big")

// ErrTooLongID is returned by NewUidGenerator if length of IDs would exceed 64 bits
var ErrTooLongID = errors.New("configured ID length is too big")

func init() {
	for i := 0; i < len(letters); i++ {
		decodeLetters[i] = 0xFF
	}

	for i := 0; i < len(letters); i++ {
		decodeLetters[letters[i]] = byte(i)
	}
}

type UniqueID uint64

type UidGenerator struct {
	mu sync.Mutex

	cfg UidGeneratorConfig

	start time.Time
	epoch UniqueID
	srvID UniqueID
	cnt   UniqueID
}

func NewUidGenerator(cfg UidGeneratorConfig, prevID UniqueID) (*UidGenerator, error) {
	err := cfg.update()
	if err != nil {
		return nil, err
	}

	if cfg.SrvID > int64(cfg.maxSrv) {
		return nil, ErrTooBigServerID
	}

	gen := &UidGenerator{
		cfg:   cfg,
		epoch: prevID & cfg.epochMask,
		srvID: UniqueID(cfg.SrvID << cfg.CntLen),
		cnt:   prevID & cfg.cntMask,
	}

	now := time.Now()
	gen.start = now.Add(time.Unix(cfg.EpochStart, 0).Sub(now))

	return gen, nil
}

func (gen *UidGenerator) NextID() UniqueID {
	since := time.Since(gen.start).Nanoseconds() >> gen.cfg.IntervalLen
	epoch := UniqueID(since) << gen.cfg.epochShift & gen.cfg.epochMask

	gen.mu.Lock()

	if epoch <= gen.epoch {
		gen.cnt++
		if gen.cnt > gen.cfg.maxCnt {
			nsec := gen.cfg.interval - int64(time.Now().Nanosecond())&gen.cfg.timeMask
			log.Printf("Exceeded max ID count per interval. Sleeping for %d msec...\n", nsec>>20)
			time.Sleep(time.Duration(nsec) * time.Nanosecond)

			gen.epoch = gen.epoch + gen.cfg.epochIota
			gen.cnt = 0
		}
	} else {
		gen.epoch = epoch
		gen.cnt = 0
	}

	id := gen.epoch + gen.srvID + gen.cnt

	gen.mu.Unlock()

	return id
}

func (gen *UidGenerator) FromBase32(str string) (UniqueID, error) {
	var id UniqueID

	var i int8

	for ; i < int8(len(str)); i++ {
		ch := decodeLetters[str[i]]
		if ch == 0xFF {
			return 0, ErrInvalidStringUID
		}

		id <<= letterLen
		id = id + UniqueID(ch)
	}

	for ; i < gen.cfg.strLen; i++ {
		id <<= letterLen
	}

	return id, nil
}

func (gen *UidGenerator) ToBase32(id UniqueID) string {
	i := gen.cfg.strLen - 1

	if gen.cfg.TruncStr {
		for id&letterMask == 0 {
			id >>= letterLen
			i--
		}
	}

	b := make([]byte, i+1)

	for ; i >= 0; i-- {
		idx := id & letterMask
		b[i] = letters[idx]
		id >>= letterLen
	}

	return string(b)
}

func (gen *UidGenerator) FromUnix(epoch int64) UniqueID {
	return UniqueID(epoch-gen.cfg.EpochStart)*1e9>>gen.cfg.IntervalLen<<gen.cfg.epochShift + gen.srvID
}

func (gen *UidGenerator) Unix(id UniqueID) int64 {
	nsec := int64(id >> gen.cfg.epochShift << gen.cfg.IntervalLen)
	unix := nsec / 1e9

	if nsec%1e9 >= 5e8 {
		unix++
	}

	return unix + gen.cfg.EpochStart
}

func (gen *UidGenerator) FromUnixNano(epoch int64) UniqueID {
	return UniqueID(epoch-gen.cfg.EpochStart*1e9)>>gen.cfg.IntervalLen<<gen.cfg.epochShift + gen.srvID
}

func (gen *UidGenerator) UnixNano(id UniqueID) int64 {
	return int64(id>>gen.cfg.epochShift<<gen.cfg.IntervalLen) + gen.cfg.EpochStart*1e9
}

func (gen *UidGenerator) ServerID(id UniqueID) int64 {
	return int64(id&gen.cfg.srvMask) >> gen.cfg.CntLen
}

func (gen *UidGenerator) Count(id UniqueID) int64 {
	return int64(id & gen.cfg.cntMask)
}
