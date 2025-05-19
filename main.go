package main

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	config.ContentConfig.AcceptContentTypes = "application/cbor"
	config.ContentConfig.ContentType = "application/cbor"

	// Create the Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}
