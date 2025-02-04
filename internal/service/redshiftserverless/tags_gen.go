// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package redshiftserverless

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/aws/aws-sdk-go/service/redshiftserverless/redshiftserverlessiface"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// ListTags lists redshiftserverless service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn redshiftserverlessiface.RedshiftServerlessAPI, identifier string) (tftags.KeyValueTags, error) {
	return ListTagsWithContext(context.Background(), conn, identifier)
}

func ListTagsWithContext(ctx context.Context, conn redshiftserverlessiface.RedshiftServerlessAPI, identifier string) (tftags.KeyValueTags, error) {
	input := &redshiftserverless.ListTagsForResourceInput{
		ResourceArn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResourceWithContext(ctx, input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.Tags), nil
}

// []*SERVICE.Tag handling

// Tags returns redshiftserverless service tags.
func Tags(tags tftags.KeyValueTags) []*redshiftserverless.Tag {
	result := make([]*redshiftserverless.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &redshiftserverless.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from redshiftserverless service tags.
func KeyValueTags(tags []*redshiftserverless.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

// UpdateTags updates redshiftserverless service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn redshiftserverlessiface.RedshiftServerlessAPI, identifier string, oldTags interface{}, newTags interface{}) error {
	return UpdateTagsWithContext(context.Background(), conn, identifier, oldTags, newTags)
}
func UpdateTagsWithContext(ctx context.Context, conn redshiftserverlessiface.RedshiftServerlessAPI, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &redshiftserverless.UntagResourceInput{
			ResourceArn: aws.String(identifier),
			TagKeys:     aws.StringSlice(removedTags.IgnoreAWS().Keys()),
		}

		_, err := conn.UntagResourceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &redshiftserverless.TagResourceInput{
			ResourceArn: aws.String(identifier),
			Tags:        Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.TagResourceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
