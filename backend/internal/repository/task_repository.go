package repository

import (
	"errors"
	"slices"
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

func (repository *TaskRepository) FindAll() []domain.ConversionTask {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	tasks := make([]domain.ConversionTask, 0, len(repository.tasks))
	for _, task := range repository.tasks {
		tasks = append(tasks, task)
	}
	slices.SortFunc(tasks, func(left domain.ConversionTask, right domain.ConversionTask) int {
		return right.CreatedAt.Compare(left.CreatedAt)
	})

	return tasks
}
