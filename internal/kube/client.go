package kube

import (
	"context"
	"io"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error)
	ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error)
	DeletePod(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error
	GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) (*appsv1.Deployment, error)
	UpdateDeployment(ctx context.Context, namespace string, dep *appsv1.Deployment) (*appsv1.Deployment, error)
	GetEvents(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.EventList, error)
	GetPodLogs(ctx context.Context, namespace, name string, tailLines int64) (string, error)
	WatchPods(ctx context.Context, namespace string) (watch.Interface, error)
	IsFake() bool
}

type realClient struct {
	clientset kubernetes.Interface
}

func NewInClusterClient() (Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &realClient{clientset: clientset}, nil
}

func NewKubeconfigClient(kubeconfigPath string) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &realClient{clientset: clientset}, nil
}

func NewFakeClient() Client {
	return &realClient{clientset: fake.NewSimpleClientset()}
}

func (c *realClient) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *realClient) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	return c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

func (c *realClient) DeletePod(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	return c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, opts)
}

func (c *realClient) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *realClient) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) (*appsv1.Deployment, error) {
	dep, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	dep.Spec.Replicas = &replicas
	return c.clientset.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{})
}

func (c *realClient) UpdateDeployment(ctx context.Context, namespace string, dep *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{})
}

func (c *realClient) GetEvents(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.EventList, error) {
	return c.clientset.CoreV1().Events(namespace).List(ctx, opts)
}

func (c *realClient) GetPodLogs(ctx context.Context, namespace, name string, tailLines int64) (string, error) {
	logOpts := &v1.PodLogOptions{}
	if tailLines > 0 {
		logOpts.TailLines = &tailLines
	}
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, logOpts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return "", err
	}
	return string(logs), nil
}

func (c *realClient) WatchPods(ctx context.Context, namespace string) (watch.Interface, error) {
	return c.clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{})
}

func (c *realClient) IsFake() bool {
	_, ok := c.clientset.(*fake.Clientset)
	return ok
}

var _ = io.Discard
