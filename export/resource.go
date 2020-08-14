package export

import (
	"context"
	"os"
	"runtime"

	"go.opencensus.io/resource"
)

var (
	// TotDetector returns a Detector that combines the resources
	// of the following detectors
	TotDetector = resource.MultiDetector(hostnameDetector, goDetector)
)

func goDetector(context.Context) (*resource.Resource, error) {
	resourceType := "golang"
	labels := map[string]string{
		resourceType + ".version": runtime.Version(),
	}

	return &resource.Resource{
		Type:   resourceType,
		Labels: labels,
	}, nil
}

func hostnameDetector(context.Context) (*resource.Resource, error) {
	resourceType := "host"

	hostname, _ := os.Hostname()

	labels := map[string]string{
		resourceType + ".hostname": hostname,
	}

	return &resource.Resource{
		Type:   resourceType,
		Labels: labels,
	}, nil
}
