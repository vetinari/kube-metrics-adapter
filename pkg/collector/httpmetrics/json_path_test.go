package httpmetrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCastSlice(t *testing.T) {
	res1, err1 := castSlice([]interface{}{1, 2, 3})
	require.NoError(t, err1)
	require.Equal(t, []float64{1.0, 2.0, 3.0}, res1)

	res2, err2 := castSlice([]interface{}{float32(1.0), float32(2.0), float32(3.0)})
	require.NoError(t, err2)
	require.Equal(t, []float64{1.0, 2.0, 3.0}, res2)

	res3, err3 := castSlice([]interface{}{float64(1.0), float64(2.0), float64(3.0)})
	require.NoError(t, err3)
	require.Equal(t, []float64{1.0, 2.0, 3.0}, res3)

	res4, err4 := castSlice([]interface{}{1, 2, "some string"})
	require.Errorf(t, err4, "slice was returned by JSONPath, but value inside is unsupported: %T", "string")
	require.Equal(t, []float64(nil), res4)
}
