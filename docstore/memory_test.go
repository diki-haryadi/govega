package docstore

import "testing"

func TestMemoryStore(t *testing.T) {
	ms := NewMemoryStore("test", "id")
	DriverCRUDTest(ms, t)
	DriverBulkTest(ms, t)
}
