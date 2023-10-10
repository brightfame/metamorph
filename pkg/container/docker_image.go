package container

import (
	"fmt"
	"strings"
)

// DockerImage contains basic information about a container image.
type DockerImage struct {
	repo string
	tag  string
}

// ParseDockerImage parses the given string into a DockerImage.
func ParseDockerImage(image string) DockerImage {
	// decode the image tag
	var repo, tag string
	splitted := strings.SplitN(image, "@", 2)
	repoTag := splitted[0]
	idx := strings.LastIndex(repoTag, ":")
	if idx < 0 {
		repo = repoTag
	} else if t := repoTag[idx+1:]; !strings.Contains(t, "/") {
		repo = repoTag[:idx]
		tag = t
	} else if t := repoTag[idx+1:]; strings.Contains(t, "/") {
		repo = image
		tag = "latest"
	}

	if tag != "" {
		return DockerImage{
			repo: repo,
			tag:  tag,
		}
	}
	if i := strings.IndexRune(image, '@'); i > -1 { // Has digest (@sha256:...)
		// when pulling images with a digest, the repository contains the sha hash, and the tag is empty
		// see: https://github.com/fsouza/go-dockerclient/blob/master/image_test.go#L471
		repo = image
	} else {
		tag = "latest"
	}
	return DockerImage{
		repo: repo,
		tag:  tag,
	}
}

// String returns the concatenated string consisting of NAME[:TAG|@DIGEST].
// TODO - add support for DIGEST.
func (i DockerImage) String() string {
	if i.tag == "" {
		return i.repo
	}
	return fmt.Sprintf("%s:%s", i.repo, i.tag)
}
