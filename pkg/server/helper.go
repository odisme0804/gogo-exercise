package server

import (
	"gogo-exercise/pkg/dao"

	"github.com/gin-gonic/gin"
)

func writeResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{
		Message: message,
	})
}

func toModelTask(task dao.Task) Task {
	return Task{
		ID:     task.ID,
		Name:   task.Name,
		Status: TaskStatus(task.Status),
	}
}

func toModelTasks(tasks []dao.Task) []Task {
	retTasks := make([]Task, 0, len(tasks))
	for i := range tasks {
		retTasks = append(retTasks, toModelTask(tasks[i]))
	}
	return retTasks
}
