package repository

import (
	"errors"
	"sync"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

var ErrTaskNotFound = errors.New("conversion task not found")

type TaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]domain.ConversionTask
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{tasks: map[string]domain.ConversionTask{}}
}

func (repository *TaskRepository) Save(task domain.ConversionTask) domain.ConversionTask {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	repository.tasks[task.ID] = task
	return task
}

func (repository *TaskRepository) FindByID(id string) (domain.ConversionTask, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	task, ok := repository.tasks[id]
	if !ok {
		return domain.ConversionTask{}, ErrTaskNotFound
	}

	return task, nil
}
