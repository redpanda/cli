package network

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/docker/cli/cli/config/configfile"
	// Import builders to get the builder function as package function
	"github.com/docker/cli/internal/test"
	. "github.com/docker/cli/internal/test/builders"
	"github.com/docker/cli/internal/test/testutil"
	"github.com/docker/docker/api/types"
	"github.com/gotestyourself/gotestyourself/golden"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNetworkListErrors(t *testing.T) {
	testCases := []struct {
		args            []string
		networkListFunc func(options types.NetworkListOptions) ([]types.NetworkResource, error)
		expectedError   string
	}{
		{
			args:          []string{"foo"},
			expectedError: "accepts no argument",
		},
		{
			networkListFunc: func(options types.NetworkListOptions) ([]types.NetworkResource, error) {
				return []types.NetworkResource{}, errors.Errorf("error listing networks")
			},
			expectedError: "error listing networks",
		},
	}

	for _, tc := range testCases {
		cmd := newListCommand(test.NewFakeCli(
			&fakeClient{networkListFunc: tc.networkListFunc},
		))
		cmd.SetArgs(tc.args)
		cmd.SetOutput(ioutil.Discard)
		testutil.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestNetworkListWithoutFormat(t *testing.T) {
	buf := new(bytes.Buffer)
	cli := test.NewFakeCli(&fakeClient{
		networkListFunc: func(options types.NetworkListOptions) ([]types.NetworkResource, error) {
			return []types.NetworkResource{
				*Network(),
				*Network(NetworkID("ID-foo"), NetworkName("foo"), NetworkDriver("host")),
				*Network(NetworkID("ID-bar"), NetworkName("bar"), NetworkLabels(map[string]string{
					"foo": "bar",
				})),
			}, nil
		},
	})
	cmd := newListCommand(cli)
	cmd.SetOutput(buf)
	assert.NoError(t, cmd.Execute())
	golden.Assert(t, cli.OutBuffer().String(), "network-list.golden")
}

func TestNetworkListWithConfigFormat(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		networkListFunc: func(options types.NetworkListOptions) ([]types.NetworkResource, error) {
			return []types.NetworkResource{
				*Network(NetworkName("foo"), NetworkDriver("host")),
				*Network(NetworkName("bar"), NetworkLabels(map[string]string{
					"foo": "bar",
				})),
			}, nil
		},
	})
	cli.SetConfigFile(&configfile.ConfigFile{
		NetworksFormat: "{{ .Name }} {{ .Driver }} {{ .Labels }}",
	})
	cmd := newListCommand(cli)
	assert.NoError(t, cmd.Execute())
	golden.Assert(t, cli.OutBuffer().String(), "network-list-with-config-format.golden")
}

func TestNetworkListWithFormat(t *testing.T) {
	cli := test.NewFakeCli(&fakeClient{
		networkListFunc: func(options types.NetworkListOptions) ([]types.NetworkResource, error) {
			return []types.NetworkResource{
				*Network(NetworkName("foo"), NetworkDriver("host")),
				*Network(NetworkName("bar"), NetworkLabels(map[string]string{
					"foo": "bar",
				})),
			}, nil
		},
	})
	cmd := newListCommand(cli)
	cmd.Flags().Set("format", "{{ .Name }} {{ .Driver }} {{ .Labels }}")
	assert.NoError(t, cmd.Execute())
	golden.Assert(t, cli.OutBuffer().String(), "network-list-with-format.golden")
}
