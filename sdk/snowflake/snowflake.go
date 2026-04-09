package snowflake

import (
	"hash/fnv"
	"sync"
	"time"
)

const (
	customEpochMs int64 = 1704067200000 // 2024-01-01 00:00:00 UTC
	workerBits    uint8 = 10
	sequenceBits  uint8 = 12

	workerMax   int64 = -1 ^ (-1 << workerBits)
	sequenceMax int64 = -1 ^ (-1 << sequenceBits)

	workerShift    uint8 = sequenceBits
	timestampShift uint8 = workerBits + sequenceBits
)

type Generator struct {
	mu           sync.Mutex
	workerID     int64
	lastTimeMs   int64
	sequence     int64
}

func NewGenerator(workerID int64) *Generator {
	if workerID < 0 {
		workerID = 0
	}
	workerID = workerID & workerMax
	return &Generator{workerID: workerID}
}

func NewGeneratorByNode(node string) *Generator {
	h := fnv.New32a()
	_, _ = h.Write([]byte(node))
	workerID := int64(h.Sum32()) & workerMax
	return NewGenerator(workerID)
}

func (g *Generator) NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	nowMs := time.Now().UnixMilli()
	if nowMs < g.lastTimeMs {
		nowMs = g.lastTimeMs
	}

	if nowMs == g.lastTimeMs {
		g.sequence = (g.sequence + 1) & sequenceMax
		if g.sequence == 0 {
			for nowMs <= g.lastTimeMs {
				nowMs = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTimeMs = nowMs
	return ((nowMs - customEpochMs) << timestampShift) | (g.workerID << workerShift) | g.sequence
}
