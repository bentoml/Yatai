==============================
Installing yatai-image-builder
==============================

Welcome to yatai-image-builder! You will learn the system requirements, software dependencies, instructions for installing this Yatai component.

See :ref:`yatai-image-builder architecture <concepts/architecture:yatai-image-builder>` for a detailed introduction of the ``yatai-image-builder`` component.

Prerequisites
-------------

- Yatai

  ``yatai-image-builder`` depends on ``yatai`` as the bento registry, you should check the documentation :doc:`yatai` first.

- Kubernetes

  Kubernetes cluster with version 1.20 or newer

  .. note::

    If you do not have a production Kubernetes cluster and want to install ``yatai-image-builder`` for development and testing purposes. You can use `minikube <https://minikube.sigs.k8s.io/docs/start/>`_ to set up a local Kubernetes cluster for testing. If you are using macOS, you should use `hyperkit <https://minikube.sigs.k8s.io/docs/drivers/hyperkit/>`_ driver to prevent the macOS docker desktop `network limitation <https://docs.docker.com/desktop/networking/#i-cannot-ping-my-containers>`_

- Helm

  Yatai uses `Helm <https://helm.sh/docs/intro/using_helm/>`_ to install ``yatai-image-builder``.

Quick Install
-------------

.. note:: This quick installation script can only be used for **development** and **testing** purposes.

This script will automatically install the following dependencies inside the :code:`yatai-image-builder` namespace of the Kubernetes cluster:

* cert-manager (if not already installed)
* docker-registry

.. code:: bash

  DEVEL=true bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-image-builder/main/scripts/quick-install-yatai-image-builder.sh")

.. _yatai-image-builder-installation-steps:

Installation Steps
------------------

.. note::

  If you don't have :code:`kubectl` installed and you are using :code:`minikube`, you can use :code:`minikube kubectl --` instead of :code:`kubectl`, for more details on using it, please check: `minikube kubectl <https://minikube.sigs.k8s.io/docs/commands/kubectl/>`_

1. Create Namespaces
^^^^^^^^^^^^^^^^^^^^

.. code:: bash

  # for yatai-image-builder deployment
  kubectl create ns yatai-image-builder

2. Install Certificate Manager
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. tab-set::

    .. tab-item:: Already installed

      Read the official documentation to verify that it works: `manual-verification <https://cert-manager.io/docs/installation/verify/#manual-verification>`_.

    .. tab-item:: Install cert-manager

      1. Install cert-manager via kubectl

      .. code:: bash

        kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml

      2. Verify the cert-manager installation

      .. code:: bash

        kubectl -n cert-manager get pod

      The output should look like this:

      .. note:: Wait until the status of all pods becomes :code:`Running` before proceeding.

      .. code:: bash

        NAME                                       READY   STATUS    RESTARTS   AGE
        cert-manager-5dd59d9d9b-7js6w              1/1     Running   0          60s
        cert-manager-cainjector-8696fc9f89-6grf8   1/1     Running   0          60s
        cert-manager-webhook-7d4b5b8c56-7wrkf      1/1     Running   0          60s

      Create an Issuer to test the webhook works okay:

      .. code:: bash

        cat <<EOF > test-resources.yaml
        apiVersion: v1
        kind: Namespace
        metadata:
          name: cert-manager-test
        ---
        apiVersion: cert-manager.io/v1
        kind: Issuer
        metadata:
          name: test-selfsigned
          namespace: cert-manager-test
        spec:
          selfSigned: {}
        ---
        apiVersion: cert-manager.io/v1
        kind: Certificate
        metadata:
          name: selfsigned-cert
          namespace: cert-manager-test
        spec:
          dnsNames:
            - example.com
          secretName: selfsigned-cert-tls
          issuerRef:
            name: test-selfsigned
        EOF

      Create the test resources:

      .. code:: bash

        kubectl apply -f test-resources.yaml

      Check the status of the newly created certificate. You may need to wait a few seconds before the cert-manager processes the certificate request.

      .. code:: bash

        kubectl describe certificate -n cert-manager-test

      The output should look like this:

      .. code:: bash

        ...
        Status:
          Conditions:
            Last Transition Time:  2022-08-12T09:11:03Z
            Message:               Certificate is up to date and has not expired
            Observed Generation:   1
            Reason:                Ready
            Status:                True
            Type:                  Ready
          Not After:               2022-11-10T09:11:03Z
          Not Before:              2022-08-12T09:11:03Z
          Renewal Time:            2022-10-11T09:11:03Z
          Revision:                1
        Events:
          Type    Reason     Age   From                                       Message
          ----    ------     ----  ----                                       -------
          Normal  Issuing    7s    cert-manager-certificates-trigger          Issuing certificate as Secret does not exist
          Normal  Generated  6s    cert-manager-certificates-key-manager      Stored new private key in temporary Secret resource "selfsigned-cert-j4jwn"
          Normal  Requested  6s    cert-manager-certificates-request-manager  Created new CertificateRequest resource "selfsigned-cert-gw8b9"
          Normal  Issuing    6s    cert-manager-certificates-issuing          The certificate has been successfully issued

      Clean up the test resources:

      .. code:: bash

        kubectl delete -f test-resources.yaml

      If all the above steps have been completed without error, you're good to go!

.. _use-aws-ecr-with-iam-role:

3. Prepare Container Registry
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. tab-set::

    .. tab-item:: Use Existing Container Registry

        `docker.io <https://docs.docker.com/engine/reference/commandline/login/>`_, `GCR <https://cloud.google.com/container-registry/docs/advanced-authentication#json-key>`_, `ECR <https://docs.aws.amazon.com/AmazonECR/latest/userguide/registry_auth.html#registry-auth-token>`_, `GHCR <https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry>`_, `quay.io <https://docs.quay.io/guides/login.html>`_ are all standard container registries, just get their connection parameters and set them to the following environment variables:

        .. note::

          Since the ECR password will expire regularly, you need to retrieve the ECR password regularly, see this article for details: `Kubernetes - pull an image from private ECR registry. Auto refresh ECR token. <https://skryvets.com/blog/2021/03/15/kubernetes-pull-image-from-private-ecr-registry/>`_

        .. code:: bash

          export DOCKER_REGISTRY_SERVER=xxx
          export DOCKER_REGISTRY_USERNAME=xxx
          export DOCKER_REGISTRY_PASSWORD=xxx
          export DOCKER_REGISTRY_SECURE=false
          export DOCKER_REGISTRY_BENTO_REPOSITORY_NAME=yatai-bentos

    .. tab-item:: Use AWS ECR with IAM Role

        1. Make sure you have an AWS account and have installed `aws-cli <https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html>`_.

        2. Make sure you node has an IAM role with the following policies:

        .. code::

           - AmazonEC2ContainerRegistryReadOnly

        3. Create an ECR repository

        .. code:: bash

          aws ecr create-repository --repository-name yatai-bentos --region YOUR-REGION

        .. note::

          Replace :code:`YOUR-REGION` with your AWS region. For example, if you are in the US East (N. Virginia) region, use :code:`us-east-1`.

          Store the :code:`repositoryArn` returned by the command for later use.

        4. Create an IAM policy for ECR push access for the bento image builder pod of yatai-image-builder

        Create a file named :code:`yatai-image-builder-pod-ecr-policy.json` with the following content:

        .. code:: json

          {
              "Version": "2012-10-17",
              "Statement": [
                  {
                      "Sid": "VisualEditor0",
                      "Effect": "Allow",
                      "Action": [
                          "ecr:PutImageTagMutability",
                          "ecr:StartImageScan",
                          "ecr:DescribeImageReplicationStatus",
                          "ecr:ListTagsForResource",
                          "ecr:UploadLayerPart",
                          "ecr:BatchDeleteImage",
                          "ecr:ListImages",
                          "ecr:BatchGetRepositoryScanningConfiguration",
                          "ecr:DeleteRepository",
                          "ecr:CompleteLayerUpload",
                          "ecr:TagResource",
                          "ecr:DescribeRepositories",
                          "ecr:BatchCheckLayerAvailability",
                          "ecr:ReplicateImage",
                          "ecr:GetLifecyclePolicy",
                          "ecr:PutLifecyclePolicy",
                          "ecr:DescribeImageScanFindings",
                          "ecr:GetLifecyclePolicyPreview",
                          "ecr:PutImageScanningConfiguration",
                          "ecr:GetDownloadUrlForLayer",
                          "ecr:DeleteLifecyclePolicy",
                          "ecr:PutImage",
                          "ecr:UntagResource",
                          "ecr:BatchGetImage",
                          "ecr:DescribeImages",
                          "ecr:StartLifecyclePolicyPreview",
                          "ecr:InitiateLayerUpload",
                          "ecr:GetRepositoryPolicy"
                      ],
                      "Resource": "YOUR-ECR-REPOSITORY-ARN"
                  },
                  {
                      "Sid": "VisualEditor1",
                      "Effect": "Allow",
                      "Action": [
                          "ecr:GetRegistryPolicy",
                          "ecr:BatchImportUpstreamImage",
                          "ecr:CreateRepository",
                          "ecr:DescribeRegistry",
                          "ecr:DescribePullThroughCacheRules",
                          "ecr:GetAuthorizationToken",
                          "ecr:PutRegistryScanningConfiguration",
                          "ecr:CreatePullThroughCacheRule",
                          "ecr:DeletePullThroughCacheRule",
                          "ecr:GetRegistryScanningConfiguration",
                          "ecr:PutReplicationConfiguration"
                      ],
                      "Resource": "*"
                  }
              ]
          }

        .. note::

          Replace :code:`YOUR-ECR-REPOSITORY-ARN` with the :code:`repositoryArn` you stored in the previous step.

        Create the IAM policy with the following command:

        .. code:: bash

          aws iam create-policy --policy-name yatai-image-builder-pod-ecr-policy --policy-document file://yatai-image-builder-pod-ecr-policy.json

        .. note::

          Store the :code:`Arn` returned by the command for later use. The ``Arn`` format is like this: :code:`arn:aws:iam::123456789012:policy/yatai-image-builder-pod-ecr-policy`

        5. Create an IAM role for the service account

        .. code:: bash

          eksctl create iamserviceaccount \
            --cluster=YOUR-CLUSTER \
            --region=YOUR-REGION \
            --namespace=yatai \
            --name=yatai-image-builder-pod \
            --attach-policy-arn=YOUR-IAM-POLICY-ARN \
            --override-existing-serviceaccounts \
            --approve

        .. note::

          Replace ``YOUR-CLUSTER`` with your EKS cluster name, ``YOUR-REGION`` with your AWS region, and ``YOUR-IAM-POLICY-ARN`` with the :code:`Arn` you stored in the previous step.
          If the namespace `yatai` is not exists, create it first.

        6. Add kubernetes label to the service account

        .. code:: bash

          kubectl label serviceaccount -n yatai yatai-image-builder-pod yatai.ai/yatai-image-builder-pod=true

        .. note:: this label is used to identify the service account for the image builder pod creation.

        7. Set the environment variables

        .. code:: bash

          export DOCKER_REGISTRY_SERVER=YOUR-ECR-REGISTRY-URL
          export DOCKER_REGISTRY_USERNAME=""
          export DOCKER_REGISTRY_PASSWORD=""
          export DOCKER_REGISTRY_SECURE=true
          export DOCKER_REGISTRY_BENTO_REPOSITORY_NAME=yatai-bentos

        .. note::

          Replace ``YOUR-ECR-REGISTRY-URL`` with your ECR registry URL. The URL format is like this: :code:`123456789012.dkr.ecr.us-east-1.amazonaws.com`

    .. tab-item:: Install Private Container Registry

        .. note:: Do not recommend for production because this installation does not guarantee high availability.

        1. Install the docker-registry helm chart

        .. code:: bash

          helm upgrade --install docker-registry docker-registry \
              --repo https://helm.twun.io \
              -n yatai-image-builder

        2. Verify the docker-registry installation

        .. code:: bash

          kubectl -n yatai-image-builder get pod -l app=docker-registry

        The output should look like this:

        .. note:: Wait until the status of all pods becomes :code:`Running` before proceeding.

        .. code:: bash

          NAME                               READY   STATUS    RESTARTS   AGE
          docker-registry-7dc8b575d4-d6stx   1/1     Running   0          10m

        3. Create a docker private registry proxy for development and testing purposes

        For **development** and **testing** purposes, sometimes it's useful to build images locally and push them directly to a Kubernetes cluster.

        This can be achieved by running a docker registry in the cluster and using a special repo prefix such as :code:`127.0.0.1:5000/` that will be seen as an insecure registry url.

        .. code:: bash

          cat <<EOF | kubectl apply -f -
          apiVersion: apps/v1
          kind: DaemonSet
          metadata:
            name: docker-private-registry-proxy
            namespace: yatai-image-builder
            labels:
              app: docker-private-registry-proxy
          spec:
            selector:
              matchLabels:
                app: docker-private-registry-proxy
            template:
              metadata:
                creationTimestamp: null
                labels:
                  app: docker-private-registry-proxy
              spec:
                containers:
                - args:
                  - tcp
                  - "5000"
                  - docker-registry.yatai-image-builder.svc.cluster.local
                  image: quay.io/bentoml/proxy-to-service:v2
                  name: tcp-proxy
                  ports:
                  - containerPort: 5000
                    hostPort: 5000
                    name: tcp
                    protocol: TCP
                  resources:
                    limits:
                      cpu: 100m
                      memory: 100Mi
          EOF

        4. Verify the docker-private-registry-proxy installation

        .. code:: bash

          kubectl -n yatai-image-builder get pod -l app=docker-private-registry-proxy

        The output should look like this:

        .. note:: Wait until the status of all pods becomes :code:`Running` before proceeding. The number of pods depends on how many nodes you have.

        .. code:: bash

          NAME                                  READY   STATUS    RESTARTS   AGE
          docker-private-registry-proxy-jzjxr   1/1     Running   0          74s

        5. Prepare the docker registry connection params

        .. code:: bash

          export DOCKER_REGISTRY_SERVER=127.0.0.1:5000
          export DOCKER_REGISTRY_IN_CLUSTER_SERVER=docker-registry.yatai-image-builder.svc.cluster.local:5000
          export DOCKER_REGISTRY_USERNAME=''
          export DOCKER_REGISTRY_PASSWORD=''
          export DOCKER_REGISTRY_SECURE=false
          export DOCKER_REGISTRY_BENTO_REPOSITORY_NAME=yatai-bentos

4. Install yatai-image-builder
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

1. Install yatai-image-builder-crds
"""""""""""""""""""""""""""""""""""

.. code:: bash

  helm upgrade --install yatai-image-builder-crds yatai-image-builder-crds \
      --repo https://bentoml.github.io/helm-charts \
      -n yatai-image-builder

.. warning::

   If you encounter error like this:

   .. code:: bash

      Error: rendered manifests contain a resource that already exists. Unable to continue with install: CustomResourceDefinition "bentodeployments.serving.yatai.ai" in namespace "" exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to "yatai-image-builder-crds"; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to "yatai-image-builder"

   It means you already have BentoDeployment CRD, you should use this command to fix it:

   .. code:: bash

      kubectl label crd bentorequests.resources.yatai.ai app.kubernetes.io/managed-by=Helm
      kubectl annotate crd bentorequests.resources.yatai.ai meta.helm.sh/release-name=yatai-image-builder-crds meta.helm.sh/release-namespace=yatai-image-builder
      kubectl label crd bentoes.resources.yatai.ai app.kubernetes.io/managed-by=Helm
      kubectl annotate crd bentoes.resources.yatai.ai meta.helm.sh/release-name=yatai-image-builder-crds meta.helm.sh/release-namespace=yatai-image-builder

   Then reinstall the ``yatai-image-builder-crds``.

2. Verify that the CRDs of yatai-image-builder has been established
"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl wait --for condition=established --timeout=120s crd/bentorequests.resources.yatai.ai
  kubectl wait --for condition=established --timeout=120s crd/bentoes.resources.yatai.ai

The output of the command above should look something like this:

.. code:: bash

  customresourcedefinition.apiextensions.k8s.io/bentorequests.resources.yatai.ai condition met
  customresourcedefinition.apiextensions.k8s.io/bentoes.resources.yatai.ai condition met

3. Install the yatai-image-builder helm chart
"""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  helm upgrade --install yatai-image-builder yatai-image-builder \
      --repo https://bentoml.github.io/helm-charts \
      -n yatai-image-builder \
      --set dockerRegistry.server=$DOCKER_REGISTRY_SERVER \
      --set dockerRegistry.inClusterServer=$DOCKER_REGISTRY_IN_CLUSTER_SERVER \
      --set dockerRegistry.username=$DOCKER_REGISTRY_USERNAME \
      --set dockerRegistry.password=$DOCKER_REGISTRY_PASSWORD \
      --set dockerRegistry.secure=$DOCKER_REGISTRY_SECURE \
      --set dockerRegistry.bentoRepositoryName=$DOCKER_REGISTRY_BENTO_REPOSITORY_NAME

.. note::

   If you are using :ref:`AWS ECR with IAM Role <use-aws-ecr-with-iam-role>`, you need to add the following option to the helm install command:

   .. code:: bash

      --set dockerRegistry.useAWSECRWithIAMRole=true \
      --set dockerRegistry.awsECRRegion=YOUR-REGION

   Replace ``YOUR-REGION`` with your AWS region.

4. Verify the yatai-image-builder installation
""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl -n yatai-image-builder get pod -l app.kubernetes.io/name=yatai-image-builder

The output should look like this:

.. note:: Wait until the status of all pods becomes :code:`Running` or :code:`Completed` before proceeding.

.. code:: bash

  NAME                                    READY   STATUS      RESTARTS   AGE
  yatai-image-builder-8b9fb98d7-xmtd5        1/1     Running     0          67s

View the logs of :code:`yatai-image-builder`:

.. code:: bash

  kubectl -n yatai-image-builder logs -f deploy/yatai-image-builder
