package mr

import "fmt"

func getFinalMapName(taskNum int) string {
	return fmt.Sprintf("temp-file-%d", taskNum)
}
