package controllersv1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type terminalRecordController struct {
	baseController
}

var TerminalRecordController = terminalRecordController{}

type GetTerminalRecordSchema struct {
	Uid string `path:"uid"`
}

func (s *GetTerminalRecordSchema) GetTerminalRecord(ctx context.Context) (*models.TerminalRecord, error) {
	return services.TerminalRecordService.GetByUid(ctx, s.Uid)
}

func (c *terminalRecordController) canView(ctx context.Context, terminalRecord *models.TerminalRecord) error {
	org, err := services.OrganizationService.GetAssociatedNullableOrganization(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if org != nil {
		return OrganizationController.canView(ctx, org)
	}
	cluster, err := services.ClusterService.GetAssociatedNullableCluster(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if cluster != nil {
		return ClusterController.canView(ctx, cluster)
	}
	deployment, err := services.DeploymentService.GetAssociatedNullableDeployment(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if deployment != nil {
		return DeploymentController.canView(ctx, deployment)
	}
	return nil
}

func (c *terminalRecordController) canUpdate(ctx context.Context, terminalRecord *models.TerminalRecord) error {
	org, err := services.OrganizationService.GetAssociatedNullableOrganization(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if org != nil {
		return OrganizationController.canUpdate(ctx, org)
	}
	cluster, err := services.ClusterService.GetAssociatedNullableCluster(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if cluster != nil {
		return ClusterController.canUpdate(ctx, cluster)
	}
	deployment, err := services.DeploymentService.GetAssociatedNullableDeployment(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if deployment != nil {
		return DeploymentController.canUpdate(ctx, deployment)
	}
	return nil
}

func (c *terminalRecordController) canOperate(ctx context.Context, terminalRecord *models.TerminalRecord) error {
	org, err := services.OrganizationService.GetAssociatedNullableOrganization(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if org != nil {
		return OrganizationController.canOperate(ctx, org)
	}
	cluster, err := services.ClusterService.GetAssociatedNullableCluster(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if cluster != nil {
		return ClusterController.canOperate(ctx, cluster)
	}
	deployment, err := services.DeploymentService.GetAssociatedNullableDeployment(ctx, terminalRecord)
	if err != nil {
		return err
	}
	if deployment != nil {
		return DeploymentController.canOperate(ctx, deployment)
	}
	return nil
}

func (c *terminalRecordController) Get(ctx *gin.Context, schema *GetTerminalRecordSchema) (*schemasv1.TerminalRecordSchema, error) {
	terminalRecord, err := schema.GetTerminalRecord(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, terminalRecord); err != nil {
		return nil, err
	}
	return transformersv1.ToTerminalRecordSchema(ctx, terminalRecord)
}

func (c *terminalRecordController) Download(ctx *gin.Context) error {
	uid := ctx.Param("uid")

	record, err := services.TerminalRecordService.GetByUid(ctx, uid)
	if err != nil {
		return err
	}

	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Header("Content-Disposition", "attachment; filename=record.cast")
	ctx.Header("Content-Type", "application/text/plain")

	meta, err := json.Marshal(record.Meta)
	if err != nil {
		return err
	}
	_content := strings.Join(record.Content, "\n")

	content := string(meta) + "\n" + _content
	ctx.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
	_, err = ctx.Writer.Write([]byte(content))
	return err
}
