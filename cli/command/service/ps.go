package service

import (
	"context"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/completion"
	"github.com/docker/cli/cli/command/idresolver"
	"github.com/docker/cli/cli/command/node"
	"github.com/docker/cli/cli/command/task"
	"github.com/docker/cli/opts"
	"github.com/moby/moby/api/types/filters"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type psOptions struct {
	services  []string
	quiet     bool
	noResolve bool
	noTrunc   bool
	format    string
	filter    opts.FilterOpt
}

func newPsCommand(dockerCli command.Cli) *cobra.Command {
	options := psOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "ps [OPTIONS] SERVICE [SERVICE...]",
		Short: "List the tasks of one or more services",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.services = args
			return runPS(cmd.Context(), dockerCli, options)
		},
		ValidArgsFunction: completeServiceNames(dockerCli),
	}
	flags := cmd.Flags()
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display task IDs")
	flags.BoolVar(&options.noTrunc, "no-trunc", false, "Do not truncate output")
	flags.BoolVar(&options.noResolve, "no-resolve", false, "Do not map IDs to Names")
	flags.StringVar(&options.format, "format", "", "Pretty-print tasks using a Go template")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")

	flags.VisitAll(func(flag *pflag.Flag) {
		// Set a default completion function if none was set. We don't look
		// up if it does already have one set, because Cobra does this for
		// us, and returns an error (which we ignore for this reason).
		_ = cmd.RegisterFlagCompletionFunc(flag.Name, completion.NoComplete)
	})
	return cmd
}

func runPS(ctx context.Context, dockerCli command.Cli, options psOptions) error {
	apiClient := dockerCli.Client()

	filter, notfound, err := createFilter(ctx, apiClient, options)
	if err != nil {
		return err
	}
	if err := updateNodeFilter(ctx, apiClient, filter); err != nil {
		return err
	}

	tasks, err := apiClient.TaskList(ctx, swarm.TaskListOptions{Filters: filter})
	if err != nil {
		return err
	}

	format := options.format
	if len(format) == 0 {
		format = task.DefaultFormat(dockerCli.ConfigFile(), options.quiet)
	}
	if options.quiet {
		options.noTrunc = true
	}
	if err := task.Print(ctx, dockerCli, tasks, idresolver.New(apiClient, options.noResolve), !options.noTrunc, options.quiet, format); err != nil {
		return err
	}
	if len(notfound) != 0 {
		return errors.New(strings.Join(notfound, "\n"))
	}
	return nil
}

func createFilter(ctx context.Context, apiClient client.APIClient, options psOptions) (filters.Args, []string, error) {
	filter := options.filter.Value()

	serviceIDFilter := filters.NewArgs()
	serviceNameFilter := filters.NewArgs()
	for _, service := range options.services {
		serviceIDFilter.Add("id", service)
		serviceNameFilter.Add("name", service)
	}
	serviceByIDList, err := apiClient.ServiceList(ctx, swarm.ServiceListOptions{Filters: serviceIDFilter})
	if err != nil {
		return filter, nil, err
	}
	serviceByNameList, err := apiClient.ServiceList(ctx, swarm.ServiceListOptions{Filters: serviceNameFilter})
	if err != nil {
		return filter, nil, err
	}

	var notfound []string
	serviceCount := 0
loop:
	// Match services by 1. Full ID, 2. Full name, 3. ID prefix. An error is returned if the ID-prefix match is ambiguous
	for _, service := range options.services {
		for _, s := range serviceByIDList {
			if s.ID == service {
				filter.Add("service", s.ID)
				serviceCount++
				continue loop
			}
		}
		for _, s := range serviceByNameList {
			if s.Spec.Annotations.Name == service {
				filter.Add("service", s.ID)
				serviceCount++
				continue loop
			}
		}
		found := false
		for _, s := range serviceByIDList {
			if strings.HasPrefix(s.ID, service) {
				if found {
					return filter, nil, errors.New("multiple services found with provided prefix: " + service)
				}
				filter.Add("service", s.ID)
				serviceCount++
				found = true
			}
		}
		if !found {
			notfound = append(notfound, "no such service: "+service)
		}
	}
	if serviceCount == 0 {
		return filter, nil, errors.New(strings.Join(notfound, "\n"))
	}
	return filter, notfound, err
}

func updateNodeFilter(ctx context.Context, apiClient client.APIClient, filter filters.Args) error {
	if filter.Contains("node") {
		nodeFilters := filter.Get("node")
		for _, nodeFilter := range nodeFilters {
			nodeReference, err := node.Reference(ctx, apiClient, nodeFilter)
			if err != nil {
				return err
			}
			filter.Del("node", nodeFilter)
			filter.Add("node", nodeReference)
		}
	}
	return nil
}
