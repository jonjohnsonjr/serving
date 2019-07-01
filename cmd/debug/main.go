package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// Loads auth plugin.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	kubeclient, err := buildClient()
	if err != nil {
		log.Fatalf("building client: %v", err)
	}

	if err := diagnoseCluster(kubeclient); err != nil {
		log.Fatalf("diagnosing cluster: %v", err)
	}
}

func diagnoseCluster(kubeclient kubernetes.Interface) error {
	ns := "knative-serving"
	pods, err := kubeclient.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("getting pods: %v", err)
	}

	for _, pod := range pods.Items {
		fmt.Println(pod.Status.Phase)
	}

	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// Cribbed from:
// https://github.com/kubernetes/client-go/blob/98902b2ea1c23a3af79f9cf29692c59540ae82c5/examples/out-of-cluster-client-configuration/main.go
func buildClient() (kubernetes.Interface, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	masterURL := flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(cfg)
}
