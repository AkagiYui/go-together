package cache

import (
	"time"

	"github.com/akagiyui/go-together/common/task"
)

func init() {
	task.Run(func() {
		println("Purge expire cache every 1 minute.")
		for {
			if err := PurgeExpired(); err != nil {
				println("purge expired cache error:", err)
			}
			time.Sleep(1 * time.Minute)
		}
	})
}
