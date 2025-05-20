package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/cbor"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"

	kcpkubernetesclientset "github.com/kcp-dev/client-go/kubernetes"
)

var contentTypes = []string{
	runtime.ContentTypeJSON,
	runtime.ContentTypeYAML,
	runtime.ContentTypeProtobuf,
	runtime.ContentTypeCBOR,
	// cbor-seq is only used for streaming cbor, not for normal requests
	// runtime.ContentTypeCBORSequence,
}

func client(path, contentType string) (*kcpkubernetesclientset.ClusterClientset, error) {
	// Load the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	kubeClient, err := kcpkubernetesclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create KCP client: %w", err)
	}

	config.ContentConfig.AcceptContentTypes = contentType
	config.ContentConfig.ContentType = contentType

	// Disable any rate limiting for the benchmark
	config.RateLimiter = flowcontrol.NewFakeAlwaysRateLimiter()
	config.QPS = math.MaxFloat32
	config.Burst = math.MaxInt64

	return kubeClient, nil
}

func benchmark(b *testing.B, fn func(context.Context, *kcpkubernetesclientset.ClusterClientset, string) error) {
	for _, serializer := range contentTypes {
		b.Run(serializer, func(b *testing.B) {
			kubeClient, err := client(".kcp/admin.kubeconfig", serializer)
			if err != nil {
				b.Fatalf("error creating kube client: %v", err)
			}

			logger := klog.FromContext(b.Context()).WithSink(nil)
			ctx := klog.NewContext(b.Context(), logger)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					name := strings.ToLower(strings.ReplaceAll(b.Name(), "/", "-"))
					if err := fn(ctx, kubeClient, name); err != nil {
						b.Errorf("error running benchmark function: %v", err)
					}
				}
			})
		})
	}
}

func BenchmarkWithClientCM(b *testing.B) {
	benchmark(b, CreateAndDeleteCM)
}

func BenchmarkWithClientRBAC(b *testing.B) {
	benchmark(b, CreateAndDeleteRBAC)
}

var serializers = map[string]runtime.Serializer{}

func init() {

	serializers["json"] = json.NewSerializerWithOptions(
		json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme,
		json.SerializerOptions{Yaml: false, Pretty: false, Strict: true},
	)

	serializers["yaml"] = json.NewSerializerWithOptions(
		json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme,
		json.SerializerOptions{Yaml: true, Pretty: false, Strict: true},
	)

	serializers["protobuf"] = protobuf.NewSerializer(scheme.Scheme, scheme.Scheme)

	serializers["cbor"] = cbor.NewSerializer(scheme.Scheme, scheme.Scheme, cbor.Strict(true))
}

func BenchmarkSerialization(b *testing.B) {
	clusterRoleGVK := schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "ClusterRole",
	}

	for contentType, serializer := range serializers {
		b.Run(contentType, func(b *testing.B) {
			name := strings.ToLower(strings.ReplaceAll(b.Name(), "/", "-"))
			clusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
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

			buf := &bytes.Buffer{}
			if err := serializer.Encode(clusterRole, buf); err != nil {
				b.Fatalf("error encoding object: %v", err)
			}

			_, _, err := serializer.Decode(buf.Bytes(), &clusterRoleGVK, nil)
			if err != nil {
				b.Fatalf("error decoding object: %v", err)
			}
		})
	}
}
