package utils

import (
	"sync"
	"time"
)

/*
* Snowflake
*
* 1                                               42           52             64
* +-----------------------------------------------+------------+---------------+
* | timestamp(ms)                                 | workerID   | sequence      |
* +-----------------------------------------------+------------+---------------+
* | 0000000000 0000000000 0000000000 0000000000 0 | 0000000000 | 0000000000 00 |
* +-----------------------------------------------+------------+---------------+
*
* 1. 41位时间截(毫秒级)，注意这是时间截的差值（当前时间截 - 开始时间截)。可以使用约70年: (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69
* 2. 10位数据机器位，可以部署在1024个节点
* 3. 12位序列，毫秒内的计数，同一机器，同一时间截并发4096个序号
 */

const (
	epoch          = int64(1483228800000)             //开始时间截 (2017-01-01)
	workerIDBits   = uint(10)                         //机器id所占的位数
	sequenceBits   = uint(12)                         //序列所占的位数
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits)) //支持的最大序列id数量
	workerIDShift  = sequenceBits                     //机器id左移位数
	timestampShift = sequenceBits + workerIDBits      //时间戳左移位数
)

// A Snowflake struct holds the basic information needed for a snowflake generator worker
type Snowflake struct {
	sync.Mutex
	timestamp int64
	workerID  int64
	sequence  int64
}

// GetSnowflake returns a single instance of snowflake worker that can be used to generate snowflake IDs
func GetSnowflake() *Snowflake {
	snowflakeSync.Do(func() {
		snowflake = &Snowflake{workerID: Epoch()}
	})
	return snowflake
}

var (
	snowflake     *Snowflake
	snowflakeSync sync.Once
)

// Generate creates and returns a unique snowflake ID
func (s *Snowflake) Generate() int64 {
	s.Lock()
	defer s.Unlock()
	now := time.Now().UnixNano() / 1000000

	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		s.sequence = 0
	}
	s.timestamp = now
	id := (now-epoch)<<timestampShift | (s.workerID << workerIDShift) | (s.sequence)
	return id
}
