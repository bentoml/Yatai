===========================
Upgrade yatai-image-builder
===========================

Prerequisites
-------------

- Helm

  yatai-image-builder uses `Helm <https://helm.sh/docs/intro/using_helm/>`_ to install/upgrade yatai-image-builder.

Upgrade Steps
-------------

1. Check current version
^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

  helm list -f "^yatai-image-builder$" -A

You should see something like this:

.. code-block:: bash

  NAME                    NAMESPACE               REVISION        UPDATED                                 STATUS          CHART                           APP VERSION
  yatai-image-builder     yatai-image-builder     1               2022-12-23 09:43:58.003704605 +0000 UTC deployed        yatai-image-builder-1.1.0-d12   1.1.0-d12

As you can see, the current version is ``1.1.0-d12``.

2. Upgrade to the target version
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

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

