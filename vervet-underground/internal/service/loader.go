package service

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Loader is a function that can load services into the service Registry.
type Loader func(context.Context) ([]string, error)

// StaticServiceLoader returns a static set of services.
func StaticServiceLoader(services []string) Loader {
	return func(context.Context) ([]string, error) {
		return services, nil
	}
}

// KubeServiceLoader discovers services VU should scrape from a kube cluster.
// It fetches services from all namespaces within the cluster and looks for AnnotationVUScrape annotation to check whether vervet-underground should scrape
// this service. The function assumes VU is operating in cluster.
func KubeServiceLoader() Loader {
	return func(ctx context.Context) ([]string, error) {
		config, err := rest.InClusterConfig() // TODO: non-default service account
		if err != nil {
			return nil, fmt.Errorf("failed to get incluster config: %w", err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		services, err := clientset.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get services: %w", err)
		}

		var serviceNames []string
		for _, service := range services.Items {
			if v := service.Annotations[AnnotationVUScrape]; v == "true" {
				port := "80"
				if p, ok := service.Annotations[AnnotationVUPort]; ok {
					port = p
				}
				serviceNames = append(serviceNames, fmt.Sprintf("http://%s.%s:%s", service.Name, service.Namespace, port))
			}
		}
		return serviceNames, nil
	}
}
