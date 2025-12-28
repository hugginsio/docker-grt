package main

import (
	"context"
	"dagger/grt/internal/dagger"
	"fmt"
	"time"
)

type Grt struct{}

// RetrieveSource downloads and unpacks the GRT program archive.
func (m *Grt) RetrieveSource() *dagger.Directory {
	archive := dag.HTTP("https://rocket.myluna.de/releases/grtools/GordonsReloadingTool-2021.2040-NIGHTLY-linux.tar.gz")

	return dag.Container().
		From("alpine:latest").
		WithMountedFile("/tmp/grt.tar.gz", archive).
		WithExec([]string{"tar", "-xzf", "/tmp/grt.tar.gz", "-C", "/out"}).
		Directory("/out")
}

// BuildImage builds the Docker image for x86 (linux/amd64).
func (m *Grt) BuildImage(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (*dagger.Container, error) {
	entries, err := source.Directory("_source").Entries(ctx)
	if err != nil || len(entries) == 0 {
		source = source.WithDirectory("_source", m.RetrieveSource())
	}

	return source.
		DockerBuild(dagger.DirectoryDockerBuildOpts{Platform: "linux/amd64"}).
		WithAnnotation("org.opencontainers.image.title", "docker-grt").
		WithAnnotation("org.opencontainers.image.created", time.Now().String()).
		WithAnnotation("org.opencontainers.image.source", "https://github.com/hugginsio/docker-grt").
		WithAnnotation("org.opencontainers.image.licenses", "BSD-3-Clause"), nil
}

func (m *Grt) ReleaseImage(
	ctx context.Context,
	tag string,
	registry string,
	// +optional
	// +default="grt"
	imageName string,
	// +optional
	username string,
	// +optional
	password *dagger.Secret,
) (string, error) {
	source := dag.Git("https://github.com/hugginsio/docker-grt.git", dagger.GitOpts{KeepGitDir: true}).Tag(tag).Tree()
	container, err := m.BuildImage(ctx, source)
	if err != nil {
		return "", err
	}

	serverContainer := container.
		WithLabel("org.opencontainers.image.version", tag).
		WithRegistryAuth(registry, username, password)

	if _, err := serverContainer.Publish(ctx, fmt.Sprintf("%s/%s:%s", registry, imageName, tag)); err != nil {
		return "", err
	}

	if _, err := serverContainer.Publish(ctx, fmt.Sprintf("%s/%s:latest", registry, imageName)); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully released %s/%s:%s", registry, imageName, tag), nil
}
