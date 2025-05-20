package main

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	kcpkubernetesclientset "github.com/kcp-dev/client-go/kubernetes"

	"github.com/kcp-dev/kcp/sdk/apis/core"
)

func main() {
	if err := doMain(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func doMain(ctx context.Context) error {
	kubeconfigPath := ".kcp/admin.kubeconfig"

	clientset, err := createClient(kubeconfigPath)
	if err != nil {
		return err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ns := range namespaces.Items {
		fmt.Println(ns.Name)
	}
	return nil
}

func createClient(path string) (*kubernetes.Clientset, error) {
	// Load the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// config.ContentConfig.AcceptContentTypes = "application/cbor"
	// config.ContentConfig.ContentType = "application/cbor"
	config.ContentConfig.AcceptContentTypes = "application/vnd.kubernetes.protobuf"
	config.ContentConfig.ContentType = "application/vnd.kubernetes.protobuf"

	// Create the Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

func CreateAndDeleteCM(ctx context.Context, kubeClient *kcpkubernetesclientset.ClusterClientset, name string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
			Namespace:    metav1.NamespaceDefault,
		},
		Data: map[string]string{
			name: name,
		},
	}

	cmi := kubeClient.Cluster(core.RootCluster.Path()).
		CoreV1().
		ConfigMaps(metav1.NamespaceDefault)

	cm, err := cmi.Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap: %w", err)
	}

	if _, err := cmi.Get(ctx, cm.Name, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("failed to get configmap: %w", err)
	}

	if err := cmi.Delete(ctx, cm.Name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete configmap: %w", err)
	}

	return nil
}

func CreateAndDeleteRBAC(ctx context.Context, kubeClient *kcpkubernetesclientset.ClusterClientset, name string) error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""}, // Core API group
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"create", "update", "delete"},
			},
		},
	}

	cri := kubeClient.Cluster(core.RootCluster.Path()).
		RbacV1().
		ClusterRoles()

	cr, err := cri.Create(ctx, clusterRole, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create cluster role: %w", err)
	}

	if _, err := cri.Get(ctx, cr.Name, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("failed to get cluster role: %w", err)
	}

	if err := cri.Delete(ctx, cr.Name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete cluster role: %w", err)
	}

	return nil
}
