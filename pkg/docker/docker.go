package docker

import (
	"fmt"
	"github.com/Tiemma/image-clone-controller/pkg/env"
	"github.com/Tiemma/image-clone-controller/pkg/errors"
	"github.com/Tiemma/image-clone-controller/pkg/metrics"
	"github.com/google/go-containerregistry/pkg/authn"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	containerRegistry "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"golang.org/x/mod/semver"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	repoURL                                   = os.Getenv(env.RepoURL)
	ephemeralContainerMinimumSupportedVersion = "v1.16"

	logger = ctrl.Log.WithValues("pkg", "docker")
)

func getAuthConfig() []remote.Option {
	return []remote.Option{
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}
}

func getImageManifest(ref name.Reference) (containerRegistry.Image, error) {
	img, err := remote.Image(ref, getAuthConfig()...)
	if err != nil {
		logger.Error(err, "error occurred getting manifest")
		return nil, err
	}

	return img, err
}

func MustCacheAndModifyPodImage(podSpec *v1.PodSpec, k8sVersion string) (string, errors.ErrType) {
	images := map[name.Reference]remote.Taggable{}

	// Duplicate images are not a problem since their tags would make them differ
	// as opposed to an overwrite if it were only the image url
	for idx, c := range podSpec.Containers {
		ref, err := getReference(c.Image)
		if err != nil {
			return c.Image, errors.ImageReference
		}

		img, err := getImageManifest(ref)
		if err != nil {
			return c.Image, errors.ImageManifest
		}
		images[getCacheImageReference(ref)] = img
		podSpec.Containers[idx].Image = getCacheImageURL(ref)
	}

	if semver.Compare(k8sVersion, ephemeralContainerMinimumSupportedVersion) == 1 {
		for idx, ec := range podSpec.EphemeralContainers {
			ref, err := getReference(ec.Image)
			if err != nil {
				return ec.Image, errors.ImageReference
			}

			img, err := getImageManifest(ref)
			if err != nil {
				return ec.Image, errors.ImageManifest
			}
			images[getCacheImageReference(ref)] = img
			podSpec.EphemeralContainers[idx].Image = getCacheImageURL(ref)
		}
	}

	for idx, ic := range podSpec.InitContainers {
		ref, err := getReference(ic.Image)
		if err != nil {
			return ic.Image, errors.ImageReference
		}

		img, err := getImageManifest(ref)
		if err != nil {
			return ic.Image, errors.ImageManifest
		}
		images[getCacheImageReference(ref)] = img
		podSpec.EphemeralContainers[idx].Image = getCacheImageURL(ref)
	}

	return "", mustCacheImages(images)
}

func getReference(image string) (name.Reference, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		logger.Error(err, "error occurred getting reference")
		return nil, err
	}

	return ref, err
}

func getCacheImageURL(ref name.Reference) string {
	imageURLParts := strings.Split(ref.Name(), "/")

	// Pick the last end of it being the image name and tag without the repository
	return fmt.Sprintf("%s/%s", repoURL, imageURLParts[len(imageURLParts)-1])
}

func getCacheImageReference(ref name.Reference) name.Reference {
	ref, err := getReference(getCacheImageURL(ref))

	// Panic in this case as cache url should always work
	errors.HandleErr(err)

	return ref
}

func mustCacheImages(images map[name.Reference]remote.Taggable) errors.ErrType {
	imageCount := len(images)
	logger.Info(fmt.Sprintf("Caching %d image(s): %s", imageCount, images))

	err := remote.MultiWrite(images, getAuthConfig()...)
	if err != nil {
		logger.Error(err, "error occurred writing images")
		return errors.ImageWrite
	}

	metrics.ImageCloneTotal.Add(float64(imageCount))

	return ""
}
