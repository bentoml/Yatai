# Helm chart configuration

THe following table lists the configurable parameters for the Yatai Helm chart and their default values.


## Yatai Kuberentes service configuration
| Parameter | Default | Description |
| --------- | ------- | ----------- |
| `replicatCount` | `1` | The image to use for the Yatai Helm chart. |
| `registry` | `quay.io/bentoml` | The version of the Yatai Helm chart to use. |
| `image.repository` | `yatai` | The namespace to use for the Yatai Helm chart. |
| `image.pullPolicy` | `IfNotPresent` | The image pull policy to use for the Yatai Helm chart. |
| `image.tag` | `latest` | The tag to use for the Yatai Helm chart. |
| `imagePullSecrets` | `[]` | The image pull secrets to use for the Yatai Helm chart. |
| `nameOverride` | `""` | The name of the Yatai Helm chart. |
| `fullnameOverride` | `""` | The full name of the Yatai Helm chart. |
| `podAnnotations` | `{}` | The annotations to use for the Yatai Helm chart pods. |
| `podSecurityContext` | `{}` | The pod security context to use for the Yatai Helm chart pods. |
| `autoscaling.enabled` | `false` | Whether to enable autoscaling for the Yatai Helm chart. |
| `autoscaling.minReplicas` | `1` | The minimum number of replicas for the Yatai Helm chart. |
| `autoscaling.maxReplicas` | `100` | The maximum number of replicas for the Yatai Helm chart. |
| `autoscaling.targetCPUUtilizationPercentage` | `80` | The target CPU utilization percentage for the Yatai Helm chart. |


## External S3 configuration
| Parameter | Default | Description |
| --------- | ------- | ----------- |
| `externalS3.enabled` | `false` | Whether to enable the S3 storage for the Yatai Helm chart. |
| `externalS3.endpoint` | `""` | The endpoint of the S3 storage for the Yatai Helm chart. |
| `externalS3.region` | `""` | The region of the S3 storage for the Yatai Helm chart. |
| `externalS3.bucketName` | `""` | The bucket name of the S3 storage for the Yatai Helm chart. |
| `externalS3.secure` | `true` | Whether to use secure S3 storage for the Yatai Helm chart. |
| `externalS3.existingSecret` | `""` | The name of an existing S3 secret for the Yatai Helm chart. |
| `externalS3.existingSecretAccessKeyKey` | `""` | The key of the S3 secret access key for the Yatai Helm chart. |
| `externalS3.existingSecretSecretKeyKey` | `""` | The secret key of the S3 secret for the Yatai Helm chart. |


## External Docker Registry configuration
| Parameter | Default | Description |
| --------- | ------- | ----------- |
| `externalDockerRegistry.enabled` | `false` | Whether to enable the Docker registry for the Yatai Helm chart. |
| `externalDockerRegistry.server` | `""` | The server of the Docker registry for the Yatai Helm chart. |
| `externalDockerRegistry.username` | `""` | The username of the Docker registry for the Yatai Helm chart. |
| `externalDockerRegistry.secure` | `true` | Whether to use secure Docker registry for the Yatai Helm chart. |
| `externalDockerRegistry.bentoRepositoryName` | `'yatai-bentos'` | The name of the Bento repository for the Yatai Helm chart. |
| `externalDockerRegistry.modelRepositoryName` | `'yatai-models'` | The name of the Model repository for the Yatai Helm chart. |
| `externalDockerRegistry.existingSecret` | `""` | The name of an existing Docker registry secret for the Yatai Helm chart. |
| `externalDockerRegistry.existingPasswordKey` | `""` | The key of the Docker registry password key for the Yatai Helm chart. |


## Yatai database configuration
| Parameter | Default | Description |
| --------- | ------- | ----------- |
| `postgresql.enabled` | `true` | Whether to enable the PostgreSQL database for the Yatai Helm chart. Set to `false` to use external postgresql. |
| `postgresql.nameOverride` | `""` | The name of the PostgreSQL database for the Yatai Helm chart. |
| `postgresql.postgresqlDatabase` | `yatai` | The PostgreSQL database for the Yatai Helm chart. |
| `postgresql.postgresqlUsername` | `postgres` | The PostgreSQL username for the Yatai Helm chart. |
| `postgresql.existingSecret` | `""` | The name of an existing PostgreSQL secret for the Yatai Helm chart. |
| `externalPostgresql.host` | `localhost` | The host of the external PostgreSQL database for the Yatai Helm chart. |
| `externalPostgresql.port` | `5432` | The port of the external PostgreSQL database for the Yatai Helm chart. |
| `externalPostgresql.database` | `yatai` | The PostgreSQL database for the Yatai Helm chart. |
| `externalPostgresql.user` | `yatai` | The PostgreSQL username for the Yatai Helm chart. |
| `externalPostgresql.existingSecret` | `""` | The name of an existing PostgreSQL secret for the Yatai Helm chart. |
| `externalPostgresql.existingSecretPasswordKey` | `""` | The key of the PostgreSQL secret for the Yatai Helm chart. |


## BentoDeployment ingress configuration
| Parameter | Default | Description |
| --------- | ------- | ----------- |
| `ingress.enabled` | `true` | Whether to enable the ingress for the Yatai Helm chart. |
| `ingress.className` | `yatai-ingress` | The class name of the ingress for the Yatai Helm chart. |
| `ingress.hosts[0].host` | `yatai.127.0.0.1.sslip.io` | The host of the ingress for the Yatai Helm chart. |
| `ingress.hosts[0].paths[0]` | `/` | The path of the ingress for the Yatai Helm chart. |
| `ingress.tls` | `[]` | The TLS configuration of the ingress for the Yatai Helm chart. |


## Common Kubernetes configuration
| Parameter | Default | Description |
| --------- | ------- | ----------- |
| `serviceAccount.create` | `true` | Whether to create a service account for the Yatai Helm chart. |
| `serviceAccount.annotations` | `{}` | The annotations to use for the Yatai Helm chart service account. |
| `serviceAccount.name` | `""` | The name of the Yatai Helm chart service account. |
| `config` | `{}` | The Yatai server configuration. See [Yatai Repo](https://github.com/bentom/yatai) for more details. |
| `securityContext` | `{}` | The security context to use for the Yatai Helm chart pods. |
| `service.type` | `NodePort` | The type of service to use for the Yatai Helm chart. |
| `service.port` | `80` | The port of the Yatai Helm chart. |
| `service.nodePort` | `8080` | The node port of the Yatai Helm chart. |
| `resources` | `{}` | The resources to use for the Yatai Helm chart. |
| `nodeSelector` | `{}` | The node selector to use for the Yatai Helm chart. |
| `tolerations` | `[]` | The tolerations to use for the Yatai Helm chart. |
| `affinity` | `{}` | The affinity to use for the Yatai Helm chart. |