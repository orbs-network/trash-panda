package proxy

import (
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/trash-panda/config"
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
	}, config.GetLogger())

	require.Len(t, results, 4)

	result := filterResponsesByBlockHeight(results)
	require.EqualValues(t, 4, result.requestResult.BlockHeight())
}
