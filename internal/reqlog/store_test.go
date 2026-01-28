package reqlog

import "testing"

func TestStoreEvictsOldest(t *testing.T) {
	store := NewStore(3)

	store.Add(Entry{})
	store.Add(Entry{})
	store.Add(Entry{})
	store.Add(Entry{})

	got := store.List()
	if len(got) != 3 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].ID != 4 || got[1].ID != 3 || got[2].ID != 2 {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestStoreFindByID(t *testing.T) {
	store := NewStore(2)

	store.Add(Entry{Method: "GET"})
	store.Add(Entry{Method: "POST"})

	entry, ok := store.Find(2)
	if !ok {
		t.Fatal("expected id=2")
	}
	if entry.Method != "POST" {
		t.Fatalf("method = %q", entry.Method)
	}
	_, ok = store.Find(1)
	if !ok {
		t.Fatal("expected id=1")
	}
}
