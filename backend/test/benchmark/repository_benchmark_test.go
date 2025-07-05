// backend/test/benchmark/repository_benchmark_test.go

package benchmark

import (
	"context"
	"testing"

	"oilgas-backend/internal/repository"
	"oilgas-backend/test/testutil"
)

func BenchmarkCustomerRepository_GetAll(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	db := testutil.SetupTestDB(b)
	defer testutil.CleanupTestDB(b, db)

	repos := repository.New(db)
	ctx := context.Background()

	// Setup test data
	for i := 0; i < 100; i++ {
		customer := &models.Customer{
			Name: fmt.Sprintf("Test Customer %d", i),
		}
		repos.Customer.Create(ctx, customer)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repos.Customer.GetAll(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReceivedRepository_GetFiltered(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	db := testutil.SetupTestDB(b)
	defer testutil.CleanupTestDB(b, db)

	repos := repository.New(db)
	ctx := context.Background()

	// Setup test data
	customer := &models.Customer{Name: "Benchmark Customer"}
	repos.Customer.Create(ctx, customer)

	for i := 0; i < 1000; i++ {
		received := &models.ReceivedItem{
			WorkOrder:  fmt.Sprintf("WO-BENCH-%d", i),
			CustomerID: customer.ID,
			Customer:   customer.Name,
			Joints:     100,
			Size:       "5 1/2\"",
			Grade:      "J55",
		}
		repos.Received.Create(ctx, received)
	}

	filters := map[string]interface{}{
		"customer_id": customer.ID,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repos.Received.GetFiltered(ctx, filters, 50, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
