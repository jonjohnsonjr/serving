package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func main() {
	flag.Parse()

	image := os.Getenv("IMAGE")
	if image == "" {
		image = "ubuntu"
	}

	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		log.Fatal(err)
	}

	kc, err := k8schain.NewNoClient()
	if err != nil {
		log.Fatal(err)
	}

	desc, err := remote.Get(tag, remote.WithAuthFromKeychain(kc))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s = %s", image, desc.Digest)
}
