====================
Bento Deployment CRD
====================

:code:`BentoDeployment` is a Kubernetes `CRD <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ defined by :code:`yatai-deployment` that is primarily used to describe bento deployments.

.. list-table:: Specification
    :widths: 25 25 50
    :header-rows: 1


    * - Field
      - Type
      - Description

    * - :code:`apiVersion`
      - :code:`string`
      - The version of the schema. Current version is `v1alpha3`

    * - :code:`kind`
      - :code:`string`
      - The type of the resource. For BentoDeployment with Yatai: `BentoDeployment`

    * - :code:`metadata`
      - :code:`object`
      - The metadata of the resource. Refer to the Kubernetes API documentation for the fields of the `metadata` field

    * - :code:`spec.bento_tag`
      - :code:`string`
      - The bento tag for this deployment. **required**

    * - :code:`spec.ingress`
      - :code:`object`
      - The ingress configuration.

    * - :code:`spec.ingress.enabled`
      - :code:`boolean`
      - Whether the ingress is enabled.

    * - :code:`spec.envs`
      - :code:`array`
      - The environment variables.

    * - :code:`spec.envs[].key`
      - :code:`string`
      - The key of the environment variable.

    * - :code:`spec.envs[].value`
      - :code:`string`
      - The value of the environment variable.

    * - :code:`spec.autoscaling`
      - :code:`object`
      - The autoscaling configuration for the API server

    * - :code:`spec.autoscaling.max_replicas`
      - :code:`int32`
      - The maximum number of replicas.

    * - :code:`spec.autoscaling.min_replicas`
      - :code:`int32`
      - The minimum number of replicas.

    * - :code:`spec.autoscaling.cpu`
      - :code:`int32`
      - The CPU usage.

    * - :code:`spec.autoscaling.memory`
      - :code:`int32`
      - The memory usage.

    * - :code:`spec.resources.limits.cpu`
      - :code:`string`
      - The CPU limit.

    * - :code:`spec.resources.limits.memory`
      - :code:`string`
      - The memory limit.

    * - :code:`spec.resources.requests.cpu`
      - :code:`string`
      - The CPU request.

    * - :code:`spec.resources.requests.memory`
      - :code:`string`
      - The memory request.

    * - :code:`spec.runners`
      - :code:`array`
      - The list of runners resources configuration.

    * - :code:`spec.runners[].name`
      - :code:`string`
      - The name of the runner.

    * - :code:`spec.runners[].autoscaling.max_replicats`
      - :code:`int32`
      - The maximum number of replicas.

    * - :code:`spec.runners[].autoscaling.min_replicats`
      - :code:`int32`
      - The minimum number of replicas.

    * - :code:`spec.runners[].resources`
      - :code:`object`
      - The resources of the runner.

    * - :code:`spec.runners[].resources.limits.cpu`
      - :code:`string`
      - The CPU limit.

    * - :code:`spec.runners[].resources.limits.memory`
      - :code:`string`
      - The memory limit.

    * - :code:`spec.runners[].resources.requests.cpu`
      - :code:`string`
      - The CPU request.

    * - :code:`spec.runners[].resources.requests.memory`
      - :code:`string`
      - The memory request.

    * - :code:`spec.runners[].envs`
      - :code:`array`
      - The environment variables.

    * - :code:`spec.runners[].envs[].key`
      - :code:`string`
      - The key of the environment variable.

    * - :code:`spec.runners[].envs[].value`
      - :code:`string`
      - The value of the environment variable.


Example of a BentoDeployment
----------------------------

.. code:: yaml

  apiVersion: serving.yatai.ai/v1alpha3
  kind: BentoDeployment
  metadata:
    name: my-bento-deployment
    namespace: my-namespace
  spec:
    bento_tag: iris:0.1.0
    ingress:
      enabled: true
    envs:
    - key: foo
      value: bar
    resources:
      limits:
          cpu: 2000m
          memory: "4Gi"
      requests:
          cpu: 1000m
          memory: "2Gi"
    autoscaling:
      max_replicas: 5
      min_replicas: 1
      cpu: 50
      memory: 50
    runners:
    - name: runner1
      resources:
        limits:
          cpu: 2000m
          memory: "4Gi"
        requests:
          cpu: 1000m
          memory: "2Gi"
        autoscaling:
          max_replicas: 2
          min_replicas: 1
    - name: runner2
      resources:
        limits:
          cpu: 2000m
          memory: "4Gi"
        requests:
          cpu: 1000m
          memory: "2Gi"
      autoscaling:
        max_replicas: 4
        min_replicas: 1
