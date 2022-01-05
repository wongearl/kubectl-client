package cmd

import (
	"archive/tar"
	"errors"
	"fmt"
	"github.com/wongearl/kubectl-client/client"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
	_ "k8s.io/kubectl/pkg/cmd/cp"
	"os"
	"path"
	"path/filepath"
	"strings"
	_ "unsafe"
)

func (i *pod) copyToPod(srcPath string, destPath string) error {
	restconfig, err, coreclient := client.InitRestClient()

	reader, writer := io.Pipe()
	defer writer.Close()
	if destPath != "/" && strings.HasSuffix(string(destPath[len(destPath)-1]), "/") {
		destPath = destPath[:len(destPath)-1]
	}
	if err := checkDestinationIsDir(i, destPath); err == nil {
		destPath = destPath + "/" + path.Base(srcPath)
	}
	var makeTarerr error
	go func() {
		makeTarerr = makeTar(srcPath, destPath, writer)
		if makeTarerr != nil {
			klog.Errorf("makeTar error %s\n", makeTarerr)
		}
	}()
	var cmdArr []string

	cmdArr = []string{"tar", "-xf", "-"}
	destDir := path.Dir(destPath)
	if len(destDir) > 0 {
		cmdArr = append(cmdArr, "-C", destDir)
	}
	//remote shell.
	req := coreclient.RESTClient().
		Post().
		Namespace(i.Namespace).
		Resource("pods").
		Name(i.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: i.ContainerName,
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		klog.Errorf("error %s\n", err)
		if makeTarerr != nil {
			return errors.New(err.Error() + "," + makeTarerr.Error())
		} else {
			return err
		}

	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: writer,
		Stderr: writer,
		Tty:    false,
	})
	if err != nil {
		klog.Errorf("error %s\n", err)
		message := []byte{}
		_, err := writer.Write(message)

		if makeTarerr != nil {
			return errors.New(err.Error() + "," + makeTarerr.Error() + "," + string(message))
		} else {
			return err
		}
	}
	return nil
}

func checkDestinationIsDir(i *pod, destPath string) error {
	return i.Exec([]string{"test", "-d", destPath})
}

////go:linkname cpMakeTar k8s.io/kubectl/pkg/cmd.makeTar
//func cpMakeTar(srcPath, destPath string, writer io.Writer) error

func (i *pod) copyFromPod(srcPath string, destPath string) error {
	restconfig, err, coreclient := client.InitRestClient()
	reader, outStream := io.Pipe()
	//todo some containers failed : tar: Refusing to write archive contents to terminal (missing -f option?) when execute `tar cf -` in container
	cmdArr := []string{"tar", "cf", "-", srcPath}
	req := coreclient.RESTClient().
		Get().
		Namespace(i.Namespace).
		Resource("pods").
		Name(i.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: i.ContainerName,
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		klog.Errorf("error %s\n", err)
		return err
	}
	go func() {
		defer outStream.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: outStream,
			Stderr: os.Stderr,
			Tty:    false,
		})
		klog.Errorf("exec.Stream error %s\n", err)
	}()

	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	prefix = stripPathShortcuts(prefix)
	destPath = path.Join(destPath, path.Base(prefix))
	err = untarAll(reader, destPath, prefix)
	if err != nil {
		klog.Errorf("untarAll error %s\n", err)
		return err
	}
	return err
}

func untarAll(reader io.Reader, destDir, prefix string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		mode := header.FileInfo().Mode()
		destFileName := filepath.Join(destDir, header.Name[len(prefix):])

		baseName := filepath.Dir(destFileName)
		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName, 0755); err != nil {
				return err
			}
			continue
		}

		evaledPath, err := filepath.EvalSymlinks(baseName)
		if err != nil {
			return err
		}

		if mode&os.ModeSymlink != 0 {
			linkname := header.Linkname

			if !filepath.IsAbs(linkname) {
				_ = filepath.Join(evaledPath, linkname)
			}

			if err := os.Symlink(linkname, destFileName); err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(destFileName)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getPrefix(file string) string {
	return strings.TrimLeft(file, "/")
}

////go:linkname cpStripPathShortcuts k8s.io/kubectl/pkg/cmd.stripPathShortcuts
//func cpStripPathShortcuts(p string) string
