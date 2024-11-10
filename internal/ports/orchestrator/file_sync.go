package ports

import (
	"context"
)

type FileSyncService interface {
	// SyncToContainer sincroniza archivos desde el origen al contenedor/pod destino
	SyncToContainer(ctx context.Context, sourcePath string, targetPath string, containerId string) error

	// SyncFromContainer sincroniza archivos desde el contenedor/pod al destino
	SyncFromContainer(ctx context.Context, containerId string, sourcePath string, targetPath string) error

	// Watch observa cambios en el directorio y sincroniza automáticamente
	Watch(ctx context.Context, sourcePath string, targetPath string, containerId string) error

	// ListFiles lista los archivos en un path dentro del contenedor/pod
	ListFiles(ctx context.Context, containerId string, path string) ([]FileInfo, error)
}

type FileInfo struct {
	Name    string
	Size    int64
	Mode    uint32
	ModTime int64
	IsDir   bool
}
