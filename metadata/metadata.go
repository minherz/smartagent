package metadata

import (
	"context"

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
		if region, err = metadata.InstanceAttributeValueWithContext(ctx, "region"); err != nil {
			return "", err
		}
	}
	return region, nil
}
