package bus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/oullin/alloy/packages/foundation/bus"
)

// mockDynamoClient implements bus.DynamoClient for testing.
type mockDynamoClient struct {
	mu             sync.Mutex
	items          map[string]map[string]any // keyed by "app:id"
	lastUpdateExpr string
	putErr         error
	getErr         error
	updErr         error
	delErr         error
}

func newMockDynamoClient() *mockDynamoClient {
	return &mockDynamoClient{items: make(map[string]map[string]any)}
}

func (c *mockDynamoClient) itemKey(key map[string]any) string {
	return fmt.Sprintf("%v:%v", key["application"], key["id"])
}

func (c *mockDynamoClient) PutItem(_ context.Context, _ string, item map[string]any) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	if c.putErr != nil {
		return c.putErr
	}

	k := fmt.Sprintf("%v:%v", item["application"], item["id"])
	c.items[k] = item

	return nil
}

func (c *mockDynamoClient) GetItem(_ context.Context, _ string, key map[string]any) (map[string]any, error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	if c.getErr != nil {
		return nil, c.getErr
	}

	return c.items[c.itemKey(key)], nil
}

func (c *mockDynamoClient) UpdateItem(_ context.Context, _ string, key map[string]any, update string, values map[string]any) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	if c.updErr != nil {
		return c.updErr
	}

	c.lastUpdateExpr = update

	k := c.itemKey(key)
	item, ok := c.items[k]

	if !ok {
		return nil
	}

	if update == "SET total_jobs = total_jobs + :amount, pending_jobs = pending_jobs + :amount" {
		amount, _ := values[":amount"].(int)
		item["total_jobs"] = numberValue(item["total_jobs"]) + amount
		item["pending_jobs"] = numberValue(item["pending_jobs"]) + amount

		return nil
	}

	if update == "SET pending_jobs = pending_jobs - :one" {
		amount, _ := values[":one"].(int)
		item["pending_jobs"] = numberValue(item["pending_jobs"]) - amount

		return nil
	}

	if update == "SET failed_jobs = failed_jobs + :one, pending_jobs = pending_jobs - :one, failed_job_ids = list_append(if_not_exists(failed_job_ids, :empty_list), :new_ids)" {
		amount, _ := values[":one"].(int)
		item["failed_jobs"] = numberValue(item["failed_jobs"]) + amount
		item["pending_jobs"] = numberValue(item["pending_jobs"]) - amount
		item["failed_job_ids"] = append(stringListValue(item["failed_job_ids"]), values[":new_ids"].([]string)...)

		return nil
	}

	// Apply simple value updates for testing.
	for vk, vv := range values {
		item[vk] = vv
	}

	return nil
}

func numberValue(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func stringListValue(value any) []string {
	switch v := value.(type) {
	case []string:
		return append([]string(nil), v...)
	case []any:
		items := make([]string, 0, len(v))

		for _, item := range v {
			if s, ok := item.(string); ok {
				items = append(items, s)
			}
		}

		return items
	default:
		return nil
	}
}

func (c *mockDynamoClient) DeleteItem(_ context.Context, _ string, key map[string]any) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	if c.delErr != nil {
		return c.delErr
	}

	delete(c.items, c.itemKey(key))

	return nil
}

func (c *mockDynamoClient) Query(_ context.Context, _ string, _ string, values map[string]any, limit int) ([]map[string]any, error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	var results []map[string]any

	app, _ := values[":app"].(string)

	for key, item := range c.items {
		if len(results) >= limit {
			break
		}

		itemApp, _ := item["application"].(string)

		if itemApp == app || key != "" {
			results = append(results, item)
		}
	}

	return results, nil
}

func TestDynamoRepoStore(t *testing.T) {
	client := newMockDynamoClient()
	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	batch := &bus.Batch{
		ID:           "batch-1",
		Name:         "dynamo-test",
		TotalJobs:    5,
		PendingJobs:  5,
		FailedJobIDs: []string{"j1"},
		Options:      map[string]any{"allowFailures": true},
		CreatedAt:    time.Now(),
	}

	err := repo.Store(context.Background(), batch)

	if err != nil {
		t.Fatal(err)
	}

	client.mu.Lock()

	defer client.mu.Unlock()

	item, ok := client.items["myapp:batch-1"]

	if !ok {
		t.Fatal("expected item to be stored")
	}

	if item["name"] != "dynamo-test" {
		t.Errorf("expected name 'dynamo-test', got %v", item["name"])
	}

	failedIDs, ok := item["failed_job_ids"].([]string)

	if !ok {
		t.Fatalf("expected failed_job_ids to be stored as []string, got %T", item["failed_job_ids"])
	}

	if len(failedIDs) != 1 || failedIDs[0] != "j1" {
		t.Fatalf("expected failed_job_ids [j1], got %v", failedIDs)
	}
}

func TestDynamoRepoGet(t *testing.T) {
	client := newMockDynamoClient()
	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	failedJSON, _ := json.Marshal([]string{})
	optsJSON, _ := json.Marshal(map[string]any{})

	client.items["myapp:batch-1"] = map[string]any{
		"id":             "batch-1",
		"name":           "test",
		"total_jobs":     float64(10),
		"pending_jobs":   float64(3),
		"failed_jobs":    float64(1),
		"failed_job_ids": string(failedJSON),
		"options":        string(optsJSON),
		"created_at":     float64(time.Now().Unix()),
	}

	batch, err := repo.Get(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}

	if batch == nil {
		t.Fatal("expected non-nil batch")
	}

	if batch.ID != "batch-1" {
		t.Errorf("expected ID 'batch-1', got %q", batch.ID)
	}

	if batch.TotalJobs != 10 {
		t.Errorf("expected TotalJobs 10, got %d", batch.TotalJobs)
	}
}

func TestDynamoRepoGetReadsFailedJobIDsList(t *testing.T) {
	client := newMockDynamoClient()
	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	client.items["myapp:batch-1"] = map[string]any{
		"id":             "batch-1",
		"name":           "test",
		"total_jobs":     float64(10),
		"pending_jobs":   float64(3),
		"failed_jobs":    float64(1),
		"failed_job_ids": []any{"failed-1", "failed-2"},
		"options":        "{}",
		"created_at":     float64(time.Now().Unix()),
	}

	batch, err := repo.Get(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}

	if len(batch.FailedJobIDs) != 2 || batch.FailedJobIDs[0] != "failed-1" || batch.FailedJobIDs[1] != "failed-2" {
		t.Fatalf("expected failed job IDs from list, got %v", batch.FailedJobIDs)
	}
}

func TestDynamoRepoGetNotFound(t *testing.T) {
	client := newMockDynamoClient()
	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	batch, err := repo.Get(context.Background(), "nonexistent")

	if err != nil {
		t.Fatal(err)
	}

	if batch != nil {
		t.Error("expected nil batch for nonexistent ID")
	}
}

func TestDynamoRepoMarkAsFinished(t *testing.T) {
	client := newMockDynamoClient()
	client.items["myapp:batch-1"] = map[string]any{
		"application": "myapp",
		"id":          "batch-1",
	}

	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	err := repo.MarkAsFinished(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}
}

func TestDynamoRepoCancel(t *testing.T) {
	client := newMockDynamoClient()
	client.items["myapp:batch-1"] = map[string]any{
		"application": "myapp",
		"id":          "batch-1",
	}

	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	err := repo.Cancel(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}
}

func TestDynamoRepoDelete(t *testing.T) {
	client := newMockDynamoClient()
	client.items["myapp:batch-1"] = map[string]any{
		"application": "myapp",
		"id":          "batch-1",
	}

	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	err := repo.Delete(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}

	client.mu.Lock()
	_, exists := client.items["myapp:batch-1"]
	client.mu.Unlock()

	if exists {
		t.Error("expected item to be deleted")
	}
}

func TestDynamoRepoStoreWithTTL(t *testing.T) {
	ttl := 86400
	client := newMockDynamoClient()
	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", &ttl, "expires_at")

	batch := &bus.Batch{
		ID:        "batch-ttl",
		Options:   map[string]any{},
		CreatedAt: time.Now(),
	}

	err := repo.Store(context.Background(), batch)

	if err != nil {
		t.Fatal(err)
	}

	client.mu.Lock()
	item := client.items["myapp:batch-ttl"]
	client.mu.Unlock()

	if item == nil {
		t.Fatal("expected item to be stored")
	}

	if _, ok := item["expires_at"]; !ok {
		t.Error("expected expires_at field to be set with TTL")
	}
}

func TestDynamoRepoTransaction(t *testing.T) {
	client := newMockDynamoClient()
	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	called := false
	err := repo.Transaction(context.Background(), func(r bus.BatchRepository) error {
		called = true

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Error("expected fn to be called")
	}
}

func TestDynamoRepoStoreError(t *testing.T) {
	client := newMockDynamoClient()
	client.putErr = fmt.Errorf("dynamo error")

	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	batch := &bus.Batch{ID: "err", Options: map[string]any{}, CreatedAt: time.Now()}
	err := repo.Store(context.Background(), batch)

	if err == nil {
		t.Error("expected error")
	}
}

func TestDynamoRepoIncrementTotalJobs(t *testing.T) {
	client := newMockDynamoClient()
	client.items["myapp:batch-1"] = map[string]any{
		"application": "myapp",
		"id":          "batch-1",
	}

	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	err := repo.IncrementTotalJobs(context.Background(), "batch-1", 5)

	if err != nil {
		t.Fatal(err)
	}
}

func TestDynamoRepoIncrementFailedJobsAppendsFailedJobIDAtomically(t *testing.T) {
	client := newMockDynamoClient()
	client.items["myapp:batch-1"] = map[string]any{
		"application":    "myapp",
		"id":             "batch-1",
		"total_jobs":     float64(3),
		"pending_jobs":   float64(2),
		"failed_jobs":    float64(1),
		"failed_job_ids": []string{"failed-1"},
	}

	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	counts, err := repo.IncrementFailedJobs(context.Background(), "batch-1", "failed-2")

	if err != nil {
		t.Fatal(err)
	}

	if counts.PendingJobs != 1 || counts.FailedJobs != 2 {
		t.Fatalf("expected counts pending=1 failed=2, got pending=%d failed=%d", counts.PendingJobs, counts.FailedJobs)
	}

	client.mu.Lock()

	defer client.mu.Unlock()

	if client.lastUpdateExpr != "SET failed_jobs = failed_jobs + :one, pending_jobs = pending_jobs - :one, failed_job_ids = list_append(if_not_exists(failed_job_ids, :empty_list), :new_ids)" {
		t.Fatalf("expected atomic list_append update, got %q", client.lastUpdateExpr)
	}

	failedIDs := client.items["myapp:batch-1"]["failed_job_ids"].([]string)

	if len(failedIDs) != 2 || failedIDs[0] != "failed-1" || failedIDs[1] != "failed-2" {
		t.Fatalf("expected appended failed job IDs, got %v", failedIDs)
	}
}

func TestDynamoRepoGetList(t *testing.T) {
	client := newMockDynamoClient()
	client.items["myapp:batch-1"] = map[string]any{
		"application":    "myapp",
		"id":             "batch-1",
		"name":           "test-1",
		"total_jobs":     float64(5),
		"pending_jobs":   float64(3),
		"failed_jobs":    float64(0),
		"failed_job_ids": "[]",
		"options":        "{}",
		"created_at":     float64(time.Now().Unix()),
	}

	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	batches, err := repo.GetList(context.Background(), 10, "")

	if err != nil {
		t.Fatal(err)
	}

	if len(batches) != 1 {
		t.Errorf("expected 1 batch, got %d", len(batches))
	}
}

func TestDynamoRepoRollBack(t *testing.T) {
	client := newMockDynamoClient()
	repo := bus.NewDynamoBatchRepository(client, "myapp", "batches", nil, "")

	err := repo.RollBack(context.Background())

	if err != nil {
		t.Errorf("expected no error from RollBack, got %v", err)
	}
}

// Ensure DynamoBatchRepository satisfies BatchRepository.
var _ bus.BatchRepository = (*bus.DynamoBatchRepository)(nil)
