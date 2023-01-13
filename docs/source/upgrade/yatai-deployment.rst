========================
Upgrade yatai-deployment
========================

Prerequisites
-------------

- Helm

  yatai-deployment uses `Helm <https://helm.sh/docs/intro/using_helm/>`_ to install/upgrade yatai-deployment.

Upgrade Steps
-------------

1. Check yatai-deployment-crds current version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

  helm list -f "^yatai-deployment-crds$" -A

You should see something like this:

.. code-block:: bash

  NAME                    NAMESPACE          REVISION   UPDATED              STATUS      CHART                         APP VERSION
  yatai-deployment-crds   yatai-deployment   1          2023-01-03 13:03:02  deployed    yatai-deployment-crds-1.1.0   1.1.0

As you can see, the current version is ``1.1.0``.

2. Upgrade yatai-deployment-crds to the target version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

If you want to upgrade to ``1.1.1``, you can run the following command:

.. warning::

   If the minor version of the target version is different from the current version, you will need to check if there is a corresponding migration document, and if so you will need to skip this step and all the next steps and follow the migration guide to complete this upgrade.

.. note::

   If your release name is not ``yatai-deployment-crds``, you need to replace ``yatai-deployment-crds`` with your release name in the following command.
   If your namespace is not ``yatai-deployment``, you need to replace ``yatai-deployment`` with your namespace in the following command.

.. code-block:: bash

   helm upgrade yatai-deployment-crds yatai-deployment-crds \
       --repo https://bentoml.github.io/helm-charts \
       --version 1.1.1 \
       --namespace yatai-deployment

.. warning::

   If you encounter error like this:

   .. code:: bash

      Error: rendered manifests contain a resource that already exists. Unable to continue with install: CustomResourceDefinition "bentodeployments.serving.yatai.ai" in namespace "" exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to "yatai-deployment-crds"; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to "yatai-deployment"

   It means you already have BentoDeployment CRD, you should use this command to fix it:

   .. code:: bash

      kubectl label crd bentodeployments.serving.yatai.ai app.kubernetes.io/managed-by=Helm
      kubectl annotate crd bentodeployments.serving.yatai.ai meta.helm.sh/release-name=yatai-deployment-crds meta.helm.sh/release-namespace=yatai-deployment

   Then upgrade the ``yatai-deployment-crds`` again.

3. Check yatai-deployment current version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

  helm list -f "^yatai-deployment$" -A

You should see something like this:

.. code-block:: bash

  NAME                    NAMESPACE            REVISION    UPDATED              STATUS    CHART                    APP VERSION
  yatai-deployment        yatai-deployment     1           2022-12-23 09:46:24  deployed  yatai-deployment-1.1.0   1.1.0

As you can see, the current version is ``1.1.0``.

4. Upgrade yatai-deployment to the target version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

If you want to upgrade to ``1.1.1``, you can run the following command:

.. warning::

   If the minor version of the target version is different from the current version, you will need to check if there is a corresponding migration document, and if so you will need to skip this step and all the next steps and follow the migration guide to complete this upgrade.

.. note::

   If your release name is not ``yatai-deployment``, you need to replace ``yatai-deployment`` with your release name in the following command.
   If your namespace is not ``yatai-deployment``, you need to replace ``yatai-deployment`` with your namespace in the following command.

.. code-block:: bash

   helm upgrade yatai-deployment yatai-deployment \
       --repo https://bentoml.github.io/helm-charts \
       --version 1.1.1 \
       --namespace yatai-deployment

