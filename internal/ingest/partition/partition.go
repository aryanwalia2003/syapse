package partition

import (
	"hash/fnv"
)

// HashPartition computes a deterministic partition index between 0 and n-1
// using the FNV-1a non-cryptographic hash algorithm.
func HashPartition(vendorOrderID string, n int) int {
	if n <= 0 {
		return UnsortedPartition()
	}
	h := fnv.New32a()
	h.Write([]byte(vendorOrderID))
	// Cast the unsigned uint32 to int and ensure it's positive before modulo
	hashInt := int(h.Sum32() & 0x7fffffff)
	return hashInt % n
}

// UnsortedPartition returns a sentinel value for the unsorted lane
func UnsortedPartition() int {
	return -1
}
