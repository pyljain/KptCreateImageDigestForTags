package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var value string

func main() {
	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error {
		for i, item := range resourceList.Items {
			metadata, err := item.GetMeta()
			if err != nil {
				return err
			}

			if metadata.Kind == "Deployment" {

				containers, err := item.Pipe(yaml.Lookup("spec", "template", "spec", "containers"))
				if err != nil {
					return err
				}

				containers.VisitElements(func(n *yaml.RNode) error {
					image, err := n.Field("image").Value.String()
					if err != nil {
						return err
					}

					image = strings.TrimSuffix(image, "\n")

					if !strings.Contains(image, "sha256") {
						repo := strings.Split(image, ":")[0]

						digest, err := getDigest(image)
						if err != nil {
							fmt.Fprintf(os.Stderr, "err = %s", err.Error())
							return err
						}

						err = n.PipeE(yaml.Lookup("image"), yaml.FieldSetter{StringValue: fmt.Sprintf("%s@%s", repo, digest)})
						if err != nil {
							fmt.Fprintf(os.Stderr, "err = %s", err.Error())
							return err
						}
					}

					return nil
				})

			}

			if err := resourceList.Items[i].PipeE(yaml.SetAnnotation("value", value)); err != nil {
				return err
			}
		}
		return nil
	})
	cmd.Flags().StringVar(&value, "value", "", "flag value")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getDigest(image string) (string, error) {
	m, err := getManifest(image)
	if err != nil {
		return "", err
	}

	return m.Digest.String(), nil
}

func getManifest(r string) (*remote.Descriptor, error) {
	ref, err := name.ParseReference(r)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %v", r, err)
	}
	return remote.Get(ref)
}
