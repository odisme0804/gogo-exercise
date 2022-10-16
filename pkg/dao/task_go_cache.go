package dao

import (
	"encoding/gob"
	"errors"
	"sort"
	"strconv"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type goCacheTaskDAO struct {
	logger *zap.SugaredLogger
	cache  *gocache.Cache
}

const (
	cacheKeyNextTaskID = "cacheKeyNextTaskID"
)

func NewGoCacheTaskDAO(logger *zap.SugaredLogger) *goCacheTaskDAO {
	return &goCacheTaskDAO{
		logger: logger,
		cache:  gocache.New(gocache.NoExpiration, 10*time.Minute),
	}
}

func (dao *goCacheTaskDAO) List() ([]Task, error) {
	items := dao.cache.Items()
	tasks := make([]Task, 0, len(items))
	for key, item := range items {
		if key == cacheKeyNextTaskID {
			continue
		}

		task, ok := item.Object.(Task)
		if !ok {
			dao.logger.Errorf("gocache.Items type assertion failed: item.Object.(Task)")
			continue
		}

		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID > tasks[j].ID
	})

	return tasks, nil
}

func (dao *goCacheTaskDAO) GetByID(id int) (Task, error) {
	key := strconv.Itoa(id)
	item, found := dao.cache.Get(key)
	if !found {
		return Task{}, ErrResourceNotFound
	}

	task, ok := item.(Task)
	if !ok {
		dao.logger.Errorf("gocache.Get type assertion failed: item.(Task)")
		return Task{}, errors.New("type assertion failed")
	}

	return task, nil
}

func (dao *goCacheTaskDAO) Create(name string) (Task, error) {
	id, err := dao.cache.IncrementInt64(cacheKeyNextTaskID, 1)
	if err != nil {
		dao.logger.Errorf("gocache.IncrementInt64 failed, err=%v", err)
		return Task{}, err
	}

	task := Task{
		ID:     int(id),
		Name:   name,
		Status: TaskStatusIncomplete,
	}
	key := strconv.Itoa(task.ID)
	dao.cache.SetDefault(key, task)

	return task, nil
}

func (dao *goCacheTaskDAO) Delete(id int) error {
	key := strconv.Itoa(id)
	dao.cache.Delete(key)

	return nil
}

func (dao *goCacheTaskDAO) Update(task *Task) error {
	if task == nil {
		return errors.New("input task is nil")
	}

	key := strconv.Itoa(task.ID)
	if _, found := dao.cache.Get(key); !found {
		return ErrResourceNotFound
	}
	dao.cache.SetDefault(key, *task)

	return nil
}

func (dao *goCacheTaskDAO) Save(filename string) error {
	return dao.cache.SaveFile(filename)
}

func (dao *goCacheTaskDAO) Load(filename string) error {
	gob.Register(Task{})
	if err := dao.cache.LoadFile(filename); err != nil {
		dao.logger.Warnf("gocache.LoadFile failed, err=%v", err)
	}

	tasks, err := dao.List()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		dao.cache.SetDefault(cacheKeyNextTaskID, int64(0))
	}

	return nil
}

var _ TaskDAO = (*goCacheTaskDAO)(nil)
