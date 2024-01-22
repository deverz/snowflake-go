package snowflake_go

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

/*
64位
1位符号位-41位毫秒时间戳位-5位数据中心Id位-5位工作机器Id位-12位序号位
[0][0000000-00000000-00000000-00000000-00000000-00][00000][0-0000][0000-00000000]
*/

// 初始时间戳 2024-01-01 00:00:00
var initialTimestamp int64 = 1704038400000

// 机器Id所占位数
var workerIdBits int64 = 5

// 数据中心Id所占位数
var dataCenterIdBits int64 = 5

// 序列号所占位数
var sequenceBits int64 = 12

// 支持的最大机器Id 2^5-1
var maxWorkerId int64 = -1 ^ (-1 << workerIdBits)

// 支持的最大数据中心Id 2^5-1
var maxDataCenterId int64 = -1 ^ (-1 << dataCenterIdBits)

// 支持的最大序号 2^12-1
var maxSequence int64 = -1 ^ (-1 << sequenceBits)

// 机器Id左移位数 - 左移12位为机器位
var workerIdShift int64 = sequenceBits

// 数据中心Id左移位数 - 左移17位为数据中心位
var dataCenterIdShift int64 = workerIdShift + workerIdBits

// 时间戳左移位数 - 左移22位为时间戳位 - 时间戳为 当前时间-基准开始时间
var timestampShift int64 = dataCenterIdShift + dataCenterIdBits

var (
	lastTimeStamp int64 // 最后生成Id的毫秒时间戳
	sequence      int64 // 当前1毫秒内的生成序号
)

var mu sync.Mutex

type Snowflake struct {
	workId       int64
	dataCenterId int64
	err          error
}

// GetInstance 生成实例
func GetInstance(workId, dataCenterId int64) *Snowflake {
	instance := &Snowflake{}
	if workId < 0 || dataCenterId < 0 {
		instance.err = errors.New("workId或dataCenterId不能小于0")
		return instance
	}
	if workId > maxWorkerId || dataCenterId > maxDataCenterId {
		instance.err = fmt.Errorf("workId或dataCenterId不能大于%v或%v", maxWorkerId, maxDataCenterId)
		return instance
	}
	instance.workId = workId
	instance.dataCenterId = dataCenterId
	return instance
}

func (s *Snowflake) GetNextId() (nextId int64, err error) {
	if s.err != nil {
		return 0, s.err
	}
	mu.Lock()
	defer mu.Unlock()
	// 获取当前毫秒时间戳
	currentMilli := time.Now().UnixMilli()
	// 如果当前毫秒时间戳小于最后一次生成id的时间戳，则时钟倒退，无法保证生成Id不重复，报错，不生成新Id
	if currentMilli < lastTimeStamp {
		s.err = errors.New("当前时间戳小于上一次生成Id时间戳")
		return 0, s.err
	}
	// 如果是同一毫秒，则对同一毫秒内排序
	if currentMilli == lastTimeStamp {
		// 判断当前毫秒内的需要是否已经使用完毕
		sequence = (sequence + 1) & maxSequence
		// 为0表示需要用尽，等待下一毫秒
		if sequence == 0 {
			currentMilli = s.nextMilli()
		}
	} else {
		// 非同一毫秒，则需要记录清零
		sequence = 0
	}
	// 记录上次生成的时间戳
	lastTimeStamp = currentMilli
	return ((currentMilli - initialTimestamp) << timestampShift) |
		(s.dataCenterId << dataCenterIdShift) | (s.workId << workerIdShift) | sequence, nil
}

func (s *Snowflake) nextMilli() int64 {
	nextMilli := time.Now().UnixMilli()
	for nextMilli <= lastTimeStamp {
		nextMilli = time.Now().UnixMilli()
	}
	return nextMilli
}
