package dao

import (
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v5"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var _ = Describe("GoCacheTaskDAO", func() {
	var (
		dao    *goCacheTaskDAO
		logger *zap.SugaredLogger
	)

	BeforeEach(func() {
		logger = zap.NewNop().Sugar()
		dao = &goCacheTaskDAO{
			logger: logger,
			cache:  gocache.New(gocache.NoExpiration, 10*time.Minute),
		}
		dao.cache.SetDefault(cacheKeyNextTaskID, int64(0))
	})

	AfterEach(func() {
		err := logger.Sync()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("List", func() {
		var (
			tasks []Task
			err   error
		)

		var (
			dbTasks []Task
		)

		JustBeforeEach(func() {
			tasks, err = dao.List()
		})

		Context("have some tasks", func() {
			BeforeEach(func() {
				dbTasks = make([]Task, 0, 10)
				for i := 0; i < 10; i++ {
					dbTasks = append(dbTasks, Task{
						ID:     i,
						Name:   gofakeit.Noun(),
						Status: TaskStatus(rand.Int() % 2),
					})
				}

				sort.Slice(dbTasks, func(i, j int) bool {
					return dbTasks[i].ID > dbTasks[j].ID
				})

				for i := range dbTasks {
					dao.cache.SetDefault(strconv.Itoa(dbTasks[i].ID), dbTasks[i])
				}
			})

			AfterEach(func() {
				for i := range dbTasks {
					dao.cache.Delete(strconv.Itoa(dbTasks[i].ID))
				}
			})

			It("should return tasks", func() {
				tasks := make([]interface{}, len(dbTasks))
				for i := range tasks {
					tasks[i] = dbTasks[i]
				}
				Expect(tasks).Should(Equal(tasks))
			})

			It("should not get an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("have no task", func() {
			It("should return empty list", func() {
				Expect(tasks).Should(BeEmpty())
			})

			It("should not get an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("GetByID", func() {
		var (
			taskID int
			task   Task
			err    error
		)

		var (
			cacheTask Task
		)

		JustBeforeEach(func() {
			task, err = dao.GetByID(taskID)
		})

		Context("task exists", func() {
			BeforeEach(func() {
				cacheTask = Task{
					ID:     rand.Int(),
					Name:   gofakeit.Noun(),
					Status: TaskStatus(rand.Int() % 2),
				}
				dao.cache.SetDefault(strconv.Itoa(cacheTask.ID), cacheTask)

				taskID = cacheTask.ID
			})

			AfterEach(func() {
				dao.cache.Delete(strconv.Itoa(cacheTask.ID))
			})

			It("should return task", func() {
				Expect(task).To(Equal(cacheTask))
			})

			It("should not get an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("task not exists", func() {
			BeforeEach(func() {
				taskID = rand.Int()
			})

			It("should get ErrResourceNotFound", func() {
				Expect(err).To(Equal(ErrResourceNotFound))
			})
		})
	})

	Describe("Create", func() {
		var (
			taskName string
			task     Task
			err      error
		)

		JustBeforeEach(func() {
			task, err = dao.Create(taskName)
		})

		Context("create task successfully", func() {
			BeforeEach(func() {
				taskName = gofakeit.Noun()
			})

			AfterEach(func() {
				dao.cache.Delete(strconv.Itoa(task.ID))
			})

			It("should create a task", func() {
				Expect(task.ID).To(BeNumerically(">", 0))
				Expect(task.Name).To(Equal(taskName))
				Expect(task.Status).To(Equal(TaskStatusIncomplete))
			})

			It("should not get an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {
		var (
			taskID int
			err    error
		)

		JustBeforeEach(func() {
			err = dao.Delete(taskID)
		})

		Context("task exists", func() {
			BeforeEach(func() {
				cacheTask := Task{
					ID:     rand.Int(),
					Name:   gofakeit.Noun(),
					Status: TaskStatus(rand.Int() % 2),
				}

				dao.cache.SetDefault(strconv.Itoa(cacheTask.ID), cacheTask)

				taskID = cacheTask.ID
			})

			AfterEach(func() {
				dao.cache.Delete(strconv.Itoa(taskID))
			})

			It("should delete task from storage", func() {
				_, found := dao.cache.Get(strconv.Itoa(taskID))
				Expect(found).To(BeFalse())
			})

			It("should not get an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("task not exists", func() {
			BeforeEach(func() {
				taskID = rand.Int()
			})

			It("should not get an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Update", func() {
		var (
			task *Task
			err  error
		)

		JustBeforeEach(func() {
			err = dao.Update(task)
		})

		Context("task exists", func() {
			BeforeEach(func() {
				task = &Task{
					ID:     rand.Int(),
					Name:   gofakeit.Noun(),
					Status: TaskStatusIncomplete,
				}

				dao.cache.SetDefault(strconv.Itoa(task.ID), task)

				task.Name = task.Name + "_v2"
				task.Status = TaskStatusComplete
			})

			AfterEach(func() {
				dao.cache.Delete(strconv.Itoa(task.ID))
			})

			It("should update task", func() {
				item, found := dao.cache.Get(strconv.Itoa(task.ID))
				Expect(found).To(BeTrue())

				cacheTask, ok := item.(Task)
				Expect(ok).To(BeTrue())

				Expect(cacheTask).Should(Equal(*task))
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("task not exists", func() {
			It("should get ErrResourceNotFound", func() {
				Expect(err).To(Equal(ErrResourceNotFound))
			})
		})
	})
})
