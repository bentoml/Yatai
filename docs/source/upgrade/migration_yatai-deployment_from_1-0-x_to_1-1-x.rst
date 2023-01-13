==============================================
Migration yatai-deployment from 1.0.x to 1.1.x
==============================================

Yatai 1.1.0 adds ``yatai-image-builder`` component, docker registry dependency migrated from yatai-deployment to yatai-image-builder

.. note:: Migrating yatai-deployment to 1.1.0 requires you to have Yatai version 1.0.x.

Breaking Changes
----------------

* ``yatai-deployment`` remove the dependency of docker registry

* You need to install ``yatai-image-builder`` before installing ``yatai-deployment``

Down Time
---------

Bento deployments will be completely unaffected and remain online.

* Yatai System

  Yatai system will not go offline

* BentoDeployment CR

  The ``BentoDeployment`` CR will stay online.

Migration steps
---------------

1. Store ``yatai-deployment`` helm release values
"""""""""""""""""""""""""""""""""""""""""""""""""

.. code-block:: bash

    helm get values yatai-deployment -n yatai-deployment > /tmp/yatai-deployment-values.yaml

.. note::

  The above command assumes you have installed yatai-deployment with release name ``yatai-deployment`` in ``yatai-deployment`` namespace. If you have installed yatai-deployment with a different release name or namespace, please adjust the command accordingly.

.. note::

  The above command assumes you have installed yatai-deployment with release name ``yatai-deployment`` in ``yatai-deployment`` namespace. If you have installed yatai-deployment with a different release name or namespace, please adjust the command accordingly.

2. Create ``yatai-image-builder`` namespace
"""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl create namespace yatai-image-builder

3. Install ``yatai-image-builder-crds``
"""""""""""""""""""""""""""""""""""""""

.. code:: bash

  helm upgrade --install yatai-image-builder-crds yatai-image-builder-crds \
      --repo https://bentoml.github.io/helm-charts \
      -n yatai-image-builder

.. warning::

   If you encounter error like this:

   .. code:: bash

      Error: rendered manifests contain a resource that already exists. Unable to continue with install: CustomResourceDefinition "bentorequests.resources.yatai.ai" in namespace "" exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to "yatai-image-builder-crds"; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to "yatai-image-builder"

   It means you already have BentoRequest CRD and Bento CRD, you should use this command to fix it:

   .. code:: bash

      kubectl label crd bentorequests.resources.yatai.ai app.kubernetes.io/managed-by=Helm
      kubectl annotate crd bentorequests.resources.yatai.ai meta.helm.sh/release-name=yatai-image-builder-crds meta.helm.sh/release-namespace=yatai-image-builder
      kubectl label crd bentoes.resources.yatai.ai app.kubernetes.io/managed-by=Helm
      kubectl annotate crd bentoes.resources.yatai.ai meta.helm.sh/release-name=yatai-image-builder-crds meta.helm.sh/release-namespace=yatai-image-builder

   Then reinstall the ``yatai-image-builder-crds``.

4. Verify that the CRDs of ``yatai-image-builder`` has been established
"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl wait --for condition=established --timeout=120s crd/bentorequests.resources.yatai.ai
  kubectl wait --for condition=established --timeout=120s crd/bentoes.resources.yatai.ai

The output of the command above should look something like this:

.. code:: bash

  customresourcedefinition.apiextensions.k8s.io/bentorequests.resources.yatai.ai condition met
  customresourcedefinition.apiextensions.k8s.io/bentoes.resources.yatai.ai condition met

5. Install the ``yatai-image-builder`` helm chart
"""""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  helm upgrade --install yatai-image-builder yatai-image-builder \
      --repo https://bentoml.github.io/helm-charts \
      -n yatai-image-builder \
      --values /tmp/yatai-deployment-values.yaml

6. Verify the ``yatai-image-builder`` installation
""""""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl -n yatai-image-builder get pod -l app.kubernetes.io/name=yatai-image-builder

The output should look like this:

.. note:: Wait until the status of all pods becomes :code:`Running` or :code:`Completed` before proceeding.

.. code:: bash

  NAME                                    READY   STATUS      RESTARTS   AGE
  yatai-image-builder-8b9fb98d7-xmtd5     1/1     Running     0          67s

View the logs of :code:`yatai-image-builder`:

.. code:: bash

  kubectl -n yatai-image-builder logs -f deploy/yatai-image-builder

7. Uninstall ``yatai-deployment`` helm release
""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  helm -n yatai-deployment uninstall yatai-deployment

8. Install ``yatai-deployment``
"""""""""""""""""""""""""""""""

Read this documentation to install yatai-deployment: :ref:`Installing yatai-deployment <yatai-deployment-installation-steps>`
