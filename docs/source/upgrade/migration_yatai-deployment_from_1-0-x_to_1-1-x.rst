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

Get the docker registry configuration from the ``yatai-deployment`` helm release:

.. code:: bash

  helm -n yatai-deployment get values yatai-deployment


Set the docker registry configuration as environment variables:

.. code:: bash

  export DOCKER_REGISTRY_SERVER=xxx
  export DOCKER_REGISTRY_IN_CLUSTER_SERVER=yyy
  export DOCKER_REGISTRY_USERNAME=xxx
  export DOCKER_REGISTRY_PASSWORD=yyy
  export DOCKER_REGISTRY_SECURE=false
  export DOCKER_REGISTRY_BENTO_REPOSITORY_NAME=bentos

3. Install yatai-image-builder
""""""""""""""""""""""""""""""

Read this documentation to install yatai-image-builder: :ref:`Installing yatai-image-builder <yatai-image-builder-installation-steps>`

4. Install yatai-deployment
"""""""""""""""""""""""""""

Read this documentation to install yatai-deployment: :ref:`Installing yatai-deployment <yatai-deployment-installation-steps>`
