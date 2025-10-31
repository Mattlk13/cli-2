package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/containerd/platforms"
	"github.com/distribution/reference"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/completion"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/cli/cli/trust"
	"github.com/docker/cli/internal/jsonstream"
	"github.com/moby/moby/api/pkg/authconfig"
	registrytypes "github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

// pullOptions defines what and how to pull.
type pullOptions struct {
	remote    string
	all       bool
	platform  string
	quiet     bool
	untrusted bool
}

// newPullCommand creates a new `docker pull` command
func newPullCommand(dockerCLI command.Cli) *cobra.Command {
	var opts pullOptions

	cmd := &cobra.Command{
		Use:   "pull [OPTIONS] NAME[:TAG|@DIGEST]",
		Short: "Download an image from a registry",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.remote = args[0]
			return runPull(cmd.Context(), dockerCLI, opts)
		},
		Annotations: map[string]string{
			"category-top": "5",
			"aliases":      "docker image pull, docker pull",
		},
		// Complete with local images to help pulling the latest version
		// of images that are in the image cache.
		ValidArgsFunction:     completion.ImageNames(dockerCLI, 1),
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.all, "all-tags", "a", false, "Download all tagged images in the repository")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Suppress verbose output")
	flags.BoolVar(&opts.untrusted, "disable-content-trust", !trust.Enabled(), "Skip image verification")
	flags.StringVar(&opts.platform, "platform", os.Getenv("DOCKER_DEFAULT_PLATFORM"), "Set platform if server is multi-platform capable")
	_ = flags.SetAnnotation("platform", "version", []string{"1.32"})
	_ = cmd.RegisterFlagCompletionFunc("platform", completion.Platforms())

	return cmd
}

// runPull performs a pull against the engine based on the specified options
func runPull(ctx context.Context, dockerCLI command.Cli, opts pullOptions) error {
	distributionRef, err := reference.ParseNormalizedNamed(opts.remote)
	switch {
	case err != nil:
		return err
	case opts.all && !reference.IsNameOnly(distributionRef):
		return errors.New("tag can't be used with --all-tags/-a")
	case !opts.all && reference.IsNameOnly(distributionRef):
		distributionRef = reference.TagNameOnly(distributionRef)
		if tagged, ok := distributionRef.(reference.Tagged); ok && !opts.quiet {
			_, _ = fmt.Fprintln(dockerCLI.Out(), "Using default tag:", tagged.Tag())
		}
	}

	if opts.platform != "" {
		// TODO(thaJeztah): add a platform option-type / flag-type.
		if _, err = platforms.Parse(opts.platform); err != nil {
			return err
		}
	}

	imgRefAndAuth, err := trust.GetImageReferencesAndAuth(ctx, authResolver(dockerCLI), distributionRef.String())
	if err != nil {
		return err
	}

	// Check if reference has a digest
	_, isCanonical := distributionRef.(reference.Canonical)
	if !opts.untrusted && !isCanonical {
		if err := trustedPull(ctx, dockerCLI, imgRefAndAuth, opts); err != nil {
			return err
		}
	} else {
		if err := imagePullPrivileged(ctx, dockerCLI, imgRefAndAuth.Reference(), imgRefAndAuth.AuthConfig(), opts); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintln(dockerCLI.Out(), imgRefAndAuth.Reference().String())
	return nil
}

// imagePullPrivileged pulls the image and displays it to the output
func imagePullPrivileged(ctx context.Context, dockerCLI command.Cli, ref reference.Named, authConfig *registrytypes.AuthConfig, opts pullOptions) error {
	encodedAuth, err := authconfig.Encode(*authConfig)
	if err != nil {
		return err
	}
	var ociPlatforms []ocispec.Platform
	if opts.platform != "" {
		// Already validated.
		ociPlatforms = append(ociPlatforms, platforms.MustParse(opts.platform))
	}

	responseBody, err := dockerCLI.Client().ImagePull(ctx, reference.FamiliarString(ref), client.ImagePullOptions{
		RegistryAuth:  encodedAuth,
		PrivilegeFunc: nil,
		All:           opts.all,
		Platforms:     ociPlatforms,
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	out := dockerCLI.Out()
	if opts.quiet {
		out = streams.NewOut(io.Discard)
	}
	return jsonstream.Display(ctx, responseBody, out)
}
