package partition

import (
	"testing"
)

func TestHashPartition(t *testing.T) {
	n := 8

	t.Run("deterministic hash", func(t *testing.T) {
		orderID := "ORD-123"
		p1 := HashPartition(orderID, n)
		p2 := HashPartition(orderID, n)

		if p1 != p2 {
			t.Errorf("Expected deterministic hash, got %d and %d", p1, p2)
		}
	})

	t.Run("distribution", func(t *testing.T) {
		// Just a basic check that not everything goes to 0
		p1 := HashPartition("ORD-1", n)
		p2 := HashPartition("ORD-2", n)
		p3 := HashPartition("ORD-3", n)
		p4 := HashPartition("ORD-4", n)

		sum := p1 + p2 + p3 + p4
		if sum == 0 {
			t.Errorf("Expected some distribution, but all hashed to 0")
		}
	})

	t.Run("boundaries", func(t *testing.T) {
		p := HashPartition("A", n)
		if p < 0 || p >= n {
			t.Errorf("Expected partition between 0 and %d, got %d", n-1, p)
		}
	})
}

func TestUnsortedPartition(t *testing.T) {
	if p := UnsortedPartition(); p != -1 {
		t.Errorf("Expected UnsortedPartition to return -1, got %d", p)
	}
}
