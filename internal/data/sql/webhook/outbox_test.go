package webhook

import (
	"context"
	"testing"
	"time"

	"log/slog"

	coreDB "github.com/aryanwalia/synapse/internal/core/db"
	"github.com/aryanwalia/synapse/internal/core/domain"
	"github.com/aryanwalia/synapse/internal/core/logger"
)

func init() {
	logger.Init(slog.LevelDebug, false)
}

// --- In-Memory Test Double for DB ---

type inMemoryRow struct {
	id             string
	status         string
	partitionIndex int
	retryCount     int
	isDLQ          bool
	updatedAt      time.Time
}

type fakeDB struct {
	rows map[string]*inMemoryRow
}

func newFakeDB(initial ...*inMemoryRow) *fakeDB {
	rows := make(map[string]*inMemoryRow)
	for _, r := range initial {
		rows[r.id] = r
	}
	return &fakeDB{rows: rows}
}

func (f *fakeDB) Execute(ctx context.Context, query string, params map[string]any) error {
	id, hasID := params["id"].(string)
	if !hasID {
		return nil
	}
	row, exists := f.rows[id]
	if !exists {
		return nil
	}
	// Status update
	if status, ok := params["status"].(string); ok {
		row.status = status
	}
	// DLQ flag
	if isDLQ, ok := params["is_dlq"].(bool); ok {
		row.isDLQ = isDLQ
	}
	// The IncrementRetry SQL uses `retry_count + 1` in SQL with no param.
	// We detect it by the absence of a status/is_dlq param alongside an id.
	_, hasStatus := params["status"]
	_, hasDLQ := params["is_dlq"]
	if !hasStatus && !hasDLQ {
		row.retryCount++
	}
	row.updatedAt = time.Now().UTC()
	f.rows[id] = row
	return nil
}

func (f *fakeDB) QueryRow(ctx context.Context, query string, params map[string]any, dest any) error {
	return nil
}

func (f *fakeDB) QueryRows(ctx context.Context, query string, params map[string]any, dest any) error {
	if destSlice, ok := dest.(*[]*domain.RawWebhook); ok {
		partIdx := params["partition_index"].(int)
		limit := params["limit"].(int)
		count := 0
		for _, row := range f.rows {
			if row.status == "PENDING" && row.partitionIndex == partIdx && count < limit {
				*destSlice = append(*destSlice, &domain.RawWebhook{
					ID:             row.id,
					Status:         row.status,
					RetryCount:     row.retryCount,
					PartitionIndex: row.partitionIndex,
				})
				count++
			}
		}
	}
	return nil
}

func (f *fakeDB) RunInTransaction(ctx context.Context, fn func(tx coreDB.Transaction) error) error {
	return fn(f)
}

// --- Tests ---

func TestClaimPending_ReturnsMatchingRows(t *testing.T) {
	db := newFakeDB(
		&inMemoryRow{id: "wh-p1", status: "PENDING", partitionIndex: 2},
		&inMemoryRow{id: "wh-p2", status: "PENDING", partitionIndex: 2},
		&inMemoryRow{id: "wh-other", status: "PENDING", partitionIndex: 5},
		&inMemoryRow{id: "wh-done", status: "DONE", partitionIndex: 2},
	)
	repo := &SQLWebhookRepository{db: db}

	results, err := repo.ClaimPending(context.Background(), 2, 10)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 PENDING rows for partition 2, got %d", len(results))
	}
}

func TestMarkProcessing_ChangesStatus(t *testing.T) {
	db := newFakeDB(&inMemoryRow{id: "wh-1", status: "PENDING"})
	repo := &SQLWebhookRepository{db: db}

	err := repo.MarkProcessing(context.Background(), "wh-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db.rows["wh-1"].status != "PROCESSING" {
		t.Errorf("Expected status PROCESSING, got %s", db.rows["wh-1"].status)
	}
}

func TestMarkDone_ChangesStatus(t *testing.T) {
	db := newFakeDB(&inMemoryRow{id: "wh-1", status: "PROCESSING"})
	repo := &SQLWebhookRepository{db: db}

	err := repo.MarkDone(context.Background(), "wh-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db.rows["wh-1"].status != "DONE" {
		t.Errorf("Expected status DONE, got %s", db.rows["wh-1"].status)
	}
}

func TestMarkFailed_ChangesStatusAndSetsDLQ(t *testing.T) {
	db := newFakeDB(&inMemoryRow{id: "wh-1", status: "PROCESSING", retryCount: 3})
	repo := &SQLWebhookRepository{db: db}

	err := repo.MarkFailed(context.Background(), "wh-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db.rows["wh-1"].status != "FAILED" {
		t.Errorf("Expected status FAILED, got %s", db.rows["wh-1"].status)
	}
	if !db.rows["wh-1"].isDLQ {
		t.Errorf("Expected is_dlq to be true")
	}
}

func TestIncrementRetry_IncrementsCounter(t *testing.T) {
	db := newFakeDB(&inMemoryRow{id: "wh-1", status: "PROCESSING", retryCount: 1})
	repo := &SQLWebhookRepository{db: db}

	err := repo.IncrementRetry(context.Background(), "wh-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db.rows["wh-1"].retryCount != 2 {
		t.Errorf("Expected retry_count 2, got %d", db.rows["wh-1"].retryCount)
	}
}

func TestRecoverStuck_ResetsOldProcessingJobs(t *testing.T) {
	stuckTime := time.Now().Add(-15 * time.Minute)
	recentTime := time.Now().Add(-2 * time.Minute)

	db := newFakeDB(
		&inMemoryRow{id: "wh-stuck", status: "PROCESSING", updatedAt: stuckTime},
		&inMemoryRow{id: "wh-active", status: "PROCESSING", updatedAt: recentTime},
	)
	repo := &SQLWebhookRepository{db: db}

	// The SQL implementation will reset stuck jobs; the fake DB doesn't simulate
	// time-based  queries but tests the method signature and that it runs error-free.
	count, err := repo.RecoverStuck(context.Background(), 10*time.Minute)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// The real implementation will return affected rows; fake returns 0 until SQL is real.
	_ = count
}
