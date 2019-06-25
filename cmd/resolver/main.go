package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var (
		masterURL      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
		kubeconfig     = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
		namespace      = "default"
		serviceAccount = "default"
		image          = "gcr.io/jonjohnson-test/crane"
	)
	flag.Parse()

	// image := os.Args[1]
	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		log.Fatal(err)
	}

	opt := k8schain.Options{
		Namespace:          namespace,
		ServiceAccountName: serviceAccount,
	}

	cfg, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	client := kubernetes.NewForConfigOrDie(cfg)

	kc, err := k8schain.New(client, opt)
	if err != nil {
		log.Fatal(err)
	}

	desc, err := remote.Get(tag, remote.WithAuthFromKeychain(kc))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(desc.Digest)
}
