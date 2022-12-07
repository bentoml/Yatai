package kube

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func Sleep(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

// KubeClient holds a clientset and a config
type KubeClient struct {
	ClientSet *kubernetes.Clientset
	Config    *rest.Config
}

// ListFilesFromKube gets list of files in path from Kubernetes (recursive)
func ListFilesFromPod(client KubeClient, path, findType, findName string) ([]string, error) {
	pSplit := strings.Split(path, "/")
	if err := validateKubePath(pSplit); err != nil {
		return nil, err
	}
	namespace, podName, containerName, findPath := initKubeVariables(pSplit)
	command := []string{"find", findPath, "-type", findType, "-name", findName}

	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++

		output := new(bytes.Buffer)
		stderr, err := Exec(client, namespace, podName, containerName, command, nil, output)
		if len(stderr) != 0 {
			if attempt == attempts {
				return nil, fmt.Errorf("STDERR: " + (string)(stderr))
			}
			Sleep(attempt)
			continue
		}
		if err != nil {
			if attempt == attempts {
				return nil, err
			}
			Sleep(attempt)
			continue
		}

		lines := strings.Split(output.String(), "\n")
		var outLines []string
		for _, line := range lines {
			if line != "" {
				outLines = append(outLines, strings.Replace(line, findPath, "", 1))
			}
		}

		return outLines, nil
	}

	return nil, nil
}

func DownloadFromPod(client KubeClient, namespace, podName, containerName, fromPath string, writer io.Writer) error {
	command := []string{"cat", fromPath}

	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++

		stderr, err := Exec(client, namespace, podName, containerName, command, nil, writer)
		if attempt == attempts {
			if len(stderr) != 0 {
				return fmt.Errorf("STDERR: " + (string)(stderr))
			}
			if err != nil {
				return err
			}
		}
		if err == nil {
			return nil
		}
		Sleep(attempt)
	}

	return nil
}

func UploadToPod(client KubeClient, namespace, podName, containerName, destPath string, reader io.Reader) error {
	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++
		dir := filepath.Dir(destPath)
		command := []string{"mkdir", "-p", dir}
		stderr, err := Exec(client, namespace, podName, containerName, command, nil, nil)

		if len(stderr) != 0 {
			if attempt == attempts {
				return fmt.Errorf("STDERR: " + (string)(stderr))
			}
			Sleep(attempt)
			continue
		}
		if err != nil {
			if attempt == attempts {
				return err
			}
			Sleep(attempt)
			continue
		}

		command = []string{"touch", destPath}
		stderr, err = Exec(client, namespace, podName, containerName, command, nil, nil)

		if len(stderr) != 0 {
			if attempt == attempts {
				return fmt.Errorf("STDERR: " + (string)(stderr))
			}
			Sleep(attempt)
			continue
		}
		if err != nil {
			if attempt == attempts {
				return err
			}
			Sleep(attempt)
			continue
		}

		command = []string{"cp", "/dev/stdin", destPath}
		stderr, err = Exec(client, namespace, podName, containerName, command, readerWrapper{reader}, nil)

		if len(stderr) != 0 {
			if attempt == attempts {
				return fmt.Errorf("STDERR: " + (string)(stderr))
			}
			Sleep(attempt)
			continue
		}
		if err != nil {
			if attempt == attempts {
				return err
			}
			Sleep(attempt)
			continue
		}
		return nil
	}

	return nil
}

type readerWrapper struct {
	reader io.Reader
}

func (r readerWrapper) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

// Exec executes a command in a given container
func Exec(client KubeClient, namespace, podName, containerName string, command []string, stdin io.Reader, stdout io.Writer) ([]byte, error) {
	clientset, config := client.ClientSet, client.Config

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	scheme := runtime.NewScheme()
	if err := core_v1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "error adding to scheme")
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&core_v1.PodExecOptions{
		Command:   command,
		Container: containerName,
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return nil, errors.Wrap(err, "error while creating Executor")
	}

	var stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error in Stream")
	}

	return stderr.Bytes(), nil
}

func validateKubePath(pathSplit []string) error {
	if len(pathSplit) >= 3 {
		return nil
	}
	return errors.Errorf("illegal path: %s", filepath.Join(pathSplit...))
}

func initKubeVariables(split []string) (string, string, string, string) {
	namespace := split[0]
	pod := split[1]
	container := split[2]
	path := getAbsPath(split[3:]...)

	return namespace, pod, container, path
}

func getAbsPath(path ...string) string {
	return filepath.Join("/", filepath.Join(path...))
}
