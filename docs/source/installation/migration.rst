=========
Migration
=========

Starting from version 1.0 of BentoML, components of Yatai are installed separately for more standard integration with the Kubernetes ecosystem. Components such as deployment, logging, and monitoring features are now add-ons that you must install separately, see the :ref:`observability <observability>` section. The new installation method allows you to choose which features of Yatai you would like, and have an easier time managing them as separate components.

.. note:: Migrating data to Yatai 1.0 requires you to have Yatai version 0.46.

Breaking Changes
----------------

* Split Yatai into two components for better modularization and separation of concerns.

  * :ref:`Yatai <concepts/architecture:Yatai>`
  * :ref:`Yatai Deployment <concepts/architecture:Yatai Deployment>`

* Removed all Yatai component operators for more standard integration with the ecosystem.

  In v0.4.x, yatai component operators will automatically install dependencies of the component. But in v1.0.0, we removed this feature.

  Some things will changes:

  * No more integration of logging and monitoring

    Because we removed all yatai component operators, yatai now does not automatically integrate logging and monitoring. See the :ref:`Observability <observability>` documentation for observability configuration.

Down Time and Data Backup
-------------------------

Your data and model files will not be affected if they are stored in a stable, external platform. If your storage is on the same cluster as Yatai, you must back up and recover the data manually. This document will walk you through on how to back up your data. BentoML deployments will be completely unaffected and remain online.

* Yatai System

  Yatai will be down during migration until you reinstall and recover the data.

* Bento Deployment CRD

  The ``BentoDeployment`` CRD will stay online.

Migration steps
---------------

1. Backup Your Data
"""""""""""""""""""

This step guides you on backing up your database and object storage data. If you stored your data in an external relational database and/or object storage, you do not have to backup PostgreSQL data and/or export object storage environment variables.

.. note:: Back up PostgreSQL data with the following commands. This step must be skipped if you stored your data in an external relational database.

.. code:: bash

  pkill kubectl
  kubectl port-forward --namespace yatai-system svc/yatai-postgresql 5433:5432 &
  sleep 6
  PGPASSWORD=$(kubectl -n yatai-system get secret yatai-postgresql -o jsonpath='{.data.postgresql-password}' | base64 -d) pg_dump -h localhost -p 5433 -U postgres -F t yatai > /tmp/yatai.tar

2. Get Object Store Environment Variables
"""""""""""""""""""""""""""""""""""""""""

.. note:: Get object storage environment variables with the following commands. This step can be skipped if you are using an external S3 bucket.

.. code:: bash

  export S3_ENDPOINT=minio.yatai-components.svc.cluster.local
  export S3_ACCESS_KEY=$(kubectl -n yatai-components get secret yatai-minio-secret -o jsonpath='{.data.accesskey}' | base64 -d)
  export S3_SECRET_KEY=$(kubectl -n yatai-components get secret yatai-minio-secret -o jsonpath='{.data.secretkey}' | base64 -d)
  export S3_SECURE=false
  export S3_BUCKET_NAME=yatai
  export S3_REGION=i-dont-known

3. Test Object Store Connection
"""""""""""""""""""""""""""""""

.. code:: bash

  kubectl -n yatai-system delete pod s3-client 2> /dev/null || true; \
  kubectl run s3-client --rm --tty -i --restart='Never' \
      --namespace yatai-system \
      --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
      --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
      --image quay.io/bentoml/s3-client:0.0.1 \
      --command -- sh -c "s3-client -e http://$S3_ENDPOINT listbuckets && echo successfully"

The output should be:

.. code:: bash

  successfully
  pod "s3-client" deleted

4. Uninstall Yatai and Yatai Component Operators
""""""""""""""""""""""""""""""""""""""""""""""""

.. code:: bash

  helm uninstall yatai -n yatai-system
  helm uninstall yatai -n yatai-components
  helm uninstall yatai-csi-driver-image-populator -n yatai-components
  helm list -n yatai-operators | tail -n +2 | awk '{print $1}' | xargs -I{} helm -n yatai-operators uninstall {}

5. Install Yatai
""""""""""""""""

Read this documentation to install Yatai: :ref:`Installing Yatai <yatai-installation-steps>`

.. note::

  You need to skip the installation of MinIO and install a new PostgreSQL as described in the documentation above. After the PostgreSQL installation, you need to run the following command to restore the old data:

  .. code:: bash

    pkill kubectl
    kubectl port-forward --namespace yatai-system svc/postgresql-ha-pgpool 5433:5432 &
    sleep 6
    PGPASSWORD=$(kubectl -n yatai-system get secret postgresql-ha-postgresql -o jsonpath='{.data.postgresql-password}' | base64 -d) pg_restore -h localhost -p 5433 -U postgres -d yatai /tmp/yatai.tar

6. Get Docker Registry Environment Variables
""""""""""""""""""""""""""""""""""""""""""""

.. note:: If you use the external docker registry, you need to skip this step.

.. code:: bash

  export DOCKER_REGISTRY_SERVER=127.0.0.1:5000
  export DOCKER_REGISTRY_IN_CLUSTER_SERVER=yatai-docker-registry.yatai-components.svc.cluster.local:5000
  export DOCKER_REGISTRY_USERNAME=''
  export DOCKER_REGISTRY_PASSWORD=''
  export DOCKER_REGISTRY_SECURE=false
  export DOCKER_REGISTRY_BENTO_REPOSITORY_NAME=bentos

7. Install Yatai Deployment
"""""""""""""""""""""""""""

Read this documentation to install yatai-deployment: :ref:`Installing yatai-deployment <yatai-deployment-installation-steps>`

.. note:: You should skip the step of Docker Registry installation because it has already been done as a part of the migration.
