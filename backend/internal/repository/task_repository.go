package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"slices"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
)

var ErrTaskNotFound = errors.New("conversion task not found")

type TaskRepository struct {
	mu      sync.RWMutex
	tasks   map[string]domain.ConversionTask
	storage taskStorage
}

type taskStorage interface {
	Save(context.Context, domain.ConversionTask) error
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{tasks: map[string]domain.ConversionTask{}}
}

func NewPostgresBackedTaskRepository(ctx context.Context, databaseURL string) (*TaskRepository, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	store := postgresTaskStorage{db: db}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := store.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}

	tasks, err := store.LoadAll(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	repository := &TaskRepository{
		tasks:   map[string]domain.ConversionTask{},
		storage: store,
	}
	for _, task := range tasks {
		repository.tasks[task.ID] = task
	}

	return repository, nil
}

func (repository *TaskRepository) Save(task domain.ConversionTask) domain.ConversionTask {
	repository.mu.Lock()
	repository.tasks[task.ID] = task
	repository.mu.Unlock()

	if repository.storage != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := repository.storage.Save(ctx, task); err != nil {
			log.Printf("persist conversion task %s: %v", task.ID, err)
		}
	}

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

type postgresTaskStorage struct {
	db *sql.DB
}

func (storage postgresTaskStorage) migrate(ctx context.Context) error {
	_, err := storage.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS conversion_tasks (
	id TEXT PRIMARY KEY,
	status TEXT NOT NULL,
	progress INTEGER NOT NULL,
	stage TEXT NOT NULL,
	source_text TEXT NOT NULL DEFAULT '',
	chapters JSONB NOT NULL DEFAULT '[]'::jsonb,
	ai_config JSONB NOT NULL DEFAULT '{}'::jsonb,
	generation_started BOOLEAN NOT NULL DEFAULT false,
	total_chunks INTEGER NOT NULL DEFAULT 0,
	completed_chunks INTEGER NOT NULL DEFAULT 0,
	current_chunk TEXT NOT NULL DEFAULT '',
	draft JSONB,
	yaml TEXT NOT NULL DEFAULT '',
	error_message TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_conversion_tasks_created_at
	ON conversion_tasks (created_at DESC);
`)
	return err
}

func (storage postgresTaskStorage) Save(ctx context.Context, task domain.ConversionTask) error {
	chapters, err := json.Marshal(task.Chapters)
	if err != nil {
		return err
	}

	aiConfig := task.AIConfig
	if task.Status == domain.TaskStatusCompleted || task.Status == domain.TaskStatusFailed {
		aiConfig.APIKey = ""
	}
	aiConfigJSON, err := json.Marshal(aiConfig)
	if err != nil {
		return err
	}

	var draftJSON []byte
	if task.Draft != nil {
		draftJSON, err = json.Marshal(task.Draft)
		if err != nil {
			return err
		}
	}

	_, err = storage.db.ExecContext(ctx, `
INSERT INTO conversion_tasks (
	id,
	status,
	progress,
	stage,
	source_text,
	chapters,
	ai_config,
	generation_started,
	total_chunks,
	completed_chunks,
	current_chunk,
	draft,
	yaml,
	error_message,
	created_at,
	updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, $9, $10, $11, $12::jsonb, $13, $14, $15, $16
) ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status,
	progress = EXCLUDED.progress,
	stage = EXCLUDED.stage,
	source_text = EXCLUDED.source_text,
	chapters = EXCLUDED.chapters,
	ai_config = EXCLUDED.ai_config,
	generation_started = EXCLUDED.generation_started,
	total_chunks = EXCLUDED.total_chunks,
	completed_chunks = EXCLUDED.completed_chunks,
	current_chunk = EXCLUDED.current_chunk,
	draft = EXCLUDED.draft,
	yaml = EXCLUDED.yaml,
	error_message = EXCLUDED.error_message,
	created_at = EXCLUDED.created_at,
	updated_at = EXCLUDED.updated_at
`,
		task.ID,
		string(task.Status),
		task.Progress,
		task.Stage,
		task.SourceText,
		string(chapters),
		string(aiConfigJSON),
		task.GenerationStarted,
		task.TotalChunks,
		task.CompletedChunks,
		task.CurrentChunk,
		nullableJSONString(draftJSON),
		task.YAML,
		task.ErrorMessage,
		task.CreatedAt,
		task.UpdatedAt,
	)
	return err
}

func (storage postgresTaskStorage) LoadAll(ctx context.Context) ([]domain.ConversionTask, error) {
	rows, err := storage.db.QueryContext(ctx, `
SELECT
	id,
	status,
	progress,
	stage,
	source_text,
	chapters::text,
	ai_config::text,
	generation_started,
	total_chunks,
	completed_chunks,
	current_chunk,
	COALESCE(draft::text, 'null'),
	yaml,
	error_message,
	created_at,
	updated_at
FROM conversion_tasks
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []domain.ConversionTask{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (domain.ConversionTask, error) {
	var task domain.ConversionTask
	var status string
	var chaptersJSON string
	var aiConfigJSON string
	var draftJSON string

	err := scanner.Scan(
		&task.ID,
		&status,
		&task.Progress,
		&task.Stage,
		&task.SourceText,
		&chaptersJSON,
		&aiConfigJSON,
		&task.GenerationStarted,
		&task.TotalChunks,
		&task.CompletedChunks,
		&task.CurrentChunk,
		&draftJSON,
		&task.YAML,
		&task.ErrorMessage,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return domain.ConversionTask{}, err
	}

	task.Status = domain.ConversionTaskStatus(status)
	if err := json.Unmarshal([]byte(chaptersJSON), &task.Chapters); err != nil {
		return domain.ConversionTask{}, err
	}
	if err := json.Unmarshal([]byte(aiConfigJSON), &task.AIConfig); err != nil {
		return domain.ConversionTask{}, err
	}
	if draftJSON != "null" {
		var draft domain.ScreenplayDraft
		if err := json.Unmarshal([]byte(draftJSON), &draft); err != nil {
			return domain.ConversionTask{}, err
		}
		task.Draft = &draft
	}

	return task, nil
}

func nullableJSONString(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}
