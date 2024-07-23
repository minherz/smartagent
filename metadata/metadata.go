package metadata

import (
	"context"
	"strings"

	"cloud.google.com/go/compute/metadata"
)

var (
	projectID string
	region    string
)

func ProjectID(ctx context.Context) (string, error) {
	if projectID == "" {
		var err error
		if projectID, err = metadata.ProjectIDWithContext(ctx); err != nil {
			return "", err
		}
	}
	return projectID, nil
}

func Region(ctx context.Context) (string, error) {
	if region == "" {
		var err error
		if region, err = metadata.GetWithContext(ctx, "instance/region"); err != nil {
			return "", err
		}
		// parse region from fully qualified name projects/<projNum>/regions/<region>
		if pos := strings.LastIndex(region, "/"); pos >= 0 {
			region = region[pos+1:]
		}
	}
	return region, nil
}
