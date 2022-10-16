package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo-exercise/internal/mock/daomock"
	"gogo-exercise/pkg/dao"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v5"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("HttpServer", func() {
	var ctrl *gomock.Controller
	var taskDAO *daomock.MockTaskDAO
	var server *http.Server

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		taskDAO = daomock.NewMockTaskDAO(ctrl)
		server = NewHttpServer(zap.NewNop().Sugar(), "", taskDAO)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("ListTasks", func() {
		var (
			req *http.Request
			rsp *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			rsp = httptest.NewRecorder()
			server.Handler.ServeHTTP(rsp, req)
		})

		Context("normal case", func() {
			var (
				tasks []dao.Task
			)

			BeforeEach(func() {
				var err error
				req, err = http.NewRequest(http.MethodGet, "/api/tasks", nil)
				Expect(err).NotTo(HaveOccurred())

				tasks = make([]dao.Task, 3)
				for i := range tasks {
					tasks = append(tasks, dao.Task{
						ID:     i,
						Name:   gofakeit.Noun(),
						Status: dao.TaskStatus(rand.Int() % 2),
					})
				}

				taskDAO.EXPECT().List().Return(tasks, nil)
			})

			It("should get tasks", func() {
				var listResp ListTasksResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &listResp)
				Expect(err).NotTo(HaveOccurred())
				Expect(listResp.Result).To(Equal(toModelTasks(tasks)))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 200", func() {
				Expect(rsp.Code).To(Equal(http.StatusOK))
			})
		})

		Context("dao error", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest(http.MethodGet, "/api/tasks", nil)
				Expect(err).NotTo(HaveOccurred())

				taskDAO.EXPECT().List().Return(nil, errors.New("dao error"))
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("something went wrong"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 500", func() {
				Expect(rsp.Code).To(Equal(http.StatusInternalServerError))
			})
		})

	})

	Describe("CreateTaskHandler", func() {
		var (
			req *http.Request
			rsp *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			rsp = httptest.NewRecorder()
			server.Handler.ServeHTTP(rsp, req)
		})

		Context("normal case", func() {
			var (
				dbTask dao.Task
			)

			BeforeEach(func() {
				createReq := CreateTaskRequest{
					Name: gofakeit.Noun(),
				}
				requestByte, err := json.Marshal(createReq)
				Expect(err).NotTo(HaveOccurred())

				req, err = http.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(requestByte))
				Expect(err).NotTo(HaveOccurred())

				dbTask = dao.Task{
					ID:     rand.Int(),
					Name:   createReq.Name,
					Status: dao.TaskStatusIncomplete,
				}
				taskDAO.EXPECT().Create(createReq.Name).Return(dbTask, nil)
			})

			It("should get the created task", func() {
				var createRsp CreateTaskResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &createRsp)
				Expect(err).NotTo(HaveOccurred())
				Expect(createRsp.Result).To(Equal(toModelTask(dbTask)))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 201", func() {
				Expect(rsp.Code).To(Equal(http.StatusCreated))
			})
		})

		Context("invalid body", func() {
			BeforeEach(func() {
				var err error
				req, err = http.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader("invalid body"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("parse input failed"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 400", func() {
				Expect(rsp.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("dao error", func() {
			BeforeEach(func() {
				createReq := CreateTaskRequest{
					Name: gofakeit.Noun(),
				}
				requestByte, err := json.Marshal(createReq)
				Expect(err).NotTo(HaveOccurred())

				req, err = http.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(requestByte))
				Expect(err).NotTo(HaveOccurred())

				taskDAO.EXPECT().Create(createReq.Name).Return(dao.Task{}, errors.New("dao error"))
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("something went wrong"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 500", func() {
				Expect(rsp.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("DeleteTaskHandler", func() {
		var (
			req *http.Request
			rsp *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			rsp = httptest.NewRecorder()
			server.Handler.ServeHTTP(rsp, req)
		})

		Context("normal case", func() {
			BeforeEach(func() {
				taskID := rand.Int()
				var err error
				url := fmt.Sprintf("/api/tasks/%d", taskID)
				req, err = http.NewRequest(http.MethodDelete, url, nil)
				Expect(err).NotTo(HaveOccurred())

				taskDAO.EXPECT().Delete(taskID).Return(nil)
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 204", func() {
				Expect(rsp.Code).To(Equal(http.StatusNoContent))
			})
		})

		Context("invalid id", func() {
			BeforeEach(func() {
				var err error
				url := "/api/tasks/nan"
				req, err = http.NewRequest(http.MethodDelete, url, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("parse input failed"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 400", func() {
				Expect(rsp.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("dao error", func() {
			BeforeEach(func() {
				taskID := rand.Int()
				var err error
				url := fmt.Sprintf("/api/tasks/%d", taskID)
				req, err = http.NewRequest(http.MethodDelete, url, nil)
				Expect(err).NotTo(HaveOccurred())

				taskDAO.EXPECT().Delete(taskID).Return(errors.New("dao error"))
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("something went wrong"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 500", func() {
				Expect(rsp.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("UpdateTaskHandler", func() {
		var (
			req *http.Request
			rsp *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			rsp = httptest.NewRecorder()
			server.Handler.ServeHTTP(rsp, req)
		})

		Context("normal case", func() {
			var (
				dbTask dao.Task
			)

			BeforeEach(func() {
				taskID := rand.Int()
				reqBody := UpdateTaskRequest{
					Name:   gofakeit.Noun(),
					Status: rand.Int() % 2,
				}
				requestByte, err := json.Marshal(reqBody)
				Expect(err).NotTo(HaveOccurred())

				url := fmt.Sprintf("/api/tasks/%d", taskID)
				req, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(requestByte))
				Expect(err).NotTo(HaveOccurred())

				dbTask = dao.Task{
					ID:     taskID,
					Name:   reqBody.Name,
					Status: dao.TaskStatus(reqBody.Status),
				}
				taskDAO.EXPECT().Update(&dbTask).Return(nil)
			})

			It("should get the updated task", func() {
				var updateRsp UpdateTaskResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &updateRsp)
				Expect(err).NotTo(HaveOccurred())
				Expect(updateRsp.Result).To(Equal(toModelTask(dbTask)))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 200", func() {
				Expect(rsp.Code).To(Equal(http.StatusOK))
			})
		})

		Context("invalid id", func() {
			BeforeEach(func() {
				reqBody := UpdateTaskRequest{
					Name:   gofakeit.Noun(),
					Status: rand.Int() % 2,
				}
				requestByte, err := json.Marshal(reqBody)
				Expect(err).NotTo(HaveOccurred())

				url := "/api/tasks/nan"
				req, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(requestByte))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("parse input failed"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 400", func() {
				Expect(rsp.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("invalid body", func() {
			BeforeEach(func() {
				var err error
				taskID := rand.Int()
				url := fmt.Sprintf("/api/tasks/%d", taskID)
				req, err = http.NewRequest(http.MethodPut, url, strings.NewReader("invalid body"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("parse input failed"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 400", func() {
				Expect(rsp.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("task not exist", func() {
			BeforeEach(func() {
				taskID := rand.Int()
				reqBody := UpdateTaskRequest{
					Name:   gofakeit.Noun(),
					Status: rand.Int() % 2,
				}
				requestByte, err := json.Marshal(reqBody)
				Expect(err).NotTo(HaveOccurred())

				url := fmt.Sprintf("/api/tasks/%d", taskID)
				req, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(requestByte))
				Expect(err).NotTo(HaveOccurred())

				taskDAO.EXPECT().Update(gomock.Any()).Return(dao.ErrResourceNotFound)
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("task not found"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 404", func() {
				Expect(rsp.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("dao error", func() {
			BeforeEach(func() {
				taskID := rand.Int()
				reqBody := UpdateTaskRequest{
					Name:   gofakeit.Noun(),
					Status: rand.Int() % 2,
				}
				requestByte, err := json.Marshal(reqBody)
				Expect(err).NotTo(HaveOccurred())

				url := fmt.Sprintf("/api/tasks/%d", taskID)
				req, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(requestByte))
				Expect(err).NotTo(HaveOccurred())

				taskDAO.EXPECT().Update(gomock.Any()).Return(errors.New("dao error"))
			})

			It("should get error message", func() {
				var body ErrorResponse
				err := json.Unmarshal(rsp.Body.Bytes(), &body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body.Message).To(Equal("something went wrong"))
			})

			It("should get content-type header", func() {
				Expect(rsp.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("should get status code 500", func() {
				Expect(rsp.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})
})

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "server Suite")
}
