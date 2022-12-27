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

1. Uninstall yatai-deployment
""""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  helm -n yatai-deployment uninstall yatai-deployment

2. Get Docker Registry Environment Variables
""""""""""""""""""""""""""""""""""""""""""""

.. note:: If you use the external docker registry, you need to skip this step.

.. code:: bash

  export DOCKER_REGISTRY_SERVER=127.0.0.1:5000
  export DOCKER_REGISTRY_IN_CLUSTER_SERVER=docker-registry.yatai-deployment.svc.cluster.local:5000
  export DOCKER_REGISTRY_USERNAME=''
  export DOCKER_REGISTRY_PASSWORD=''
  export DOCKER_REGISTRY_SECURE=false
  export DOCKER_REGISTRY_BENTO_REPOSITORY_NAME=bentos

3. Install yatai-image-builder
""""""""""""""""""""""""""""""

Read this documentation to install yatai-image-builder: :ref:`Installing yatai-image-builder <yatai-image-builder-installation-steps>`

.. note:: You should skip the step of Docker Registry installation because it has already been done as a part of the migration.

4. Install yatai-deployment
"""""""""""""""""""""""""""

Read this documentation to install yatai-deployment: :ref:`Installing yatai-deployment <yatai-deployment-installation-steps>`
