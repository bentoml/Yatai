# Bento Deployment Custom Resource Definitions


| Field | Type | Description |
| --------- | ----- | ----------------- |
| `apiVersion` | `string` | The version of the schema. Current version is `v1alpha2` |
| `kind` | `string` | The type of the resource. For BentoDeployment with Yatai: `BentoDeployment` |
| `metadata` | `object` | The metadata of the resource. Refer to the Kubernetes API documentation for the fields of the `metadata` field|
| `spec.bento_tag` | `string` | The bento tag for this deployment. **required** |
| `spec.ingress` | `object` | The ingress configuration. |
| `spec.ingress.enabled` | `boolean` | Whether the ingress is enabled. |
| `spec.env` | `object` | The environment variables. |
| `spec.env.key` | `string` | The key of the environment variable. |
| `spec.env.value` | `string` | The value of the environment variable. |
| `spec.autoscaling` | `object` | The autoscaling configuration for the API server |
| `spec.autoscaling.max_replicas` | `int32` |  The maximum number of replicas. |
| `spec.autoscaling.min_replicas` | `int32` |  The minimum number of replicas. |
| `spec.autoscaling.cpu` | `int32` |  The CPU usage. |
| `spec.autoscaling.memory` | `int32` |  The memory usage. |
| `spec.resources.limits.cpu` | `string` |  The CPU limit. |
| `spec.resources.limits.memory` | `string` |  The memory limit. |
| `spec.resources.requests.cpu` | `string` |  The CPU request. |
| `spec.resources.requests.memory` | `string` |  The memory request. |
| `spec.runners` | `array` |  The list of runners resources configuration. |
| `spec.runners[].name` | `string` |  The name of the runner. |
| `spec.runners[].autoscaling.maxReplicats` | `int32` |  The maximum number of replicas. |
| `spec.runners[].autoscaling.minReplicats` | `int32` |  The minimum number of replicas. |
| `spec.runners[].resources` | `object` |  The resources of the runner. |
| `spec.runners[].resources.limits.cpu` | `string` |  The CPU limit. |
| `spec.runners[].resources.limits.memory` | `string` |  The memory limit. |
| `spec.runners[].resources.requests.cpu` | `string` |  The CPU request. |
| `spec.runners[].resources.requests.memory` | `string` |  The memory request. |
| `spec.runners[].autoscaling.maxReplicats` | `int32` |  The maximum number of replicas. |
| `spec.runners[].autoscaling.minReplicats` | `int32` |  The minimum number of replicas. |


### Example of a BentoDeployment

```yaml
apiVersion: v1alpha2
kind: BentoDeployment
metadata:
  name: my-bento-deployment
  namespace: my-namespace
spec:
  bento_tag: iris:0.1.0
  ingress:
    enabled: true
  env:
    key: values
  resources:
    limits:
        cpu: "1"
        memory: "2Gi"
    requests:
        cpu: "1"
        memory: "2Gi"
  autoScaling:
    maxReplicas: 3
    minReplicas: 1
    cpu: 50
    memory: 50
  runners:
  - name: runner1
    resources:
      limits:
        cpu: "1"
        memory: "2Gi"
      requests:
        cpu: "1"
        memory: "2Gi"
      autoscaling:
        maxReplicas: 1
        minReplicas: 1
  - name: runner2
    resources:
      limits:
        cpu: "1"
        memory: "2Gi"
      requests:
        cpu: "1"
        memory: "2Gi"
    autoscaling:
      maxReplicas: 1
      minReplicas: 1
```