package snowflake_go

import (
	"sync"
	"testing"
)

func TestSingleGetSnowflakeId(t *testing.T) {
	id, err := GetInstance(1, 1).GetNextId()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(id)
}

func TestGetSnowflakeId(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i <= 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := GetInstance(10, 10).GetNextId()
			if err != nil {
				t.Error(err)
				return
			}
			t.Log(id)
		}()
	}
	wg.Wait()
}
