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

type BundleAssociate struct {
	BundleId              uint    `json:"bundle_id"`
	AssociatedBundleCache *Bundle `gorm:"foreignkey:BundleId"`
}

func (a *BundleAssociate) GetAssociatedBundleId() uint {
	return a.BundleId
}

func (a *BundleAssociate) GetAssociatedBundleCache() *Bundle {
	return a.AssociatedBundleCache
}

func (a *BundleAssociate) SetAssociatedBundleCache(bundle *Bundle) {
	a.AssociatedBundleCache = bundle
}

type BundleVersionAssociate struct {
	BundleVersionId              uint           `json:"bundle_version_id"`
	AssociatedBundleVersionCache *BundleVersion `gorm:"foreignkey:BundleVersionId"`
}

func (a *BundleVersionAssociate) GetAssociatedBundleVersionId() uint {
	return a.BundleVersionId
}

func (a *BundleVersionAssociate) GetAssociatedBundleVersionCache() *BundleVersion {
	return a.AssociatedBundleVersionCache
}

func (a *BundleVersionAssociate) SetAssociatedBundleVersionCache(bundle *BundleVersion) {
	a.AssociatedBundleVersionCache = bundle
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

func (a *DeploymentAssociate) SetAssociatedDeploymentCache(bundle *Deployment) {
	a.AssociatedDeploymentCache = bundle
}
