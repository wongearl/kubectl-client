package cmd

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
	_ "k8s.io/kubectl/pkg/cmd/cp"
	"os"
	"strings"
	_ "unsafe"
)

func (i *pod) Exec(cmd []string) error {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		klog.Error(err)
	}

	coreclient, err := corev1client.NewForConfig(restconfig)
	if err != nil {
		klog.Error(err)
	}

	req := coreclient.RESTClient().
		Post().
		Namespace(i.Namespace).
		Resource("pods").
		Name(i.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: i.ContainerName,
			Command:   cmd,
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
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  strings.NewReader(""),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		return err
	}
	return nil
}

type fakeMassiveDataPty struct {
	message []byte
}

func (s *fakeMassiveDataPty) Read(p []byte) (int, error) {
	return copy(p, s.message), nil
}

func (s *fakeMassiveDataPty) Write(p []byte) (int, error) {
	s.message = append(s.message, p...)
	return len(p), nil
}

type fakeMassiveDataPtyErr struct {
	message []byte
}

func (s *fakeMassiveDataPtyErr) Read(p []byte) (int, error) {
	return copy(p, s.message), nil
}

func (s *fakeMassiveDataPtyErr) Write(p []byte) (int, error) {
	s.message = p
	return len(p), nil
}
func (i *pod) KubectlExec(cmd []string) ([]byte, error) {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		klog.Error(err)
	}

	coreclient, err := corev1client.NewForConfig(restconfig)
	if err != nil {
		klog.Error(err)
	}

	req := coreclient.RESTClient().
		Post().
		Namespace(i.Namespace).
		Resource("pods").
		Name(i.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: i.ContainerName,
			Command:   cmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		klog.Errorf("error %s\n", err)
		return nil, err
	}

	f := &fakeMassiveDataPty{}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  strings.NewReader(""),
		Stdout: f,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		return nil, err
	}
	if err != nil {
		klog.Error("err:", err.Error())
	}
	klog.Info("message:", string(f.message))
	return f.message, nil
}
