=========
Migration
=========

Since v1.0.0 has many breaking changes, we need to do some preparation to migrate your existing v0.4.x version of yatai.

.. note:: If your yatai is below v0.4.6, you must upgrade it to v0.4.6 first.

Breaking changes
----------------

* Split Yatai into two components for better modularization and separation of concerns.

  * yatai

    Dashboard and bento registry

  * yatai-deployment

    Deploy bento to Kubernetes

* Removed all Yatai component operators for more standard integration with the ecosystem.

  In v0.4.x, yatai component operators will automatically install dependencies of the component. But in v1.0.0, we removed this feature.

  Some things will changes:

  * No more integration of logging and monitoring

    Because we removed all yatai component operators, yatai now does not automatically integrate logging and monitoring. See the :ref:`Observability <observability>` documentation for observability configuration.

Down time during migration
--------------------------

* Yatai system

  Depending on how quickly you reinstall yatai and recover data

* BentoDeployment

  Won't go offline

Migration steps
---------------

1. Backup PostgreSQL data
"""""""""""""""""""""""""

.. note:: If you use the external PostgreSQL, you need to skip this step.

.. code:: bash

  pkill kubectl
  kubectl port-forward --namespace yatai-system svc/yatai-postgresql 5433:5432 &
  sleep 6
  PGPASSWORD=$(kubectl -n yatai-system get secret yatai-postgresql -o jsonpath='{.data.postgresql-password}' | base64 -d) pg_dump -h localhost -p 5433 -U postgres -F t yatai > /tmp/yatai.tar

2. Get object store environment variables
"""""""""""""""""""""""""""""""""""""""""

.. note:: If you use the external S3, you need to skip this step.

.. code:: bash

  export S3_ENDPOINT=minio.yatai-components.svc.cluster.local
  export S3_ACCESS_KEY=$(kubectl -n yatai-components get secret yatai-minio-secret -o jsonpath='{.data.accesskey}' | base64 -d)
  export S3_SECRET_KEY=$(kubectl -n yatai-components get secret yatai-minio-secret -o jsonpath='{.data.secretkey}' | base64 -d)
  export S3_SECURE=false
  export S3_BUCKET_NAME=yatai
  export S3_REGION=i-dont-known

3. Test object store connection
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

4. Uninstall yatai and yatai component operators
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

6. Get docker registry environment variables
""""""""""""""""""""""""""""""""""""""""""""

.. note:: If you use the external docker registry, you need to skip this step.

.. code:: bash

  export DOCKER_REGISTRY_SERVER=127.0.0.1:5000
  export DOCKER_REGISTRY_IN_CLUSTER_SERVER=yatai-docker-registry.yatai-components.svc.cluster.local:5000
  export DOCKER_REGISTRY_USERNAME=''
  export DOCKER_REGISTRY_PASSWORD=''
  export DOCKER_REGISTRY_SECURE=false
  export DOCKER_REGISTRY_BENTO_REPOSITORY_NAME=bentos

7. Install yatai-deployment
"""""""""""""""""""""""""""

Read this documentation to install yatai-deployment: :ref:`Installing yatai-deployment <yatai-deployment-installation-steps>`

.. note:: You need to skip the installation of docker-registry.

