package proxy

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_aggregateAndFilter(t *testing.T) {
	endpoints := []string{"a", "b", "c", "d"}

	results := aggregateRequest(4, endpoints, func(endpoint string) response {
		switch endpoint {
		case "b":
			return response{
				blockHeight: 3,
			}
		case "c":
			return response{
				blockHeight: 4,
			}
		case "d":
			return response{
				blockHeight: 4,
			}
		default:
			return response{
				blockHeight: 0,
				httpErr: &HttpErr{
					Message: "timeout",
				},
			}
		}
	})

	require.Len(t, results, 4)

	require.EqualValues(t, 0, results[0].blockHeight)
	require.EqualValues(t, 3, results[1].blockHeight)
	require.EqualValues(t, 4, results[2].blockHeight)
	require.EqualValues(t, 4, results[3].blockHeight)

	result := filterResponses(results)
	require.EqualValues(t, 4, result.blockHeight)
}
