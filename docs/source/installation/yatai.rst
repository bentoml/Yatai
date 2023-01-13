================
Installing Yatai
================

Welcome to Yatai! You will learn the system requirements, software dependencies, instructions for installation. See :ref:`Yatai architecture <concepts/architecture:Yatai>` for a detailed introduction of the ``yatai`` component.

Prerequisites
-------------

- Kubernetes

  Kubernetes cluster with version 1.20 or newer

  .. note::

      If you do not have a production Kubernetes cluster and want to install yatai for development and testing purposes. You can use `minikube <https://minikube.sigs.k8s.io/docs/start/>`_ to set up a local Kubernetes cluster for testing. If you are using macOS, you should use `hyperkit <https://minikube.sigs.k8s.io/docs/drivers/hyperkit/>`_ driver to prevent the macOS docker desktop `network limitation <https://docs.docker.com/desktop/networking/#i-cannot-ping-my-containers>`_

- Helm

  Yatai uses `Helm <https://helm.sh/docs/intro/using_helm/>`_ to install yatai.

Quick Installation
------------------

.. note:: This quick installation script should only be used for **development** and **testing** purposes.

This script will automatically install the following dependencies inside the :code:`yatai-system` namespace of the Kubernetes cluster:

* PostgreSQL
* MinIO

.. code:: bash

  bash <(curl -s "https://raw.githubusercontent.com/bentoml/yatai/main/scripts/quick-install-yatai.sh")

.. _yatai-installation-steps:

Installation Steps
------------------

.. note::

  If you don't have :code:`kubectl` installed and you are using :code:`minikube`, you can use :code:`minikube kubectl --` instead of :code:`kubectl`, for more details on using it, please check: `minikube kubectl <https://minikube.sigs.k8s.io/docs/commands/kubectl/>`_

1. Create Namespace
^^^^^^^^^^^^^^^^^^^

.. code:: bash

  kubectl create namespace yatai-system

2. Prepare PostgreSQL
^^^^^^^^^^^^^^^^^^^^^

.. tab-set::

    .. tab-item:: Use Existing PostgreSQL

        1. Prepare PostgreSQL connection params

        .. code:: bash

          export PG_PASSWORD=xxx
          export PG_HOST=1.1.1.1
          export PG_PORT=5432
          export PG_DATABASE=yatai
          export PG_USER=postgres
          export PG_SSLMODE=disable

        2. Create the PostgreSQL database :code:`$PG_DATABASE`

        .. code:: bash

          PGPASSWORD=$PG_PASSWORD psql \
              -h $PG_HOST \
              -p $PG_PORT \
              -U $PG_USER \
              -d postgres \
              -c "create database $PG_DATABASE"

    .. tab-item:: Create AWS RDS

        Prerequisites:

        - :code:`jq` command line tool. Follow the `official installation guide <https://stedolan.github.io/jq/download/>`__ to install :code:`jq`.

        - AWS CLI with RDS permission. Follow the `official installation guide <https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html>`__ to install AWS CLI.

        1. Prepare params

        .. code:: bash

          export PG_PASSWORD=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 20)
          export PG_USER=yatai
          export PG_DATABASE=yatai
          export PG_SSLMODE=disable
          export RDS_INSTANCE_IDENTIFIER=yatai-postgresql

          aws rds create-db-instance \
              --db-name $PG_DATABASE \
              --db-instance-identifier $RDS_INSTANCE_IDENTIFIER \
              --db-instance-class db.t3.micro \
              --engine postgres \
              --master-username $PG_USER \
              --master-user-password $PG_PASSWORD \
              --allocated-storage 20

        2. Get the RDS instance host and port

        .. code:: bash

          read PG_HOST PG_PORT < <(echo $(aws rds describe-db-instances --db-instance-identifier $RDS_INSTANCE_IDENTIFIER | jq '.DBInstances[0].Endpoint.Address, .DBInstances[0].Endpoint.Port'))
          PG_HOST=$(sh -c "echo $PG_HOST")

        3. Test the connection

        .. code:: bash

          kubectl -n yatai-system delete pod postgresql-ha-client 2> /dev/null || true; \
          kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
              --namespace yatai-system \
              --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
              --env="PGPASSWORD=$PG_PASSWORD" \
              --command -- psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DATABASE -c "select 1"

        Expected output:

        .. code:: bash

          ?column?
          ----------
                  1
          (1 row)

          pod "postgresql-ha-client" deleted

        .. note:: If there is no response for a long time, you can check if the VPC security group of the AWS RDS instance has opened port 5432 for public access

    .. tab-item:: Install New PostgreSQL

        .. note:: Do not recommend for production because this installation of PostgreSQL does not provide high availability and data replication.

        1. Install the :code:`postgresql-ha` helm chart:

        .. code:: bash

          helm repo add bitnami https://charts.bitnami.com/bitnami
          helm repo update bitnami
          helm upgrade --install postgresql-ha bitnami/postgresql-ha -n yatai-system --version 10.0.6

        2. Verify the :code:`postgresql-ha` installation:

        Monitor the postgresql-ha components until all of the components show a :code:`STATUS` of :code:`Running` or :code:`Completed`. You can do this by running the following command and inspecting the output:

        .. code:: bash

          kubectl -n yatai-system get pod -l app.kubernetes.io/name=postgresql-ha

        Example output:

        .. note:: You need to be patient for a while until the status of all pods becomes :code:`Running`, the number of pods depends on how many nodes you have

        .. code:: bash

          NAME                                    READY   STATUS    RESTARTS   AGE
          postgresql-ha-postgresql-0              1/1     Running   0          3m42s
          postgresql-ha-pgpool-56cf7b6b98-fs7g4   1/1     Running   0          3m42s
          postgresql-ha-postgresql-1              1/1     Running   0          3m41s
          postgresql-ha-postgresql-2              1/1     Running   0          3m41s

        3. Get the PostgreSQL connection params

        .. code:: bash

          export PG_PASSWORD=$(kubectl get secret --namespace yatai-system postgresql-ha-postgresql -o jsonpath="{.data.password}" | base64 -d)
          export PG_HOST=postgresql-ha-pgpool.yatai-system.svc.cluster.local
          export PG_PORT=5432
          export PG_DATABASE=yatai
          export PG_USER=postgres
          export PG_SSLMODE=disable

        4. Test PostgreSQL connection

        You can create a connection test by running the following command and inspecting the output:

        .. code:: bash

          kubectl -n yatai-system delete pod postgresql-ha-client 2> /dev/null || true; \
          kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
              --namespace yatai-system \
              --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
              --env="PGPASSWORD=$PG_PASSWORD" \
              --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "select 1"

        Expected output:

        .. code:: bash

          ?column?
          ----------
                  1
          (1 row)

          pod "postgresql-ha-client" deleted

        5. Create the PostgreSQL database :code:`$PG_DATABASE`

        You can create the database :code:`$PG_DATABASE` by running the following command and inspecting the output:

        .. code:: bash

          kubectl -n yatai-system delete pod postgresql-ha-client 2> /dev/null || true; \
          kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
              --namespace yatai-system \
              --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
              --env="PGPASSWORD=$PG_PASSWORD" \
              --command -- psql -h postgresql-ha-pgpool -p 5432 -U postgres -d postgres -c "create database $PG_DATABASE"

        Expected output:

        .. code:: bash

          If you don't see a command prompt, try pressing enter.
          CREATE DATABASE
          pod "postgresql-ha-client" deleted

Test PostgreSQL environment variables
"""""""""""""""""""""""""""""""""""""

You can create a connection test by running the following command and inspecting the output:

.. code:: bash

  kubectl -n yatai-system delete pod postgresql-ha-client 2> /dev/null || true; \
  kubectl run postgresql-ha-client --rm --tty -i --restart='Never' \
      --namespace yatai-system \
      --image docker.io/bitnami/postgresql-repmgr:14.4.0-debian-11-r13 \
      --env="PGPASSWORD=$PG_PASSWORD" \
      --command -- psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DATABASE -c "select 1"

Expected output:

.. code:: bash

  ?column?
  ----------
          1
  (1 row)

  pod "postgresql-ha-client" deleted

.. note:: If the above command does not respond for a long time and you are using AWS RDS, you can check if the VPC security group of the AWS RDS instance has port 5432 open for public access

3. Prepare Object Storage
^^^^^^^^^^^^^^^^^^^^^^^^^

.. note:: Now Yatai only support S3 protocol

.. tab-set::

    .. tab-item:: Use Existing AWS S3

      1. Prepare S3 connection params

      .. code:: bash

        export S3_REGION=YOUR-S3-REGION
        export S3_ENDPOINT="s3.${S3_REGION}.amazonaws.com"
        export S3_BUCKET_NAME=YOUR-BUCKET-NAME
        export S3_ACCESS_KEY=$(aws configure get default.aws_access_key_id)
        export S3_SECRET_KEY=$(aws configure get default.aws_secret_access_key)
        export S3_SECURE=true

      .. note:: Remember to replace YOUR-S3-REGION with your S3 region, replace YOUR-BUCKET-NAME with your S3 bucket name

    .. tab-item:: Use Existing AWS S3 with IAM

      1. Prepare S3 connection params

      .. code:: bash

        export S3_REGION=YOUR-S3-REGION
        export S3_ENDPOINT="s3.amazonaws.com"
        export S3_BUCKET_NAME=YOUR-BUCKET-NAME
        export S3_ACCESS_KEY=""
        export S3_SECRET_KEY=""
        export S3_SECURE=true

      .. note:: Remember to replace YOUR-S3-REGION with your S3 region, replace YOUR-BUCKET-NAME with your S3 bucket name

      2. Create IAM policy for S3 bucket access

      Create a file named :code:`s3-iam-policy.json` with the following content:

      .. code:: json

        {
           "Version":"2012-10-17",
           "Statement":[
              {
                 "Effect":"Allow",
                 "Action": "s3:ListAllMyBuckets",
                 "Resource":"*"
              },
              {
                 "Effect":"Allow",
                 "Action":["s3:ListBucket","s3:GetBucketLocation"],
                 "Resource":"arn:aws:s3:::${S3_BUCKET_NAME}"
              },
              {
                 "Effect":"Allow",
                 "Action":[
                    "s3:PutObject",
                    "s3:PutObjectAcl",
                    "s3:GetObject",
                    "s3:GetObjectAcl",
                    "s3:DeleteObject"
                 ],
                 "Resource":"arn:aws:s3:::${S3_BUCKET_NAME}/*"
              }
           ]
        }

      Replace :code:`${S3_BUCKET_NAME}` with your S3 bucket name:

      .. code:: bash

        envsubst < s3-iam-policy.json > s3-iam-policy.json

      .. note:: If you don't have :code:`envsubst` installed and you are using macOS, you can install it by running :code:`brew install gettext && brew link --force gettext`

      Create the IAM policy:

      .. code:: bash

        aws iam create-policy \
            --policy-name yatai-s3-access \
            --policy-document file://s3-iam-policy.json

      .. note:: Please store the ``Arn`` of the created policy, you will need it in the next step. The ``Arn`` format is like this: ``arn:aws:iam::ACCOUNT_ID:policy/yatai-s3-access``

      3. Create IAM ServiceAccount for S3 access

      Create ``yatai-system`` namespace:

      .. code:: bash

        kubectl create namespace yatai-system

      Create IAM ServiceAccount:

      .. code:: bash

        eksctl create iamserviceaccount \
            --name yatai \
            --namespace yatai-system \
            --cluster YOUR-CLUSTER \
            --region YOUR-REGION \
            --attach-policy-arn YOUR-IAM-POLICY-ARN \
            --approve

      .. note:: Remember to replace YOUR-CLUSTER with your EKS cluster name, replace YOUR-REGION with your EKS cluster region, replace YOUR-IAM-POLICY-ARN with the ``Arn`` of the created IAM policy

    .. tab-item:: Create New AWS S3

        Prerequisites:

        - AWS CLI with AWS S3 permission. Follow the `official installation guide <https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html>`__ to install AWS CLI

        1. Prepare params

        .. code:: bash

          export S3_BUCKET_NAME=yatai-registry
          export S3_REGION=ap-northeast-3
          export S3_ENDPOINT="s3.${S3_REGION}.amazonaws.com"
          export S3_SECURE=true

        2. Create AWS S3 bucket

        .. code:: bash

          aws s3api create-bucket \
              --bucket $S3_BUCKET_NAME \
              --region $S3_REGION \
              --create-bucket-configuration LocationConstraint=$S3_REGION

        3. Get :code:`ACCESS_KEY` and :code:`SECRET_KEY`

        .. code:: bash

          export S3_ACCESS_KEY=$(aws configure get default.aws_access_key_id)
          export S3_SECRET_KEY=$(aws configure get default.aws_secret_access_key)

        4. Verify S3 connection

        .. code:: bash

          kubectl -n yatai-system delete pod s3-client 2> /dev/null || true; \
          kubectl run s3-client --rm --tty -i --restart='Never' \
              --namespace yatai-system \
              --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
              --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
              --image quay.io/bentoml/s3-client:0.0.1 \
              --command -- sh -c "s3-client -e https://$S3_ENDPOINT listobj $S3_BUCKET_NAME && echo successfully"

        The output should be:

        .. code:: bash

          successfully
          pod "s3-client" deleted

    .. tab-item:: Install MinIO

        .. note::

          Do not recommend for production. Because you need to maintain the stability and data security of this important blob storage cluster yourself, it is recommended to use the blob storage provided by the public cloud vendor since many public cloud vendors (e.g. AWS) already have very mature blob storage.

        1. Install the :code:`minio-operator` helm chart

        .. code:: bash

          helm repo add minio https://operator.min.io/
          helm repo update minio

          export S3_ACCESS_KEY=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 20)
          export S3_SECRET_KEY=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 20)

          cat <<EOF | helm upgrade --install minio-operator minio/minio-operator -n yatai-system -f -
          tenants:
          - image:
              pullPolicy: IfNotPresent
              repository: quay.io/bentoml/minio-minio
              tag: RELEASE.2021-10-06T23-36-31Z
            metrics:
              enabled: false
              port: 9000
            mountPath: /export
            name: yatai-minio
            namespace: yatai-system
            pools:
            - servers: 4
              size: 20Gi
              volumesPerServer: 4
            secrets:
              accessKey: $S3_ACCESS_KEY
              enabled: true
              name: yatai-minio
              secretKey: $S3_SECRET_KEY
            subPath: /data
          EOF

        2. Verify the :code:`minio-operator` installation

        Monitor the minio-operator components until all of the components show a :code:`STATUS` of :code:`Running` or :code:`Completed`. You can do this by running the following command and inspecting the output:

        .. code:: bash

          kubectl -n yatai-system get pod -l app.kubernetes.io/name=minio-operator

        Expected output:

        .. note:: Wait until the status of all pods becomes :code:`Running` before proceeding

        .. code:: bash

          NAME                                     READY   STATUS    RESTARTS   AGE
          minio-operator-console-9d9cbbcc8-flzrw   1/1     Running   0          2m39s
          minio-operator-6c984995c9-l8j2j          1/1     Running   0          2m39s

        3. Verify the MinIO tenant installation

        Monitor the MinIO tenant components until all of the components show a :code:`STATUS` of :code:`Running` or :code:`Completed`. You can do this by running the following command and inspecting the output:

        .. code:: bash

          kubectl -n yatai-system get pod -l app=minio

        Expected output:

        .. note:: Since the pods are created by the :code:`minio-operator`, it may take a minute for these pods to be created. Wait until the status of all pods becomes :code:`Running` before proceeding.

        .. code:: bash

          NAME                 READY   STATUS    RESTARTS   AGE
          yatai-minio-ss-0-0   1/1     Running   0          143m
          yatai-minio-ss-0-1   1/1     Running   0          143m
          yatai-minio-ss-0-2   1/1     Running   0          143m
          yatai-minio-ss-0-3   1/1     Running   0          143m

        4. Prepare S3 connection params

        .. code:: bash

          export S3_ENDPOINT=minio.yatai-system.svc.cluster.local
          export S3_REGION=foo
          export S3_BUCKET_NAME=yatai
          export S3_SECURE=false
          export S3_ACCESS_KEY=$(kubectl -n yatai-system get secret yatai-minio -o jsonpath='{.data.accesskey}' | base64 -d)
          export S3_SECRET_KEY=$(kubectl -n yatai-system get secret yatai-minio -o jsonpath='{.data.secretkey}' | base64 -d)

        5. Test S3 connection

        .. code:: bash

          kubectl -n yatai-system delete pod s3-client 2> /dev/null || true; \
          kubectl run s3-client --rm --tty -i --restart='Never' \
              --namespace yatai-system \
              --env "AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY" \
              --env "AWS_SECRET_ACCESS_KEY=$S3_SECRET_KEY" \
              --image quay.io/bentoml/s3-client:0.0.1 \
              --command -- sh -c "s3-client -e http://$S3_ENDPOINT listbuckets && echo successfully"

        The output should be:

        .. note:: If the previous command reports an error that the service has not been initialized, please retry several times

        .. code:: bash

          successfully
          pod "s3-client" deleted


4. Install Yatai
^^^^^^^^^^^^^^^^

1. Install the Yatai Helm chart
"""""""""""""""""""""""""""""""

.. code:: bash

  helm upgrade --install yatai yatai \
      --repo https://bentoml.github.io/helm-charts \
      -n yatai-system \
      --set postgresql.host=$PG_HOST \
      --set postgresql.port=$PG_PORT \
      --set postgresql.user=$PG_USER \
      --set postgresql.database=$PG_DATABASE \
      --set postgresql.password=$PG_PASSWORD \
      --set postgresql.sslmode=$PG_SSLMODE \
      --set s3.endpoint=$S3_ENDPOINT \
      --set s3.region=$S3_REGION \
      --set s3.bucketName=$S3_BUCKET_NAME \
      --set s3.secure=$S3_SECURE \
      --set s3.accessKey=$S3_ACCESS_KEY \
      --set s3.secretKey=$S3_SECRET_KEY

.. note::

   If you are using AWS S3 With IAM Role, you should add the following flags to the helm command:

   .. code:: bash

      --set serviceAccount.create=false \
      --set serviceAccount.name=yatai

2. Verify the Yatai Installation
""""""""""""""""""""""""""""""""

.. code:: bash

  kubectl -n yatai-system get pod -l app.kubernetes.io/name=yatai

The output should look like the following:

.. note:: Wait until the status of all pods becomes :code:`Running`.

.. code:: bash

  NAME                    READY   STATUS    RESTARTS   AGE
  yatai-dbfbbb66f-67cq4   1/1     Running   0          45m
