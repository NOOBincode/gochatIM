package lock

import (
	"GochatIM/pkg/redis"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

var Redsync *redsync.Redsync

func Setup() {
	pool := goredis.NewPool(redis.GetClient()) //创建连接池
	Redsync = redsync.New(pool)                //创建redsync
}

func AcquireLock(name string, expiry time.Duration) (*redsync.Mutex, bool, error) {
	mutex := Redsync.NewMutex(name, redsync.WithExpiry(expiry), redsync.WithTries(3), redsync.WithRetryDelay(time.Millisecond*100))
	err := mutex.Lock()
	if err != nil {
		return nil, false, err
	}
	return mutex, true, nil
}

func ReleaseLock(mutex *redsync.Mutex) (bool, error) {
	ok, err := mutex.Unlock()
	return ok, err
}
