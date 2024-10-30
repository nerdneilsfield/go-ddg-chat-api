package example

import (
	"fmt"
	"reflect"
)

type Matrix struct {
	Rows int
	Cols int
	Data [][]int
}

func NewMatrix(rows int, cols int, data [][]int) *Matrix {
	return &Matrix{
		Rows: rows,
		Cols: cols,
		Data: data,
	}
}

func (m *Matrix) Equals(n *Matrix) bool {
	return m.Rows == n.Rows && m.Cols == n.Cols && reflect.DeepEqual(m.Data, n.Data)
}

// Add adds two matrices.
// Returns a new matrix with the result of the addition.
// Returns an error if the matrices have different dimensions.
func (m *Matrix) Add(n *Matrix) (*Matrix, error) {
	if m.Rows != n.Rows || m.Cols != n.Cols {
		return nil, fmt.Errorf("matrix dimensions do not match")
	}

	result := &Matrix{
		Rows: m.Rows,
		Cols: m.Cols,
		Data: make([][]int, m.Rows),
	}

	for i := 0; i < m.Rows; i++ {
		result.Data[i] = make([]int, m.Cols)
		for j := 0; j < m.Cols; j++ {
			result.Data[i][j] = m.Data[i][j] + n.Data[i][j]
		}
	}

	return result, nil
}

// Sub subtracts two matrices.
// Returns a new matrix with the result of the subtraction.
// Returns an error if the matrices have different dimensions.
func (m *Matrix) Sub(n *Matrix) (*Matrix, error) {
	if m.Rows != n.Rows || m.Cols != n.Cols {
		return nil, fmt.Errorf("matrix dimensions do not match")
	}

	result := &Matrix{
		Rows: m.Rows,
		Cols: m.Cols,
		Data: make([][]int, m.Rows),
	}

	for i := 0; i < m.Rows; i++ {
		result.Data[i] = make([]int, m.Cols)
		for j := 0; j < m.Cols; j++ {
			result.Data[i][j] = m.Data[i][j] - n.Data[i][j]
		}
	}

	return result, nil
}

// Mul multiplies two matrices.
// Returns a new matrix with the result of the multiplication.
// Returns an error if the matrices have different dimensions.
func (m *Matrix) Mul(n *Matrix) (*Matrix, error) {
	if m.Cols != n.Rows {
		return nil, fmt.Errorf("matrix dimensions do not match")
	}

	result := &Matrix{
		Rows: m.Rows,
		Cols: n.Cols,
		Data: make([][]int, m.Rows),
	}

	for i := 0; i < m.Rows; i++ {
		result.Data[i] = make([]int, n.Cols)
		for j := 0; j < n.Cols; j++ {
			for k := 0; k < m.Cols; k++ {
				result.Data[i][j] += m.Data[i][k] * n.Data[k][j]
			}
		}
	}

	return result, nil
}

// Print prints the matrix.
func (m *Matrix) Print() {
	for _, row := range m.Data {
		for _, col := range row {
			fmt.Printf("%d ", col)
		}
		fmt.Println()
	}
}
