package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsDescription(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	md, err := st.GetMetricsDescription(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, md)
	assert.Len(t, md, 2)
}

func TestGetMetrics(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	_, err := st.GetMetrics(ctx, "", "")
	assert.Error(t, err)

	nodes := generateNodes(ctx, t, st, 1, []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}, 0)
	_, err = st.GetMetrics(ctx, "testpod", nodes[0])
	assert.NoError(t, err)
}
