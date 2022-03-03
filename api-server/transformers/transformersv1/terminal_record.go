package transformersv1

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/utils"
)

func ToTerminalRecordSchema(ctx context.Context, record *models.TerminalRecord) (*schemasv1.TerminalRecordSchema, error) {
	ss, err := ToTerminalRecordSchemas(ctx, []*models.TerminalRecord{record})
	if err != nil {
		return nil, err
	}
	return ss[0], nil
}

func ToTerminalRecordSchemas(ctx context.Context, records []*models.TerminalRecord) ([]*schemasv1.TerminalRecordSchema, error) {
	resourceSchemasMap, err := ToResourceSchemasMap(ctx, records)
	if err != nil {
		return nil, errors.Wrap(err, "failed to transform terminal records to resource schemas")
	}
	ss := make([]*schemasv1.TerminalRecordSchema, 0, len(records))
	for _, r := range records {
		creatorSchema, err := GetAssociatedCreatorSchema(ctx, r)
		if err != nil {
			return nil, err
		}
		orgSchema, err := GetAssociatedNullableOrganizationSchema(ctx, r)
		if err != nil {
			return nil, err
		}
		clusterSchema, err := GetAssociatedNullableClusterSchema(ctx, r)
		if err != nil {
			return nil, err
		}
		deploymentSchema, err := GetAssociatedNullableDeploymentSchema(ctx, r)
		if err != nil {
			return nil, err
		}
		resource, err := services.ResourceService.Get(ctx, r.ResourceType, r.ResourceId)
		if err != nil && !utils.IsNotFound(err) {
			return nil, err
		}
		var rs *schemasv1.ResourceSchema
		if !utils.IsNotFound(err) {
			rs_, err := ToResourceSchema(ctx, resource)
			if err != nil {
				return nil, errors.Wrap(err, "failed to transform resource to resource schema")
			}
			rs = &rs_
		}
		resourceSchema, ok := resourceSchemasMap[r.GetUid()]
		if !ok {
			return nil, errors.Errorf("resource schema not found for terminal record %s", r.GetUid())
		}
		ss = append(ss, &schemasv1.TerminalRecordSchema{
			ResourceSchema: resourceSchema,
			Creator:        creatorSchema,
			Organization:   orgSchema,
			Cluster:        clusterSchema,
			Deployment:     deploymentSchema,
			Resource:       rs,
			PodName:        r.PodName,
			ContainerName:  r.ContainerName,
		})
	}

	return ss, nil
}
