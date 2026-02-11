package sonyflake

import (
	"fmt"
	"sync"
	"time"

	"github.com/sony/sonyflake/v2"
)

var (
	sf   *sonyflake.Sonyflake
	once  sync.Once
)


func InitSonyFlake() error {
	var initErr error
	once.Do(func() {
		var st sonyflake.Settings
		st.StartTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		
		var err error
		sf, err = sonyflake.New(st)
		if err != nil {
			initErr = fmt.Errorf("failed to initialize Sonyflake: %w", err)
		}
	})
	return initErr
}

func GenerateSonyFlakeID() (int64, error) {
	if sf == nil {
		return 0, fmt.Errorf("sonyflake not initialized - call InitSonyFlake first")
	}
	
	id, err := sf.NextID()
	if err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}
	
	return int64(id), nil
}