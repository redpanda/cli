package network

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/docker/cli/internal/test"
	// Import builders to get the builder function as package function
	. "github.com/docker/cli/internal/test/builders"
	"github.com/docker/cli/internal/test/testutil"
	"github.com/docker/docker/api/types"
	"github.com/gotestyourself/gotestyourself/golden"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNetworkInspectErrors(t *testing.T) {
	testCases := []struct {
		args                      []string
		flags                     map[string]string
		networkInspectWithRawFunc func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error)
		expectedError             string
	}{
		{
			expectedError: "requires at least 1 argument",
		},
		{
			args: []string{"foo"},
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				return types.NetworkResource{}, nil, errors.New("error while inspecting the network")
			},
			expectedError: "error while inspecting the network",
		},
		{
			args: []string{"foo"},
			flags: map[string]string{
				"format": "{{invalid format}}",
			},
			expectedError: "Template parsing error",
		},
		{
			args: []string{"foo", "bar"},
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				if networkID == "foo" {
					return *Network(NetworkName(networkID)), []byte(""), nil
				}
				return types.NetworkResource{}, nil, errors.Errorf("error while inspecting the network")
			},
			expectedError: "error while inspecting the network",
		},
	}

	for _, tc := range testCases {
		cmd := newInspectCommand(
			test.NewFakeCli(&fakeClient{
				networkInspectWithRawFunc: tc.networkInspectWithRawFunc,
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

func TestNetworkInspectWithoutFormat(t *testing.T) {
	testCases := []struct {
		name                      string
		args                      []string
		networkInspectWithRawFunc func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error)
	}{
		{
			name: "single-network",
			args: []string{"foo"},
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				if networkID != "foo" {
					return *Network(), nil, errors.Errorf("Invalid name, expected %s, got %s", "foo", networkID)
				}
				return *Network(NetworkID("ID-"+networkID), NetworkName(networkID)), nil, nil
			},
		},
		{
			name: "multiple-networks-with-labels",
			args: []string{"foo", "bar"},
			networkInspectWithRawFunc: func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
				return *Network(NetworkID("ID-"+networkID), NetworkName(networkID), NetworkLabels(map[string]string{
					"foo": "bar",
				})), nil, nil
			},
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			networkInspectWithRawFunc: tc.networkInspectWithRawFunc,
		})
		cmd := newInspectCommand(cli)
		cmd.SetArgs(tc.args)
		assert.NoError(t, cmd.Execute())
		golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("network-inspect-without-format.%s.golden", tc.name))
	}
}

func TestNetworkInspectWithFormat(t *testing.T) {
	networkInspectWithRawFunc := func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error) {
		return *Network(NetworkName("foo"), NetworkLabels(map[string]string{
			"foo": "bar",
		})), nil, nil
	}
	testCases := []struct {
		name                      string
		format                    string
		args                      []string
		networkInspectWithRawFunc func(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, []byte, error)
	}{
		{
			name:   "simple-template",
			format: "{{.Name}}",
			args:   []string{"foo"},
			networkInspectWithRawFunc: networkInspectWithRawFunc,
		},
		{
			name:   "json-template",
			format: "{{json .Labels}}",
			args:   []string{"foo"},
			networkInspectWithRawFunc: networkInspectWithRawFunc,
		},
	}
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{
			networkInspectWithRawFunc: tc.networkInspectWithRawFunc,
		})
		cmd := newInspectCommand(cli)
		cmd.SetArgs(tc.args)
		cmd.Flags().Set("format", tc.format)
		assert.NoError(t, cmd.Execute())
		golden.Assert(t, cli.OutBuffer().String(), fmt.Sprintf("network-inspect-with-format.%s.golden", tc.name))
	}
}
