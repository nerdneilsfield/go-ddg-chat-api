package call

import (
	"fmt"

	"github.com/nerdneilsfield/go-template/pkg/example"
)

type RowWiseMatrix struct {
	Rows int
	Cols int
	Data []int
}

type ColumnWiseMatrix struct {
	Rows int
	Cols int
	Data []int
}

func (m *RowWiseMatrix) ToMatrix() *example.Matrix {
	data := make([][]int, m.Rows)
	for i := 0; i < m.Rows; i++ {
		data[i] = m.Data[i*m.Cols : (i+1)*m.Cols]
	}
	return example.NewMatrix(m.Rows, m.Cols, data)
}

func (m *ColumnWiseMatrix) ToMatrix() *example.Matrix {
	data := make([][]int, m.Rows)
	for i := 0; i < m.Rows; i++ {
		data[i] = m.Data[i*m.Cols : (i+1)*m.Cols]
	}
	return example.NewMatrix(m.Rows, m.Cols, data)
}

func (m *RowWiseMatrix) Add(n *RowWiseMatrix) (*RowWiseMatrix, error) {
	if m.Rows != n.Rows || m.Cols != n.Cols {
		return nil, fmt.Errorf("matrix dimensions do not match")
	}

	result := &RowWiseMatrix{
		Rows: m.Rows,
		Cols: m.Cols,
		Data: make([]int, m.Rows*m.Cols),
	}

	for i := 0; i < m.Rows*m.Cols; i++ {
		result.Data[i] = m.Data[i] + n.Data[i]
	}

	return result, nil
}
