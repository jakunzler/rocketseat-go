package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

func TestConcurrentUserCreates(t *testing.T) {
	mem := New()
	users, _, _ := Repos(mem)

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := uuid.Must(uuid.NewV7())
			now := time.Now().UTC()
			_, err := users.Create(context.Background(), store.User{
				ID:           id,
				Name:         "Ada",
				Email:        fmt.Sprintf("user%d@example.com", i),
				PasswordHash: "hash",
				CreatedAt:    now,
				UpdatedAt:    now,
			})
			if err != nil {
				t.Errorf("create user %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()
}
