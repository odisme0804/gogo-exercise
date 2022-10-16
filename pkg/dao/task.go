package dao

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

type TaskDAO interface {
	List() ([]Task, error)
	GetByID(id int) (Task, error)
	Create(namg string) (Task, error)
	Delete(id int) error
	Update(task *Task) error
}
