package proxy

import (
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_aggregateAndFilter(t *testing.T) {
	endpoints := []string{"a", "b", "c", "d"}

	results := aggregateRequest(4, endpoints, func(endpoint string) response {
		switch endpoint {
		case "b":
			return response{
				requestResult: (&client.RequestResultBuilder{BlockHeight: 3}).Build(),
			}
		case "c":
			return response{
				requestResult: (&client.RequestResultBuilder{BlockHeight: 4}).Build(),
			}
		case "d":
			return response{
				requestResult: (&client.RequestResultBuilder{BlockHeight: 4}).Build(),
			}
		default:
			return response{
				httpErr: &HttpErr{
					Message: "timeout",
				},
			}
		}
	})

	require.Len(t, results, 4)

	require.Nil(t, results[0].requestResult)
	require.EqualValues(t, 3, results[1].requestResult.BlockHeight())
	require.EqualValues(t, 4, results[2].requestResult.BlockHeight())
	require.EqualValues(t, 4, results[3].requestResult.BlockHeight())

	result := filterResponsesByBlockHeight(results)
	require.EqualValues(t, 4, result.requestResult.BlockHeight())
}
