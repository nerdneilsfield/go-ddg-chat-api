package example

import "testing"

func TestMatrix_Add(t *testing.T) {
	m := NewMatrix(2, 2, [][]int{{1, 2}, {3, 4}})

	n := NewMatrix(2, 2, [][]int{{5, 6}, {7, 8}})

	result, err := m.Add(n)
	if err != nil {
		t.Fatalf("failed to add matrices: %v", err)
	}

	gold := NewMatrix(2, 2, [][]int{{6, 8}, {10, 12}})

	if !result.Equals(gold) {
		t.Fatalf("result does not match gold")
	}
}

func TestMatrix_Sub(t *testing.T) {
	m := NewMatrix(2, 2, [][]int{{1, 2}, {3, 4}})

	n := NewMatrix(2, 2, [][]int{{5, 6}, {7, 8}})

	result, err := m.Sub(n)
	if err != nil {
		t.Fatalf("failed to subtract matrices: %v", err)
	}

	gold := NewMatrix(2, 2, [][]int{{-4, -4}, {-4, -4}})

	if !result.Equals(gold) {
		t.Fatalf("result does not match gold")
	}
}

func TestMatrix_Mul(t *testing.T) {
	m := NewMatrix(2, 2, [][]int{{1, 2}, {3, 4}})

	n := NewMatrix(2, 2, [][]int{{5, 6}, {7, 8}})

	result, err := m.Mul(n)
	if err != nil {
		t.Fatalf("failed to multiply matrices: %v", err)
	}

	gold := NewMatrix(2, 2, [][]int{{19, 22}, {43, 50}})

	if !result.Equals(gold) {
		t.Fatalf("result does not match gold")
	}
}
