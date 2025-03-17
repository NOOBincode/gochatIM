package snowflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)


const (
	epoch int64 = 1640995200000 // 2022-01-01 00:00:00 UTC

	workerIDBits     =  10  // 工作机器ID所占用的位数
	sequenceBits     =  12 //序列号所占用的位数

	maxWorkerID      = -1 ^ (-1 << workerIDBits) //工作机器ID的最大值
	maxSequence      = -1 ^ (-1 << sequenceBits) //序列号的最大值

	workerIDShift    = sequenceBits //工作机器ID左移位数
	timestampLeftShift = sequenceBits + workerIDBits //时间戳左移位数
)


type Generator struct {
	workerID int64 // 工作机器ID
	sequence int64 //序列号
	lastTimestamp int64 //上一次生成ID的时间戳
	mu sync.Mutex//互斥锁
}

// NewGenerator 创建一个新的ID生成器
func NewGenerator(workerID int64) (*Generator, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, errors.New("worker ID must be between 0 and maxWorkerID")
	}
	return &Generator{
		workerID: workerID,
		sequence: 0,
		lastTimestamp: 0,
		mu: sync.Mutex{},
	}, nil
}


func (g *Generator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := time.Now().UnixMilli()

	//回退处理
	if timestamp < g.lastTimestamp {
		//等待大于上次生成ID的时间戳的时间
		offset := g.lastTimestamp-timestamp
		if offset <= 5 {
			time.Sleep(time.Millisecond * time.Duration(offset))
			timestamp = time.Now().UnixMilli()
			if timestamp < g.lastTimestamp {
				//依旧小于上个时间戳则使用上个时间戳
				timestamp = g.lastTimestamp
			}
		}else{
			timestamp = g.lastTimestamp
		}
	}
	//同一时间生成则增加序列号递增
	if g.lastTimestamp == timestamp{
		g.sequence = (g.sequence + 1) & maxSequence
		//序列号达到上限则等待下一毫秒
		if g.sequence == 0 {
			for timestamp <= g.lastTimestamp {
				timestamp = time.Now().UnixMilli()
			}
		}
	}else{
		//时间戳改变则重置
		g.sequence = 0

	}
	g.lastTimestamp = timestamp
	//通过移位运算拼接成64位ID
	id := ((timestamp - epoch) << timestampLeftShift) |
		(g.workerID << workerIDShift) |
		g.sequence

	return fmt.Sprintf("%d", id)
}

func (g *Generator) GenerateInt64() int64 {
	//实现方法与生成string 类型的相同
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := time.Now().UnixMilli()

	//回退处理
	if timestamp < g.lastTimestamp {
		//等待大于上次生成ID的时间戳的时间
		offset := g.lastTimestamp-timestamp
		if offset <= 5 {
			time.Sleep(time.Millisecond * time.Duration(offset))
			timestamp = time.Now().UnixMilli()
			if timestamp < g.lastTimestamp {
				//依旧小于上个时间戳则使用上个时间戳
				timestamp = g.lastTimestamp
			}
		}else{
			timestamp = g.lastTimestamp
		}
	}
	//同一时间生成则增加序列号递增
	if g.lastTimestamp == timestamp{
		g.sequence = (g.sequence + 1) & maxSequence
		//序列号达到上限则等待下一毫秒
		if g.sequence == 0 {
			for timestamp <= g.lastTimestamp {
				timestamp = time.Now().UnixMilli()
			}
		}
	}else{
		//时间戳改变则重置
		g.sequence = 0

	}
	g.lastTimestamp = timestamp
	//通过移位运算拼接成64位ID
	id := ((timestamp - epoch) << timestampLeftShift) |
		(g.workerID << workerIDShift) |
		g.sequence

	return id

}

//解析ID,返回时间戳,机器ID和序列号
func Parse(id int64) (timestamp int64,workerID int64,sequence int64) {
	sequence = id & maxSequence
	workerID = (id >> workerIDShift) & maxWorkerID
	timestamp = (id >> timestampLeftShift) + epoch
	return
}

func GetTimestamp(id int64) time.Time{
	timestamp,_,_ := Parse(id)
	return time.UnixMilli(timestamp)
}