package ck

// API is the api that the ck provided
type API interface {
	// Remove remove the table partitions
	Remove(table string, partitions ...uint64) error

	// Create create the table partitions
	Create(table string, partitions ...uint64) error
}
