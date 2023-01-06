===========================
Upgrade yatai-image-builder
===========================

Prerequisites
-------------

- Helm

  yatai-image-builder uses `Helm <https://helm.sh/docs/intro/using_helm/>`_ to install/upgrade yatai-image-builder.

Upgrade Steps
-------------

1. Check yatai-image-builder-crds current version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

  helm list -f "^yatai-image-builder-crds$" -A

You should see something like this:

.. code-block:: bash

  NAME                       NAMESPACE                  REVISION        UPDATED                                   STATUS          CHART                                  APP VERSION
  yatai-image-builder-crds   yatai-image-builder        1               2023-01-03 13:03:02.783856038 +0000 UTC   deployed        yatai-image-builder-crds-1.1.0-d12     1.1.0-d12

As you can see, the current version is ``1.1.0-d12``.

2. Upgrade yatai-image-builder-crds to the target version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

If you want to upgrade to ``1.1.0-d13``, you can run the following command:

.. warning::

  If the minor version of the target version is different from the current version, you need to skip this step and follow the migration guide to complete this upgrade.

.. note::

   If your release name is not ``yatai-image-builder-crds``, you need to replace ``yatai-image-builder-crds`` with your release name in the following command.
   If your namespace is not ``yatai-image-builder``, you need to replace ``yatai-image-builder`` with your namespace in the following command.

.. code-block:: bash

   helm upgrade yatai-image-builder-crds yatai-image-builder-crds \
       --repo https://bentoml.github.io/helm-charts \
       --version 1.1.0-d13 \
       --namespace yatai-image-builder

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

   Then upgrade the ``yatai-image-builder-crds`` again.

3. Check yatai-image-builder current version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

  helm list -f "^yatai-image-builder$" -A

You should see something like this:

.. code-block:: bash

  NAME                    NAMESPACE               REVISION        UPDATED                                 STATUS          CHART                           APP VERSION
  yatai-image-builder     yatai-image-builder     1               2022-12-23 09:43:58.003704605 +0000 UTC deployed        yatai-image-builder-1.1.0-d12   1.1.0-d12

As you can see, the current version is ``1.1.0-d12``.

4. Upgrade yatai-image-builder to the target version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

If you want to upgrade to ``1.1.0-d13``, you can run the following command:

.. warning::

  If the minor version of the target version is different from the current version, you need to skip this step and follow the migration guide to complete this upgrade.

.. note::

   If your release name is not ``yatai-image-builder``, you need to replace ``yatai-image-builder`` with your release name in the following command.
   If your namespace is not ``yatai-image-builder``, you need to replace ``yatai-image-builder`` with your namespace in the following command.

.. code-block:: bash

   helm upgrade yatai-image-builder yatai-image-builder \
       --repo https://bentoml.github.io/helm-charts \
       --version 1.1.0-d13 \
       --namespace yatai-image-builder

