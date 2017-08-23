package network

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/internal/test"
	// Import builders to get the builder function as package function
	. "github.com/docker/cli/internal/test/builders"
	"github.com/docker/cli/internal/test/testutil"
	"github.com/docker/docker/api/types"
	"github.com/gotestyourself/gotestyourself/golden"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNetworkRemoveErrors(t *testing.T) {
	testCases := []struct {
		args                      []string
		networkInspectWithRawFunc func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error)
		networkRemoveFunc         func(networkID string) error
		expectedError             string
	}{
		{
			expectedError: "requires at least 1 argument",
		},
		{
			args: []string{"foo"},
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				return types.NetworkResource{}, nil, errors.Errorf("error while inspecting the network")
			},
			expectedError: "error while inspecting the network",
		},
		{
			args: []string{"foo"},
			networkRemoveFunc: func(networkID string) error {
				return errors.Errorf("error while removing the network")
			},
			expectedError: "error while removing the network",
		},
	}

	for _, tc := range testCases {
		cmd := newRemoveCommand(test.NewFakeCli(
			&fakeClient{
				networkInspectWithRawFunc: tc.networkInspectWithRawFunc,
				networkRemoveFunc:         tc.networkRemoveFunc,
			},
		))
		cmd.SetArgs(tc.args)
		cmd.SetOutput(ioutil.Discard)
		testutil.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNetworkRemovePromptNo(t *testing.T) {
	for _, input := range []string{"n", "N", "no", "anything", "really"} {
		cli := test.NewFakeCli(&fakeClient{
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				return *Network(NetworkID("ID-foo"), NetworkName("foo"), Ingress()), nil, nil
			},
		})

		cli.SetIn(command.NewInStream(ioutil.NopCloser(strings.NewReader(input))))
		cmd := newRemoveCommand(cli)
		cmd.SetArgs([]string{"foo"})
		assert.NoError(t, cmd.Execute())
		golden.Assert(t, cli.OutBuffer().String(), "network-remove-no.golden")
	}
}

func TestNetworkRemovePromptYes(t *testing.T) {
	for _, input := range []string{"y", "Y"} {
		cli := test.NewFakeCli(&fakeClient{
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				return *Network(NetworkID("ID-foo"), NetworkName("foo"), Ingress()), nil, nil
			},
		})

		cli.SetIn(command.NewInStream(ioutil.NopCloser(strings.NewReader(input))))
		cmd := newRemoveCommand(cli)
		cmd.SetArgs([]string{"foo"})
		assert.NoError(t, cmd.Execute())
		golden.Assert(t, cli.OutBuffer().String(), "network-remove-yes.golden")
	}
}

func TestNetworkRemove(t *testing.T) {
	testCases := []struct {
		name                      string
		args                      []string
		networkInspectWithRawFunc func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error)
	}{
		{
			name: "simple",
			args: []string{"foo"},
		},
		{
			name: "multiple",
			args: []string{"foo", "bar"},
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				return *Network(NetworkID("ID-"+networkID), NetworkName(networkID)), nil, nil
			},
		},
	}

	for _, tc := range testCases {
		cli := test.NewFakeCli(
			&fakeClient{
				networkInspectWithRawFunc: tc.networkInspectWithRawFunc,
			},
		)
		cmd := newRemoveCommand(cli)
		cmd.SetArgs(tc.args)
		assert.NoError(t, cmd.Execute())
	}
}
