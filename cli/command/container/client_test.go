package container

import (
	"context"
	"io"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type fakeClient struct {
	client.Client
	inspectFunc         func(string) (container.InspectResponse, error)
	execInspectFunc     func(execID string) (client.ExecInspect, error)
	execCreateFunc      func(containerID string, options client.ExecCreateOptions) (container.ExecCreateResponse, error)
	createContainerFunc func(config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string) (container.CreateResponse, error)
	containerStartFunc      func(containerID string, options client.ContainerStartOptions) error
	imageCreateFunc         func(ctx context.Context, parentReference string, options client.ImageCreateOptions) (io.ReadCloser, error)
	infoFunc                func() (system.Info, error)
	containerStatPathFunc   func(containerID, path string) (container.PathStat, error)
	containerCopyFromFunc   func(containerID, srcPath string) (io.ReadCloser, container.PathStat, error)
	logFunc                 func(string, client.ContainerLogsOptions) (io.ReadCloser, error)
	waitFunc                func(string) (<-chan container.WaitResponse, <-chan error)
	containerListFunc       func(client.ContainerListOptions) ([]container.Summary, error)
	containerExportFunc     func(string) (io.ReadCloser, error)
	containerExecResizeFunc func(id string, options client.ContainerResizeOptions) error
	containerRemoveFunc     func(ctx context.Context, containerID string, options client.ContainerRemoveOptions) error
	containerRestartFunc    func(ctx context.Context, containerID string, options client.ContainerStopOptions) error
	containerStopFunc       func(ctx context.Context, containerID string, options client.ContainerStopOptions) error
	containerKillFunc       func(ctx context.Context, containerID, signal string) error
	containerPruneFunc      func(ctx context.Context, options client.ContainerPruneOptions) (client.ContainerPruneResult, error)
	containerAttachFunc     func(ctx context.Context, containerID string, options client.ContainerAttachOptions) (client.HijackedResponse, error)
	containerDiffFunc       func(ctx context.Context, containerID string) ([]container.FilesystemChange, error)
	containerRenameFunc     func(ctx context.Context, oldName, newName string) error
	containerCommitFunc     func(ctx context.Context, container string, options client.ContainerCommitOptions) (container.CommitResponse, error)
	containerPauseFunc      func(ctx context.Context, container string) error
	Version                 string
}

func (f *fakeClient) ContainerList(_ context.Context, options client.ContainerListOptions) ([]container.Summary, error) {
	if f.containerListFunc != nil {
		return f.containerListFunc(options)
	}
	return []container.Summary{}, nil
}

func (f *fakeClient) ContainerInspect(_ context.Context, containerID string) (container.InspectResponse, error) {
	if f.inspectFunc != nil {
		return f.inspectFunc(containerID)
	}
	return container.InspectResponse{}, nil
}

func (f *fakeClient) ContainerExecCreate(_ context.Context, containerID string, config client.ExecCreateOptions) (container.ExecCreateResponse, error) {
	if f.execCreateFunc != nil {
		return f.execCreateFunc(containerID, config)
	}
	return container.ExecCreateResponse{}, nil
}

func (f *fakeClient) ContainerExecInspect(_ context.Context, execID string) (client.ExecInspect, error) {
	if f.execInspectFunc != nil {
		return f.execInspectFunc(execID)
	}
	return client.ExecInspect{}, nil
}

func (*fakeClient) ContainerExecStart(context.Context, string, client.ExecStartOptions) error {
	return nil
}

func (f *fakeClient) ContainerCreate(
	_ context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	platform *ocispec.Platform,
	containerName string,
) (container.CreateResponse, error) {
	if f.createContainerFunc != nil {
		return f.createContainerFunc(config, hostConfig, networkingConfig, platform, containerName)
	}
	return container.CreateResponse{}, nil
}

func (f *fakeClient) ContainerRemove(ctx context.Context, containerID string, options client.ContainerRemoveOptions) error {
	if f.containerRemoveFunc != nil {
		return f.containerRemoveFunc(ctx, containerID, options)
	}
	return nil
}

func (f *fakeClient) ImageCreate(ctx context.Context, parentReference string, options client.ImageCreateOptions) (io.ReadCloser, error) {
	if f.imageCreateFunc != nil {
		return f.imageCreateFunc(ctx, parentReference, options)
	}
	return nil, nil
}

func (f *fakeClient) Info(_ context.Context) (system.Info, error) {
	if f.infoFunc != nil {
		return f.infoFunc()
	}
	return system.Info{}, nil
}

func (f *fakeClient) ContainerStatPath(_ context.Context, containerID, path string) (container.PathStat, error) {
	if f.containerStatPathFunc != nil {
		return f.containerStatPathFunc(containerID, path)
	}
	return container.PathStat{}, nil
}

func (f *fakeClient) CopyFromContainer(_ context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error) {
	if f.containerCopyFromFunc != nil {
		return f.containerCopyFromFunc(containerID, srcPath)
	}
	return nil, container.PathStat{}, nil
}

func (f *fakeClient) ContainerLogs(_ context.Context, containerID string, options client.ContainerLogsOptions) (io.ReadCloser, error) {
	if f.logFunc != nil {
		return f.logFunc(containerID, options)
	}
	return nil, nil
}

func (f *fakeClient) ClientVersion() string {
	return f.Version
}

func (f *fakeClient) ContainerWait(_ context.Context, containerID string, _ container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	if f.waitFunc != nil {
		return f.waitFunc(containerID)
	}
	return nil, nil
}

func (f *fakeClient) ContainerStart(_ context.Context, containerID string, options client.ContainerStartOptions) error {
	if f.containerStartFunc != nil {
		return f.containerStartFunc(containerID, options)
	}
	return nil
}

func (f *fakeClient) ContainerExport(_ context.Context, containerID string) (io.ReadCloser, error) {
	if f.containerExportFunc != nil {
		return f.containerExportFunc(containerID)
	}
	return nil, nil
}

func (f *fakeClient) ContainerExecResize(_ context.Context, id string, options client.ContainerResizeOptions) error {
	if f.containerExecResizeFunc != nil {
		return f.containerExecResizeFunc(id, options)
	}
	return nil
}

func (f *fakeClient) ContainerKill(ctx context.Context, containerID, signal string) error {
	if f.containerKillFunc != nil {
		return f.containerKillFunc(ctx, containerID, signal)
	}
	return nil
}

func (f *fakeClient) ContainersPrune(ctx context.Context, options client.ContainerPruneOptions) (client.ContainerPruneResult, error) {
	if f.containerPruneFunc != nil {
		return f.containerPruneFunc(ctx, options)
	}
	return client.ContainerPruneResult{}, nil
}

func (f *fakeClient) ContainerRestart(ctx context.Context, containerID string, options client.ContainerStopOptions) error {
	if f.containerRestartFunc != nil {
		return f.containerRestartFunc(ctx, containerID, options)
	}
	return nil
}

func (f *fakeClient) ContainerStop(ctx context.Context, containerID string, options client.ContainerStopOptions) error {
	if f.containerStopFunc != nil {
		return f.containerStopFunc(ctx, containerID, options)
	}
	return nil
}

func (f *fakeClient) ContainerAttach(ctx context.Context, containerID string, options client.ContainerAttachOptions) (client.HijackedResponse, error) {
	if f.containerAttachFunc != nil {
		return f.containerAttachFunc(ctx, containerID, options)
	}
	return client.HijackedResponse{}, nil
}

func (f *fakeClient) ContainerDiff(ctx context.Context, containerID string) ([]container.FilesystemChange, error) {
	if f.containerDiffFunc != nil {
		return f.containerDiffFunc(ctx, containerID)
	}

	return []container.FilesystemChange{}, nil
}

func (f *fakeClient) ContainerRename(ctx context.Context, oldName, newName string) error {
	if f.containerRenameFunc != nil {
		return f.containerRenameFunc(ctx, oldName, newName)
	}

	return nil
}

func (f *fakeClient) ContainerCommit(ctx context.Context, containerID string, options client.ContainerCommitOptions) (container.CommitResponse, error) {
	if f.containerCommitFunc != nil {
		return f.containerCommitFunc(ctx, containerID, options)
	}
	return container.CommitResponse{}, nil
}

func (f *fakeClient) ContainerPause(ctx context.Context, containerID string) error {
	if f.containerPauseFunc != nil {
		return f.containerPauseFunc(ctx, containerID)
	}

	return nil
}
