package taskqueue

import (
	"github.com/beneath-core/beneath-go/taskqueue/task"
)

func init() {
	registerTask(&task.CleanupInstance{})
}
