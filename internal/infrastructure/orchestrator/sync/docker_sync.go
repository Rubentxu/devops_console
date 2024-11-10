// internal/infrastructure/sync/docker_sync.go
package adapters

import (
	"archive/tar"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DockerFileSync struct {
	client *client.Client
}

func NewDockerFileSync() (*DockerFileSync, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerFileSync{client: cli}, nil
}

func (d *DockerFileSync) SyncToContainer(ctx context.Context, sourcePath string, targetPath string, containerId string) error {
	// Crear un archivo tar en memoria
	tarBuffer := new(strings.Builder)
	err := d.createTar(sourcePath, tarBuffer)
	if err != nil {
		return fmt.Errorf("error creating tar: %v", err)
	}

	// Copiar el archivo tar al contenedor
	return d.client.CopyToContainer(ctx, containerId, targetPath, strings.NewReader(tarBuffer.String()), types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
}

func (d *DockerFileSync) SyncFromContainer(ctx context.Context, containerId string, sourcePath string, targetPath string) error {
	// Obtener el contenido del contenedor
	reader, _, err := d.client.CopyFromContainer(ctx, containerId, sourcePath)
	if err != nil {
		return fmt.Errorf("error copying from container: %v", err)
	}
	defer reader.Close()

	// Extraer el contenido al sistema de archivos local
	return d.extractTar(reader, targetPath)
}

func (d *DockerFileSync) Watch(ctx context.Context, sourcePath string, targetPath string, containerId string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					err := d.SyncToContainer(ctx, sourcePath, targetPath, containerId)
					if err != nil {
						fmt.Printf("Error syncing changes: %v\n", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Error watching files: %v\n", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return watcher.Add(sourcePath)
}

func (d *DockerFileSync) createTar(sourcePath string, writer io.Writer) error {
	tw := tar.NewWriter(writer)
	defer tw.Close()

	return filepath.Walk(sourcePath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourcePath, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

func (d *DockerFileSync) extractTar(reader io.Reader, targetPath string) error {
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(targetPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}

	return nil
}