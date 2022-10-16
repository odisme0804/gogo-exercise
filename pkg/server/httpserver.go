package server

import (
	"errors"
	"gogo-exercise/pkg/dao"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type httpServerImpl struct {
	logger  *zap.SugaredLogger
	addr    string
	taskDAO dao.TaskDAO
}

func NewHttpServer(logger *zap.SugaredLogger, addr string, taskDAO dao.TaskDAO) *http.Server {
	server := &httpServerImpl{
		logger:  logger,
		addr:    addr,
		taskDAO: taskDAO,
	}

	router := gin.Default()
	router.RedirectTrailingSlash = true
	apiRouter := router.Group("/api")
	apiRouter.Use(server.ContentTypeMiddleware())

	tasksRouter := apiRouter.Group("/tasks")
	{
		tasksRouter.GET("", server.ListTasksHandler)
		tasksRouter.POST("", server.CreateTaskHandler)
		tasksRouter.PUT("/:id", server.UpdateTaskHandler)
		tasksRouter.DELETE("/:id", server.DeleteTaskHandler)
	}

	return &http.Server{
		Handler: router,
		Addr:    addr,
	}
}

func (s *httpServerImpl) ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

func (s *httpServerImpl) ListTasksHandler(c *gin.Context) {
	tasks, err := s.taskDAO.List()
	if err != nil {
		s.logger.Errorf("taskDAO.List failed, err=%v", err)
		writeResponseError(c, http.StatusInternalServerError, "something went wrong")
		return
	}

	rsp := ListTasksResponse{
		Result: toModelTasks(tasks),
	}
	c.JSON(http.StatusOK, rsp)
}

func (s *httpServerImpl) CreateTaskHandler(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		writeResponseError(c, http.StatusBadRequest, "parse input failed")
		return
	}

	task, err := s.taskDAO.Create(req.Name)
	if err != nil {
		s.logger.Errorf("taskDAO.Create failed, err=%v", err)
		writeResponseError(c, http.StatusInternalServerError, "something went wrong")
		return
	}

	rsp := CreateTaskResponse{
		Result: toModelTask(task),
	}
	c.JSON(http.StatusCreated, rsp)
}

func (s *httpServerImpl) DeleteTaskHandler(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		writeResponseError(c, http.StatusBadRequest, "parse input failed")
		return
	}

	if err = s.taskDAO.Delete(taskID); err != nil {
		s.logger.Errorf("taskDAO.Delete failed, err=%v, taskID=%v", err, taskID)
		writeResponseError(c, http.StatusInternalServerError, "something went wrong")
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s *httpServerImpl) UpdateTaskHandler(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		writeResponseError(c, http.StatusBadRequest, "parse input failed")
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		writeResponseError(c, http.StatusBadRequest, "parse input failed")
		return
	}

	task := dao.Task{
		ID:     taskID,
		Name:   req.Name,
		Status: dao.TaskStatus(req.Status),
	}
	err = s.taskDAO.Update(&task)
	if errors.Is(err, dao.ErrResourceNotFound) {
		writeResponseError(c, http.StatusNotFound, "task not found")
		return
	} else if err != nil {
		s.logger.Errorf("taskDAO.Update failed, err=%v, taskID=%v", err, taskID)
		writeResponseError(c, http.StatusInternalServerError, "something went wrong")
		return
	}

	rsp := UpdateTaskResponse{
		Result: toModelTask(task),
	}
	c.JSON(http.StatusOK, rsp)
}
