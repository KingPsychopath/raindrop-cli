package raindrop

import "testing"

func TestCollectionPathsAndCycles(t *testing.T) {
	rows, err := BuildCollectionTree([]Collection{{ID: 1, Title: "Work"}, {ID: 2, Title: "Docs", Parent: &Ref{ID: 1}}, {ID: 3, Title: "Orphan", Parent: &Ref{ID: 99}}})
	if err != nil {
		t.Fatal(err)
	}
	paths := map[int64]string{}
	for _, row := range rows {
		paths[row.ID] = row.Path
	}
	if paths[2] != "Work / Docs" || paths[3] != "[missing parent 99] / Orphan" {
		t.Fatalf("paths=%v", paths)
	}
	if _, err := BuildCollectionTree([]Collection{{ID: 1, Parent: &Ref{ID: 2}}, {ID: 2, Parent: &Ref{ID: 1}}}); err == nil {
		t.Fatal("cycle ignored")
	}
}
