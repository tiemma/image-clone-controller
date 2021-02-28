package errors

import "fmt"

type ErrType string

const (
	ImageManifest  ErrType = "IMAGE_MANIFEST"
	ImageReference ErrType = "IMAGE_REFERENCE"
	ImageWrite     ErrType = "IMAGE_WRITE"
	SpecUpdate     ErrType = "SPEC_UPDATE"
	SpecGet        ErrType = "SPEC_GET"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ErrorGettingResource(kind string, err error) error {
	return fmt.Errorf("error occurred getting %s: %s", kind, err)
}

func ErrorUpdatingResource(name string, namespace string, kind string, err error) error {
	return fmt.Errorf("error occurred updating %s %s in namespace %s: %s", kind, name, namespace, err)
}

func ErrorCloningImage(image string, errType ErrType) error {
	return fmt.Errorf("error occurred cloning image %s, reason: %s", image, errType)
}
