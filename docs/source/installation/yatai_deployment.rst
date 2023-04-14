===========================
Installing yatai-deployment
===========================

Welcome to yatai-deployment! You will learn the system requirements, software dependencies, instructions for installing this Yatai component.

See :ref:`yatai-deployment architecture <concepts/architecture:yatai-deployment>` for a detailed introduction of the ``yatai-deployment`` component.

Prerequisites
-------------

- yatai-image-builder

  ``yatai-deployment`` relies on Bento CR to get the image and runners information, you should check the documentation :doc:`yatai_image_builder` first.

- Kubernetes

  Kubernetes cluster with version from 1.20 to 1.25.

  .. note::

    If you do not have a production Kubernetes cluster and want to install ``yatai-deployment`` for development and testing purposes. You can use `minikube <https://minikube.sigs.k8s.io/docs/start/>`_ to set up a local Kubernetes cluster for testing. If you are using macOS, you should use `hyperkit <https://minikube.sigs.k8s.io/docs/drivers/hyperkit/>`_ driver to prevent the macOS docker desktop `network limitation <https://docs.docker.com/desktop/networking/#i-cannot-ping-my-containers>`_

- Helm

  Yatai uses `Helm <https://helm.sh/docs/intro/using_helm/>`_ to install ``yatai-deployment``.

- Ingress Controller

  Yatai uses ingress controller to facilitate access to bento deployments.
  You can use the following command to check if you have an ingress controller installed in your cluster:

  .. code:: bash

    kubectl get ingressclass

  The output should look like this:

  .. code:: bash

    NAME    CONTROLLER             PARAMETERS   AGE
    nginx   k8s.io/ingress-nginx   <none>       10d

  If no value is returned, there is no ingress controller installed in your cluster. You need to select an ingress controller and install it, for example, you can install `nginx-ingress <https://kubernetes.github.io/ingress-nginx/deploy/#quick-start>`_.
  If you are using :code:`minikube`, you don't need to install ingress controller manually, just enable :code:`ingress addon` with the following command:

  .. code:: bash

    minikube addons enable ingress

Quick Install
-------------

.. note:: This quick installation script can only be used for **development** and **testing** purposes.

This script will automatically install the following dependencies inside the :code:`yatai-deployment` namespace of the Kubernetes cluster:

* cert-manager (if not already installed)
* metrics-server (if not already installed)

.. code:: bash

  bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai-deployment/main/scripts/quick-install-yatai-deployment.sh")

.. _yatai-deployment-installation-steps:

Installation Steps
------------------

.. note::

  If you don't have :code:`kubectl` installed and you are using :code:`minikube`, you can use :code:`minikube kubectl --` instead of :code:`kubectl`, for more details on using it, please check: `minikube kubectl <https://minikube.sigs.k8s.io/docs/commands/kubectl/>`_

1. Create Namespaces
^^^^^^^^^^^^^^^^^^^^

.. code:: bash

  # for yatai-deployment deployment
  kubectl create ns yatai-deployment
  # for bento deployment resources
  kubectl create ns yatai

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

3. Install Metrics Server
^^^^^^^^^^^^^^^^^^^^^^^^^

Read its official documentation for `installation <https://github.com/kubernetes-sigs/metrics-server#installation>`_

.. note::

   If you are using :code:`minikube`, you can install metrics-server with the following command:

   .. code:: bash

    minikube addons enable metrics-server

4. Configure network
^^^^^^^^^^^^^^^^^^^^

The network config is for :code:`BentoDeployment` access.

1. Ingress Class
""""""""""""""""

Set `ingress class <https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class>`_ for :code:`BentoDeployment` ingress.

Store your ingress class in environment var:

.. code:: bash

  export INGRESS_CLASS=$(kubectl get ingressclass -o jsonpath='{.items[0].metadata.name}' 2> /dev/null)
  echo $INGRESS_CLASS

.. note:: If no value returned, it means you do not have any ingress class, please install a ingress controller first!

**After the yatai-deployment helm chart has been installed** you can configure it in this way:

.. code:: bash

  kubectl -n yatai-deployment patch cm/network --type merge --patch '{"data":{"ingress-class":"'${INGRESS_CLASS}'"}}'

Verify that this ingress class is working properly
**************************************************

.. note::

   You should make sure that the :code:`$INGRESS_CLASS` environment variable is not empty and contains the correct value, otherwise the following command will not work.

.. code:: bash

  cat <<EOF | kubectl apply -f -
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: test-ingress
  spec:
    ingressClassName: ${INGRESS_CLASS}
    rules:
    - http:
        paths:
        - path: /testpath
          pathType: Prefix
          backend:
            service:
              name: test
              port:
                number: 80
  EOF

Wait for ingress to be successfully assigned address:

.. note:: The following command will wait 5 minutes for the above ingress to be assigned address

.. code:: bash

  timeout 5m bash -c "until kubectl get ing test-ingress -o yaml -o jsonpath='{.status.loadBalancer}' | grep ingress; do : ; done" && echo 'successfully' || echo 'failed'

If the above command returns :code:`successfully`, it means that the ingress class is working properly. Otherwise, you need to check the ingress controller logs to see what went wrong.

2. Ingress Annotations
""""""""""""""""""""""

Set annotations for :code:`BentoDeployment` ingress resource

For example, if you want to set ingress annotation: `"foo": "bar"`, you should add the follow option after the `helm install` command:

.. code:: bash

  --set layers.network.ingressAnnotations.foo=bar

After the ``yatai-deployment`` helm chart has been installed you can configure it in this way:

.. code:: bash

    kubectl -n yatai-deployment patch cm/network --type merge --patch '{"data": {"ingress-annotations": "{\"foo\":\"bar\"}"}}'

3. DNS for domain suffix
""""""""""""""""""""""""

The domain suffix is used to generate ingress hosts for :code:`BentoDeployment`.

You need to configure your DNS in one of the following two options:

  .. tab-set::

      .. tab-item:: Magic DNS(sslip.io)

        You don't need to do anything because Yatai will use `sslip.io <https://sslip.io/>`_ to automatically generate :code:`domain-suffix` for :code:`BentoDeployment` ingress host.

      .. tab-item:: Real DNS

        First, you must register a domain name. The following example assumes that you already have a domain name of :code:`example.com`

        To configure DNS for Yatai, take the External IP or CNAME from setting up networking, and configure it with your domain **DNS provider** as follows:

        * If the kubernetes networking layer (LoadBalancer) produced an External IP address, then configure a wildcard A record for the domain:

        .. code:: bash

          # Here yatai.example.com is the domain suffix for your cluster
          *.yatai.example.com == A 35.233.41.212

        * If the networking layer produced a CNAME, then configure a CNAME record for the domain:

        .. code:: bash

          # Here yatai.example.com is the domain suffix for your cluster
          *.yatai.example.com == CNAME a317a278525d111e89f272a164fd35fb-1510370581.eu-central-1.elb.amazonaws.com

        Once your DNS provider has been configured, direct yatai to use that domain:

        .. code:: bash

          export DOMAIN_SUFFIX=yatai.example.com

        After the ``yatai-deployment`` helm chart has been installed you can configure it in this way:

        .. code:: bash

          # Replace yatai.example.com with your domain suffix
          kubectl -n yatai-deployment patch cm/network --type merge --patch '{"data":{"domain-suffix":"'${DOMAIN_SUFFIX}'"}}'

5. Install yatai-deployment
^^^^^^^^^^^^^^^^^^^^^^^^^^^

1. Install yatai-deployment-crds
""""""""""""""""""""""""""""""""

.. code:: bash

  helm upgrade --install yatai-deployment-crds yatai-deployment-crds \
      --repo https://bentoml.github.io/helm-charts \
      -n yatai-deployment

.. warning::

   If you encounter error like this:

   .. code:: bash

      Error: rendered manifests contain a resource that already exists. Unable to continue with install: CustomResourceDefinition "bentodeployments.serving.yatai.ai" in namespace "" exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to "yatai-deployment-crds"; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to "yatai-deployment"

   It means you already have BentoDeployment CRD, you should use this command to fix it:

   .. code:: bash

      kubectl label crd bentodeployments.serving.yatai.ai app.kubernetes.io/managed-by=Helm
      kubectl annotate crd bentodeployments.serving.yatai.ai meta.helm.sh/release-name=yatai-deployment-crds meta.helm.sh/release-namespace=yatai-deployment

   Then reinstall the ``yatai-deployment-crds``.


2. Verify that the CRDs of yatai-deployment has been established
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl wait --for condition=established --timeout=120s crd/bentodeployments.serving.yatai.ai

The output of the command above should look something like this:

.. code:: bash

  customresourcedefinition.apiextensions.k8s.io/bentodeployments.serving.yatai.ai condition met

3. Install the yatai-deployment helm chart
""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  helm upgrade --install yatai-deployment yatai-deployment \
      --repo https://bentoml.github.io/helm-charts \
      -n yatai-deployment \
      --set dockerRegistry.server=$DOCKER_REGISTRY_SERVER \
      --set dockerRegistry.inClusterServer=$DOCKER_REGISTRY_IN_CLUSTER_SERVER \
      --set dockerRegistry.username=$DOCKER_REGISTRY_USERNAME \
      --set dockerRegistry.password=$DOCKER_REGISTRY_PASSWORD \
      --set dockerRegistry.secure=$DOCKER_REGISTRY_SECURE \
      --set dockerRegistry.bentoRepositoryName=$DOCKER_REGISTRY_BENTO_REPOSITORY_NAME \
      --set layers.network.ingressClass=$INGRESS_CLASS

4. Verify the yatai-deployment installation
"""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl -n yatai-deployment get pod -l app.kubernetes.io/name=yatai-deployment

The output should look like this:

.. note:: Wait until the status of all pods becomes :code:`Running` or :code:`Completed` before proceeding.

.. code:: bash

  NAME                                    READY   STATUS      RESTARTS   AGE
  yatai-deployment-8b9fb98d7-xmtd5        1/1     Running     0          67s
  yatai-deployment-default-domain-s8rh9   0/1     Completed   0          67s

View the logs of :code:`yatai-deployment-default-domain`:

.. code:: bash

  kubectl -n yatai-deployment logs -f job/yatai-deployment-default-domain

The logs of :code:`yatai-deployment-default-domain` should be like this:

.. note:: Automatic domain-suffix generation will take about 1 minute.

.. code:: bash

  time="2022-08-16T14:48:11Z" level=info msg="Creating ingress default-domain- to get a ingress IP automatically"
  time="2022-08-16T14:48:11Z" level=info msg="Waiting for ingress default-domain-rrlb9 to be ready"
  time="2022-08-16T14:48:41Z" level=info msg="Ingress default-domain-rrlb9 is ready"
  time="2022-08-16T14:48:41Z" level=info msg="you have not set the domain-suffix in the network config, so use magic DNS to generate a domain suffix automatically: `10.0.0.116.sslip.io`, and set it to the network config"

View the logs of :code:`yatai-deployment`:

.. code:: bash

  kubectl -n yatai-deployment logs -f deploy/yatai-deployment
