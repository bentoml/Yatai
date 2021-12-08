package models

type UserAssociate struct {
	UserId              uint  `json:"user_id"`
	AssociatedUserCache *User `gorm:"foreignkey:UserId"`
}

func (a *UserAssociate) GetAssociatedUserId() uint {
	return a.UserId
}

func (a *UserAssociate) GetAssociatedUserCache() *User {
	return a.AssociatedUserCache
}

func (a *UserAssociate) SetAssociatedUserCache(user *User) {
	a.AssociatedUserCache = user
}

type CreatorAssociate struct {
	CreatorId              uint  `json:"creator_id"`
	AssociatedCreatorCache *User `gorm:"foreignkey:CreatorId"`
}

func (a *CreatorAssociate) GetAssociatedCreatorId() uint {
	return a.CreatorId
}

func (a *CreatorAssociate) GetAssociatedCreatorCache() *User {
	return a.AssociatedCreatorCache
}

func (a *CreatorAssociate) SetAssociatedCreatorCache(user *User) {
	a.AssociatedCreatorCache = user
}

type UserGroupAssociate struct {
	UserGroupId              uint       `json:"user_group_id"`
	AssociatedUserGroupCache *UserGroup `gorm:"foreignkey:UserGroupId"`
}

func (a *UserGroupAssociate) GetAssociatedUserGroupId() uint {
	return a.UserGroupId
}

func (a *UserGroupAssociate) GetAssociatedUserGroupCache() *UserGroup {
	return a.AssociatedUserGroupCache
}

func (a *UserGroupAssociate) SetAssociatedUserGroupCache(userGroup *UserGroup) {
	a.AssociatedUserGroupCache = userGroup
}

type OrganizationAssociate struct {
	OrganizationId              uint          `json:"organization_id"`
	AssociatedOrganizationCache *Organization `gorm:"foreignkey:OrganizationId"`
}

func (a *OrganizationAssociate) GetAssociatedOrganizationId() uint {
	return a.OrganizationId
}

func (a *OrganizationAssociate) GetAssociatedOrganizationCache() *Organization {
	return a.AssociatedOrganizationCache
}

func (a *OrganizationAssociate) SetAssociatedOrganizationCache(organization *Organization) {
	a.AssociatedOrganizationCache = organization
}

type NullableOrganizationAssociate struct {
	OrganizationId              *uint         `json:"organization_id"`
	AssociatedOrganizationCache *Organization `gorm:"foreignkey:OrganizationId"`
}

func (a *NullableOrganizationAssociate) GetAssociatedOrganizationId() *uint {
	return a.OrganizationId
}

func (a *NullableOrganizationAssociate) GetAssociatedOrganizationCache() *Organization {
	return a.AssociatedOrganizationCache
}

func (a *NullableOrganizationAssociate) SetAssociatedOrganizationCache(organization *Organization) {
	a.AssociatedOrganizationCache = organization
}

type ClusterAssociate struct {
	ClusterId              uint     `json:"cluster_id"`
	AssociatedClusterCache *Cluster `gorm:"foreignkey:ClusterId"`
}

func (a *ClusterAssociate) GetAssociatedClusterId() uint {
	return a.ClusterId
}

func (a *ClusterAssociate) GetAssociatedClusterCache() *Cluster {
	return a.AssociatedClusterCache
}

func (a *ClusterAssociate) SetAssociatedClusterCache(cluster *Cluster) {
	a.AssociatedClusterCache = cluster
}

type NullableClusterAssociate struct {
	ClusterId              *uint    `json:"cluster_id"`
	AssociatedClusterCache *Cluster `gorm:"foreignkey:ClusterId"`
}

func (a *NullableClusterAssociate) GetAssociatedClusterId() *uint {
	return a.ClusterId
}

func (a *NullableClusterAssociate) GetAssociatedClusterCache() *Cluster {
	return a.AssociatedClusterCache
}

func (a *NullableClusterAssociate) SetAssociatedClusterCache(cluster *Cluster) {
	a.AssociatedClusterCache = cluster
}

type BentoRepositoryAssociate struct {
	BentoRepositoryId              uint             `json:"bento_repository_id"`
	AssociatedBentoRepositoryCache *BentoRepository `gorm:"foreignkey:BentoRepositoryId"`
}

func (a *BentoRepositoryAssociate) GetAssociatedBentoRepositoryId() uint {
	return a.BentoRepositoryId
}

func (a *BentoRepositoryAssociate) GetAssociatedBentoRepositoryCache() *BentoRepository {
	return a.AssociatedBentoRepositoryCache
}

func (a *BentoRepositoryAssociate) SetAssociatedBentoRepositoryCache(bentoRepository *BentoRepository) {
	a.AssociatedBentoRepositoryCache = bentoRepository
}

type BentoAssociate struct {
	BentoId              uint   `json:"bento_id"`
	AssociatedBentoCache *Bento `gorm:"foreignkey:BentoId"`
}

func (a *BentoAssociate) GetAssociatedBentoId() uint {
	return a.BentoId
}

func (a *BentoAssociate) GetAssociatedBentoCache() *Bento {
	return a.AssociatedBentoCache
}

func (a *BentoAssociate) SetAssociatedBentoCache(bento *Bento) {
	a.AssociatedBentoCache = bento
}

type DeploymentAssociate struct {
	DeploymentId              uint        `json:"deployment_id"`
	AssociatedDeploymentCache *Deployment `gorm:"foreignkey:DeploymentId"`
}

func (a *DeploymentAssociate) GetAssociatedDeploymentId() uint {
	return a.DeploymentId
}

func (a *DeploymentAssociate) GetAssociatedDeploymentCache() *Deployment {
	return a.AssociatedDeploymentCache
}

func (a *DeploymentAssociate) SetAssociatedDeploymentCache(deployment *Deployment) {
	a.AssociatedDeploymentCache = deployment
}

type NullableDeploymentAssociate struct {
	DeploymentId              *uint       `json:"deployment_id"`
	AssociatedDeploymentCache *Deployment `gorm:"foreignkey:DeploymentId"`
}

func (a *NullableDeploymentAssociate) GetAssociatedDeploymentId() *uint {
	return a.DeploymentId
}

func (a *NullableDeploymentAssociate) GetAssociatedDeploymentCache() *Deployment {
	return a.AssociatedDeploymentCache
}

func (a *NullableDeploymentAssociate) SetAssociatedDeploymentCache(deployment *Deployment) {
	a.AssociatedDeploymentCache = deployment
}

type DeploymentRevisionAssociate struct {
	DeploymentRevisionId              uint                `json:"deployment_revision_id"`
	AssociatedDeploymentRevisionCache *DeploymentRevision `gorm:"foreignkey:DeploymentRevisionId"`
}

func (a *DeploymentRevisionAssociate) GetAssociatedDeploymentRevisionId() uint {
	return a.DeploymentRevisionId
}

func (a *DeploymentRevisionAssociate) GetAssociatedDeploymentRevisionCache() *DeploymentRevision {
	return a.AssociatedDeploymentRevisionCache
}

func (a *DeploymentRevisionAssociate) SetAssociatedDeploymentRevisionCache(deploymentRevision *DeploymentRevision) {
	a.AssociatedDeploymentRevisionCache = deploymentRevision
}

type ModelRepositoryAssociate struct {
	ModelRepositoryId              uint             `json:"model_repository_id"`
	AssociatedModelRepositoryCache *ModelRepository `gorm:"foreignkey:ModelRepositoryId"`
}

func (a *ModelRepositoryAssociate) GetAssociatedModelRepositoryId() uint {
	return a.ModelRepositoryId
}

func (a *ModelRepositoryAssociate) GetAssociatedModelRepositoryCache() *ModelRepository {
	return a.AssociatedModelRepositoryCache
}

func (a *ModelRepositoryAssociate) SetAssociatedModelRepositoryCache(modelRepository *ModelRepository) {
	a.AssociatedModelRepositoryCache = modelRepository
}

type ModelAssociate struct {
	ModelId              uint   `json:"model_id"`
	AssociatedModelCache *Model `gorm:"foreignkey:ModelId"`
}

func (a *ModelAssociate) GetAssociatedModelId() uint {
	return a.ModelId
}

func (a *ModelAssociate) GetAssociatedModelCache() *Model {
	return a.AssociatedModelCache
}

func (a *ModelAssociate) SetAssociatedModelCache(model *Model) {
	a.AssociatedModelCache = model
}
