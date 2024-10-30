package call

import "testing"

func TestRowWiseMatrix_Add(t *testing.T) {
	m := &RowWiseMatrix{
		Rows: 2,
		Cols: 2,
		Data: []int{1, 2, 3, 4},
	}

	n := &RowWiseMatrix{
		Rows: 2,
		Cols: 2,
		Data: []int{5, 6, 7, 8},
	}

	result, err := m.Add(n)
	if err != nil {
		t.Fatalf("failed to add matrices: %v", err)
	}

	gold := &RowWiseMatrix{
		Rows: 2,
		Cols: 2,
		Data: []int{6, 8, 10, 12},
	}

	result.ToMatrix().Equals(gold.ToMatrix())
}
