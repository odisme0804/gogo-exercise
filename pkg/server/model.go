package server

type ListTasksRequest struct {
}

type ListTasksResponse struct {
	Result []Task `json:"result"`
}

type CreateTaskRequest struct {
	Name string `json:"name"`
}

type CreateTaskResponse struct {
	Result Task `json:"result"`
}

type UpdateTaskRequest struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type UpdateTaskResponse struct {
	Result Task `json:"result"`
}

type DeleteTaskRequest struct {
	ID int `json:"id"`
}

type DeleteTaskResponse struct {
}

type Task struct {
	ID     int        `json:"id"`
	Name   string     `json:"name"`
	Status TaskStatus `json:"status"`
}

type TaskStatus int

const (
	TaskStatusIncomplete TaskStatus = iota
	TaskStatusComplete
)

type ErrorResponse struct {
	Message string `json:"message"`
}
