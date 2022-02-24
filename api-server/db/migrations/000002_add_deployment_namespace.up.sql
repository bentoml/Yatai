DROP INDEX "uk_deployment_clusterId_name";
CREATE UNIQUE INDEX "uk_deployment_clusterId_kubeNamespace_name" ON "deployment" ("cluster_id", "kube_namespace", "name");

