package movidesk

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

func TestPaginate_StopsOnShortPage(t *testing.T) {
	pages := [][]int{
		{1, 2, 3},
		{4, 5}, // < pageSize → stop
	}
	calls := 0
	fetch := func(ctx context.Context, q odata.Query) ([]byte, error) {
		page := pages[calls]
		calls++
		return json.Marshal(page)
	}
	got, err := Paginate(context.Background(), odata.Query{}, fetch, 3, 0)
	require.NoError(t, err)
	assert.Len(t, got, 5)
	assert.Equal(t, 2, calls)
}

func TestPaginate_StopsOnEmptyPage(t *testing.T) {
	pages := [][]int{{1, 2}, {}}
	calls := 0
	fetch := func(ctx context.Context, q odata.Query) ([]byte, error) {
		page := pages[calls]
		calls++
		return json.Marshal(page)
	}
	got, err := Paginate(context.Background(), odata.Query{}, fetch, 2, 0)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, 2, calls)
}

func TestPaginate_RespectsMax(t *testing.T) {
	calls := 0
	fetch := func(ctx context.Context, q odata.Query) ([]byte, error) {
		calls++
		page := []int{calls*10 + 1, calls*10 + 2, calls*10 + 3}
		return json.Marshal(page)
	}
	got, err := Paginate(context.Background(), odata.Query{}, fetch, 3, 5)
	require.NoError(t, err)
	assert.Len(t, got, 5)
}

func TestPaginate_PropagatesError(t *testing.T) {
	fetch := func(ctx context.Context, q odata.Query) ([]byte, error) {
		return nil, fmt.Errorf("boom")
	}
	_, err := Paginate(context.Background(), odata.Query{}, fetch, 10, 0)
	require.Error(t, err)
}
