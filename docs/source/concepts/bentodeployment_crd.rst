===================
BentoDeployment CRD
===================

:code:`BentoDeployment` is a Kubernetes `CRD <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ defined by :ref:`yatai-deployment <concepts/architecture:yatai-deployment>` component.

It is primarily used to describe bento deployments.

.. list-table:: Specification
    :widths: 25 25 50
    :header-rows: 1


    * - Field
      - Type
      - Description

    * - :code:`apiVersion`
      - :code:`string`
      - The version of the schema. Current version is ``serving.yatai.ai/v2alpha1``

    * - :code:`kind`
      - :code:`string`
      - The type of the resource. ``BentoDeployment``

    * - :code:`metadata`
      - :code:`object`
      - The metadata of the resource. Refer to the Kubernetes API documentation for the fields of the ``metadata`` field

    * - :code:`spec.bento`
      - :code:`string`
      - The name of :ref:`Bento <concepts/bento_crd:Bento CRD>` CR. If this Bento CR not found. yatai-deployment will look for the :ref:`BentoRequest <concepts/bentorequest_crd:BentoRequest CRD>` CR by this name and wait for the BentoRequest CR to generate the Bento CR. **required**

    * - :code:`spec.ingress`
      - :code:`object`
      - The ingress configuration.

    * - :code:`spec.ingress.enabled`
      - :code:`boolean`
      - Whether the ingress is enabled.

    * - :code:`spec.envs`
      - :code:`array`
      - The environment variables.

    * - :code:`spec.envs[].name`
      - :code:`string`
      - The name of the environment variable.

    * - :code:`spec.envs[].value`
      - :code:`string`
      - The value of the environment variable.

    * - :code:`spec.autoscaling`
      - :code:`object`
      - The autoscaling configuration for the API server

    * - :code:`spec.autoscaling.maxReplicas`
      - :code:`int32`
      - The maximum number of replicas.

    * - :code:`spec.autoscaling.minReplicas`
      - :code:`int32`
      - The minimum number of replicas.

    * - :code:`spec.autoscaling.metrics`
      - :code:`object`
      - The `metrics <https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-resource-metrics>`_ definition

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

    * - :code:`spec.runners[].autoscaling.maxReplicats`
      - :code:`int32`
      - The maximum number of replicas.

    * - :code:`spec.runners[].autoscaling.minReplicats`
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

    * - :code:`spec.runners[].envs[].name`
      - :code:`string`
      - The name of the environment variable.

    * - :code:`spec.runners[].envs[].value`
      - :code:`string`
      - The value of the environment variable.


Example of a BentoDeployment
----------------------------

.. code:: yaml

  apiVersion: serving.yatai.ai/v2alpha1
  kind: BentoDeployment
  metadata:
    name: my-bento-deployment
    namespace: my-namespace
  spec:
    bento: iris-1
    ingress:
      enabled: true
    envs:
    - name: foo
      value: bar
    resources:
      limits:
          cpu: 2000m
          memory: "4Gi"
      requests:
          cpu: 1000m
          memory: "2Gi"
    autoscaling:
      maxReplicas: 5
      minReplicas: 1
      metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 60
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
        maxReplicas: 2
        minReplicas: 1
    - name: runner2
      resources:
        limits:
          cpu: 2000m
          memory: "4Gi"
        requests:
          cpu: 1000m
          memory: "2Gi"
      autoscaling:
        maxReplicas: 4
        minReplicas: 1
