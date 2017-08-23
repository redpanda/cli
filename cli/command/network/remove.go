package network

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newRemoveCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:     "rm NETWORK [NETWORK...]",
		Aliases: []string{"remove"},
		Short:   "Remove one or more networks",
		Args:    cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(dockerCli, args)
		},
	}
}

const ingressWarning = "WARNING! Before removing the routing-mesh network, " +
	"make sure all the nodes in your swarm run the same docker engine version. " +
	"Otherwise, removal may not be effective and functionality of newly create " +
	"ingress networks will be impaired.\nAre you sure you want to continue?"

func runRemove(dockerCli command.Cli, networks []string) error {
	client := dockerCli.Client()
	ctx := context.Background()

	var errs []string

	for _, name := range networks {
		nw, _, err := client.NetworkInspectWithRaw(ctx, name, types.NetworkInspectOptions{})
		if err != nil {
			errs = append(errs, err.Error())
			continue
		} else if nw.Ingress && !command.PromptForConfirmation(dockerCli.In(), dockerCli.Out(), ingressWarning) {
			continue
		}

		if err := client.NetworkRemove(ctx, name); err != nil {
			errs = append(errs, err.Error())
			continue
		}

		fmt.Fprintf(dockerCli.Out(), "%s\n", name)
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, "\n"))
	}

	return nil
}
