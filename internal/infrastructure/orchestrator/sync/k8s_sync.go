// internal/infrastructure/sync/k8s_sync.go
package adapters

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"archive/tar"
	"compress/gzip"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"strings"
)

type K8sFileSync struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

func NewK8sFileSync(namespace string) (*K8sFileSync, error) {
	var config *rest.Config
	var err error

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err := os.Stat(kubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	// Configurar el codec para la API
	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sFileSync{
		clientset: clientset,
		config:    config,
		namespace: namespace,
	}, nil
}

func (k *K8sFileSync) SyncToContainer(ctx context.Context, sourcePath string, targetPath string, podName string) error {

	// Verificar permisos del directorio destino
	if err := k.verifyDirectoryPermissions(ctx, podName, filepath.Dir(targetPath)); err != nil {
		return fmt.Errorf("target directory permission check failed: %v", err)
	}

	if err := createDirectoryInPod(ctx, k, podName, targetPath); err != nil {
		return fmt.Errorf("error creating directory in pod create directory: %s %v", targetPath, err)
	}

	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(k.namespace).
		SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error adding to scheme: %v", err)
	}

	paramCodec := runtime.NewParameterCodec(scheme)

	option := &corev1.PodExecOptions{
		Command: []string{"tar", "xf", "-", "-C", targetPath},
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}

	req.VersionedParams(
		option,
		paramCodec,
	)

	executor, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating SPDY executor: %v", err)
	}

	var mkdirStdout, mkdirStderr bytes.Buffer
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &mkdirStdout,
		Stderr: &mkdirStderr,
	})

	if err != nil {
		return fmt.Errorf("error executing command tar: %v, stderr: %s", err, mkdirStderr.String())
	}

	reader, writer := io.Pipe()

	// Goroutine para crear el archivo tar
	go func() {
		defer writer.Close()
		if err := createTarArchive(sourcePath, writer); err != nil {
			fmt.Printf("Error creating tar: %v\n", err)
		}
	}()

	// Configurar opciones de streaming
	streamOptions := remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	}

	return executor.Stream(streamOptions)
}

func (k *K8sFileSync) SyncFromContainer(ctx context.Context, podName string, sourcePath string, targetPath string, eventStream orchestrator2.TaskEventStream, executionID string) error {
	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(k.namespace).
		SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error adding to scheme: %v", err)
	}

	paramCodec := runtime.NewParameterCodec(scheme)

	option := &corev1.PodExecOptions{
		Command: []string{"tar", "cf", "-", sourcePath},
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}

	req.VersionedParams(option, paramCodec)

	executor, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating SPDY executor: %v", err)
	}

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		err := executor.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdout: writer,
			Stderr: os.Stderr,
		})
		if err != nil {
			eventStream.Publish(orchestrator.TaskEvent{
				ExecutionID: executionID,
				Timestamp:   time.Now(),
				EventType:   orchestrator.EventTypeTaskError,
				Payload:     fmt.Sprintf("Error streaming from pod: %v", err),
			})
		}
	}()

	eventStream.Publish(orchestrator.TaskEvent{
		ExecutionID: executionID,
		Timestamp:   time.Now(),
		EventType:   orchestrator.EventTypeTaskOutput,
		Payload:     "Starting to extract tar archive from pod",
	})

	err = extractTarArchive(reader, targetPath)
	if err != nil {
		eventStream.Publish(orchestrator.TaskEvent{
			ExecutionID: executionID,
			Timestamp:   time.Now(),
			EventType:   orchestrator.EventTypeTaskError,
			Payload:     fmt.Sprintf("Error extracting tar archive: %v", err),
		})
		return err
	}

	eventStream.Publish(orchestrator.TaskEvent{
		ExecutionID: executionID,
		Timestamp:   time.Now(),
		EventType:   orchestrator.EventTypeTaskOutput,
		Payload:     "Successfully extracted tar archive from pod",
	})

	return nil
}

// createTarArchive crea un archivo tar a partir de un directorio fuente
func createTarArchive(sourcePath string, writer io.Writer) error {
	// Crear un nuevo writer tar
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	// Asegurarse de que el sourcePath termine con /
	if !strings.HasSuffix(sourcePath, "/") {
		sourcePath = sourcePath + "/"
	}

	// Caminar por el árbol de directorios
	return filepath.Walk(sourcePath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Obtener la información del archivo
		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return err
		}

		// Modificar el nombre para que sea relativo al directorio raíz
		relPath, err := filepath.Rel(sourcePath, filePath)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		header.Name = relPath

		// Manejar enlaces simbólicos
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(filePath)
			if err != nil {
				return err
			}
			header.Linkname = linkTarget
			header.Size = 0
		}

		// Escribir el header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Si es un archivo regular, escribir el contenido
		if fileInfo.Mode().IsRegular() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// extractTarArchive extrae un archivo tar a un directorio destino
func extractTarArchive(reader io.Reader, targetPath string) error {
	tarReader := tar.NewReader(reader)

	// Crear el directorio destino si no existe
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return err
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Construir la ruta completa del archivo
		target := filepath.Join(targetPath, header.Name)

		// Asegurarse de que el directorio padre existe
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Crear directorio con los permisos especificados
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg, tar.TypeRegA:
			// Crear archivo
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			// Copiar contenido
			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}

		case tar.TypeSymlink:
			// Crear enlace simbólico
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}

		case tar.TypeLink:
			// Crear hard link
			if err := os.Link(filepath.Join(targetPath, header.Linkname), target); err != nil {
				return err
			}

		default:
			return fmt.Errorf("tipo de archivo no soportado en tar: %b en %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

// Función auxiliar para validar rutas
func validatePath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("la ruta debe ser absoluta: %s", path)
	}

	// Verificar que la ruta no contenga '..'
	if strings.Contains(path, "..") {
		return fmt.Errorf("la ruta no debe contener '..': %s", path)
	}

	return nil
}

// Función auxiliar para manejar errores de permisos
func handlePermissionError(err error, path string) error {
	if os.IsPermission(err) {
		return fmt.Errorf("error de permisos al acceder a %s: %v", path, err)
	}
	return err
}

// Versión con soporte para compresión gzip
func createTarGzArchive(sourcePath string, writer io.Writer) error {
	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	return createTarArchive(sourcePath, gzipWriter)
}

func extractTarGzArchive(reader io.Reader, targetPath string) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	return extractTarArchive(gzipReader, targetPath)
}

// Función auxiliar para verificar si un archivo es un tar válido
func isValidTar(reader io.Reader) bool {
	tarReader := tar.NewReader(reader)
	_, err := tarReader.Next()
	return err == nil
}

// Función auxiliar para copiar permisos de archivo
func copyFilePermissions(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// Watch implementa la observación de cambios en el directorio
func (k *K8sFileSync) Watch(ctx context.Context, sourcePath string, targetPath string, podName string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %v", err)
	}
	defer watcher.Close()

	// Añadir recursivamente todos los directorios al watcher
	err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory: %v", err)
	}

	// Canal para debouncing
	debounce := make(chan bool, 1)
	var lastEvent time.Time
	var mutex sync.Mutex

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			mutex.Lock()
			now := time.Now()
			if now.Sub(lastEvent) < 100*time.Millisecond {
				mutex.Unlock()
				continue
			}
			lastEvent = now
			mutex.Unlock()

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				select {
				case debounce <- true:
					// Esperar un poco antes de sincronizar
					go func() {
						time.Sleep(200 * time.Millisecond)
						err := k.SyncToContainer(ctx, sourcePath, targetPath, podName)
						if err != nil {
							fmt.Printf("Error syncing changes: %v\n", err)
						}
						<-debounce
					}()
				default:
					// Ya hay una sincronización en proceso
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("watcher error: %v", err)
		}
	}
}

// ListFiles implementa el listado de archivos en un pod
func (k *K8sFileSync) ListFiles(ctx context.Context, podName string, path string) ([]orchestrator2.FileInfo, error) {
	// Crear el comando para listar archivos
	cmd := []string{"ls", "-la", "--time-style=full-iso", path, "|| true"}

	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(k.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("error creating SPDY executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v, stderr: %s", err, stderr.String())
	}

	// Parsear la salida del comando ls
	var files []orchestrator2.FileInfo
	scanner := bufio.NewScanner(strings.NewReader(stdout.String()))
	// Saltar la primera línea (total)
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parsear la línea del ls
		file, err := parseFileInfo(line)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// Función auxiliar para parsear la salida del ls
func parseFileInfo(line string) (orchestrator2.FileInfo, error) {
	// Formato esperado:
	// -rw-r--r-- 1 root root 1234 2023-08-21 10:00:00.000000000 +0000 filename
	fields := strings.Fields(line)
	if len(fields) < 8 {
		return orchestrator2.FileInfo{}, fmt.Errorf("invalid ls output format: %s", line)
	}

	// Parsear el modo
	mode, err := parseFileMode(fields[0])
	if err != nil {
		return orchestrator2.FileInfo{}, err
	}

	// Parsear el tamaño
	size, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		return orchestrator2.FileInfo{}, err
	}

	// Parsear el tiempo
	timeStr := fields[5] + " " + fields[6]
	modTime, err := time.Parse("2006-01-02 15:04:05.000000000 -0700", timeStr+" "+fields[7])
	if err != nil {
		return orchestrator2.FileInfo{}, err
	}

	// El nombre del archivo es el último campo
	name := fields[len(fields)-1]

	return orchestrator2.FileInfo{
		Name:    name,
		Size:    size,
		Mode:    uint32(mode),
		ModTime: modTime.Unix(),
		IsDir:   fields[0][0] == 'd',
	}, nil
}

// Función auxiliar para parsear los permisos de archivo
func parseFileMode(modeStr string) (os.FileMode, error) {
	var mode os.FileMode
	if len(modeStr) != 10 {
		return 0, fmt.Errorf("invalid mode string length: %s", modeStr)
	}

	// Tipo de archivo
	switch modeStr[0] {
	case 'd':
		mode |= os.ModeDir
	case 'l':
		mode |= os.ModeSymlink
	case '-':
		// archivo regular
	default:
		return 0, fmt.Errorf("unknown file type: %c", modeStr[0])
	}

	// Permisos
	if modeStr[1] == 'r' {
		mode |= 0400
	}
	if modeStr[2] == 'w' {
		mode |= 0200
	}
	if modeStr[3] == 'x' {
		mode |= 0100
	}
	if modeStr[4] == 'r' {
		mode |= 0040
	}
	if modeStr[5] == 'w' {
		mode |= 0020
	}
	if modeStr[6] == 'x' {
		mode |= 0010
	}
	if modeStr[7] == 'r' {
		mode |= 0004
	}
	if modeStr[8] == 'w' {
		mode |= 0002
	}
	if modeStr[9] == 'x' {
		mode |= 0001
	}

	return mode, nil
}

func (k *K8sFileSync) verifyDirectoryPermissions(ctx context.Context, podName string, path string) error {
	// Verificar si el directorio existe y tiene permisos adecuados
	cmd := []string{"sh", "-c", fmt.Sprintf("test -d %s && test -w %s && test -r %s", path, path, path)}

	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(k.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating SPDY executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	}

	// Mostrar la salida estándar
	fmt.Printf("Standard Output:\n%s\n", stdout.String())

	// Mostrar la salida de error estándar
	fmt.Printf("Standard Error:\n%s\n", stderr.String())

	return nil
}

func createDirectoryInPod(ctx context.Context, k *K8sFileSync, podName string, targetPath string) error {
	mkdirCmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p %s", targetPath)}

	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(k.namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: mkdirCmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating SPDY executor for mkdir: %v", err)
	}

	var mkdirStdout, mkdirStderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &mkdirStdout,
		Stderr: &mkdirStderr,
	})
	if err != nil {
		return fmt.Errorf("error creating directory: %v, stderr: %s", err, mkdirStderr.String())
	}

	return nil
}
