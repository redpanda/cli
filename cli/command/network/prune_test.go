package network

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/docker/cli/internal/test"
	"github.com/docker/cli/internal/test/testutil"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/gotestyourself/gotestyourself/golden"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNetworkPruneErrors(t *testing.T) {
	testCases := []struct {
		args             []string
		flags            map[string]string
		networkPruneFunc func(args filters.Args) (types.NetworksPruneReport, error)
		expectedError    string
	}{
		{
			args:          []string{"foo"},
			expectedError: "accepts no argument",
		},
		{
			flags: map[string]string{
				"force": "true",
			},
			networkPruneFunc: func(args filters.Args) (types.NetworksPruneReport, error) {
				return types.NetworksPruneReport{}, errors.Errorf("error pruning networks")
			},
			expectedError: "error pruning networks",
		},
	}
	for _, tc := range testCases {
		cmd := NewPruneCommand(
			test.NewFakeCli(&fakeClient{
				networkPruneFunc: tc.networkPruneFunc,
			}),
		)
		cmd.SetArgs(tc.args)
		for key, value := range tc.flags {
			cmd.Flags().Set(key, value)
		}
		cmd.SetOutput(ioutil.Discard)
		testutil.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNetworkPruneForce(t *testing.T) {
	testCases := []struct {
		name             string
		networkPruneFunc func(args filters.Args) (types.NetworksPruneReport, error)
	}{
		{
			name: "empty",
		},
		{
			name: "deleted-networks",
			networkPruneFunc: func(args filters.Args) (types.NetworksPruneReport, error) {
				return types.NetworksPruneReport{
					NetworksDeleted: []string{
						"foo", "bar", "baz",
					},
				}, nil
			},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			networkPruneFunc: tc.networkPruneFunc,
		})
		cmd := NewPruneCommand(cli)
		cmd.Flags().Set("force", "true")
		assert.NoError(t, cmd.Execute())
		golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("network-prune.%s.golden", tc.name))
	}
}
