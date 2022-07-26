# Yatai Administrator's Guide

This guide helps you to install and configure Yatai on a Kubernetes Cluster for your machine
learning team, using the official [Yatai Helm chart](https://github.com/bentoml/yatai-chart).


- [System Overview](#system-overview)
- [Install for prototyping](#for-prototyping)
- [Install for production](#for-production)
  - Configure Database dependency
    - [AWS RDS](#aws-rds)
    - [GCP Cloud SQL](#gcp-cloud-sql)
    - [Azure SQL](#azure-sql)
  - Configure blob storage dependency
    - [AWS S3](#aws-s3)
    - [GCP Cloud Storage](#gcp-cloud-storage)
    - [Azure Storage](#azure-storage)
  - Configure image registry dependency
    - [AWR ECR](#aws-ecr)
    - [Docker hub](#dockerhub)
    - [GCP Container Registry](#gcp-container-registry)
    - [Azure Container Registry](#azure-container-registry)
  - [Verify installation](#verify-installation)

## System Overview

### Dependencies

1. Database dependency
    Yatai depends on a PostgreSQL database to store metadata, deployment configuration, users and other information.
2. Blob storage dependency
    Yatai uses a blob storage to store models and bentos.
3. Image registry dependency
    Yatai builds and stores bento images in an image registry for deployment.

### Namespaces:

When deploying Yatai with Helm,  `yatai-system`, `yatai-components`,  `yatai-operators`, `yatai-builder`, and `yatai` in the Kubernetes cluster.

* **Yatai-system:**

    All the services that run the Yatai platform are grouped under `yatai-system`` namespace.  These services include the Yatai application and the default PostgreSQL database.

* **Yatai-components:**

    Yatai groups dependency services into components for easy management. These dependency services are deployed in the `yatai-components` namespace. Yatai installs the default deployment component after the service start.

    *Components and their managed services:*

    * *Deployment*: Nginx Ingress Controller, Minio, Docker Registry


* **Yatai-operators:**

    Yatai uses controllers to manage the lifecycle of Yatai component services. The controllers are deployed in the `yatai-operators` namespace.

* **Yatai-builders:**

    All the automated jobs such as building docker images for models and bentos are executed in the `yatai-builder` namespace.

* **Yatai:**

    By default Yatai server will create a `yatai` namespace on the Kubernetes cluster for managing all the user-created bento deployments. Users can configure this namespace value in the Yatai web UI.


## Installation

### For prototyping

**prerequisites**
- Minikube version 1.20 or newer. Please follow the [official installation guide](https://minikube.sigs.k8s.io/docs/start/) to install Minikube.
- Recommend system with 6 CPUs and 8GB of RAM or more

**Step 1. Start a new Minikube cluster**

```bash
minikube delete
# Start a new minikube cluster
minikube start --cpus 4 --memory 4096
```

**Step 2. Add and update Yatai helm chart**

```bash
helm repo add yatai https://bentoml.github.io/yatai-chart
helm repo update
```

**Step 3. Install Yatai chart**

The following command will create a namespace `yatai-system` in the Minikube cluster, and install Yatai and all its dependency services.

```bash

```bash
helm install yatai yatai/yatai -n yatai-system --create-namespace
```


**Verify installation**

Check installation status:

```bash
helm status yatai -n yatai-system
```

Check Yatai containers status in the Minikube cluster:

```bash
# Run the following command:
kubectl get pod --all-namespaces

# If kubectl is not installed, run the following command:
minikube kubectl -- get pod --all-namespaces

# With default Yatai installation, the following pods should be running.
yatai-components   minio-operator-99f8cf4f4-6kzcz                                    1/1     Running             0               45s
yatai-components   yatai-ingress-controller-ingress-nginx-controller-7cf9494f59z5k   1/1     Running             0               112s
yatai-components   yatai-minio-console-84568cc987-twqtp                              0/1     ContainerCreating   0               45s
yatai-operators    deployment-yatai-deployment-comp-operator-58fd6b7667-x65cb        1/1     Running             0               2m16s
yatai-operators    yatai-csi-driver-image-populator-mkcz2                            2/2     Running             0               2m5s
yatai-system       yatai-6658d565d8-drk9f                                            1/1     Running             3 (2m41s ago)   4m25s
yatai-system       yatai-postgresql-0                                                1/1     Running             0               4m24s
```

Use Minikube tunnel to expose Yatai Web UI locally::

```bash
# this requires enter your system password
sudo minikube tunnel
```

Once established the Minikube tunnel, you can access the Yatai Web UI: http://localhost:8001/setup?token=<token>. You can find the URL link and the token again using `helm get notes yatai -n yatai-system` command.


### For production

In a Production environment, Yatai recommends using a Kubernetes cluster that is managed by a cloud provider. Uses external services for storage, database, and other services in case the Kubernetes cluster went down.

Prerequisites:


- Kubernetes cluster with version 1.20 or newer
- **LoadBalancer** (If you are using AWS EKS, or Google GKS, your cluster is likely already configured with a working LoadBalancer. If you are using Kubernetes in a private data center, contact your system admin)
- Helm installed and configured
- `jq` command line tool. Follow the [official installation guide](https://stedolan.github.io/jq/download/) to install `jq`
- `yatai-system` namespace. Run the following command to create the namespace:
    ```
    kubectl create namespace yatai-system
    ```
- Cloud provider CLI tools installed and configured.
    - For AWS:
       -  Download the AWS CLI and follow the [installation guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) to install the AWS CLI.
        - Run: `aws configure` to configure the AWS CLI.
    - For Google cloud
        - Download the Google Cloud SDK and follow the [installation guide](https://cloud.google.com/sdk/docs/downloads-intall) to install the Google Cloud SDK.
        - Run: `gcloud init` to configure the Google Cloud SDK.
    - For Azure cloud
        - Download the Azure CLI and follow the [installation guide](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) to install the Azure CLI.
        - Run: `az login` to login to Azure.


### Configure database dependency

#### AWS RDS:

1. Create a RDS DB instance

    ```bash
    DB_USER_PASSWORD=secret999
    DB_USER_NAME=admin
    DB_NAME=yatai
    INSTANCE_IDENTIFIER=yatai-postgres

    aws rds create-db-instance \
        --db-name $DB_NAME \
        --db-instance-identifier $INSTANCE_IDENTIFIER \
        --db-instance-class db.t3.micro \
        --engine postgres \
        --master-username $DB_USER_NAME \
        --master-user-password $DB_USER_PASSWORD \
        --allocated-storage 20
    ```

2. Create a secret on the Kubernetes cluster

    ```bash
    kubectl create secret generic rds-password \
        --from-literal=password=$DB_USER_PASSWORD
        -n yatai-system
    ```

3. Get the RDS instance's endpoint and port information

    ```bash
    read DB_ENDPOINT DB_PORT < <(echo $(aws rds describe-db-instances --db-instance-identifier $INSTANCE_IDENTIFIER | jq '.DBInstances[0].Endpoint.Address, .DBInstances[0].Endpoint.Port'))
    ```


#### GCP Cloud SQL:

*TODO*

#### Azure SQL:

*TODO*


### Configure blob storage dependency

#### AWS S3:

1. Configure AWS S3 bucket for Yatai

    ```bash
    BUCKET_NAME=my_yatai_bucket
    BLOB_REGION=us-west-1
    BLOB_ENDPOINT='https://s3.us-west-1.amazonaws.com'

    aws s3 create-bucket --bucket $BUCKET_NAME --region BLOB_REGION
    ```

2. Create Kubernetes secrets

    ```bash
    AWS_ACCESS_KEY_ID=$(aws configure get default.aws_access_key_id)
    AWS_SECRET_ACCESS_KEY=$(aws configure get default.aws_secret_access_key)
    kubectl create secret generic yatai-s3-credentials \
        --from-literal=accessKeyId=$AWS_ACCESS_KEY_ID \
        --from-literal=secretAccessKey=$AWS_SECRET_ACCESS_KEY
        -n yatai-system
    ```

#### GCP Cloud Storage:

*TODO*


#### Azure Storage:

*TODO*


### Configure Image Registry dependency

#### AWS ECR:

1. Create AWS ECR repositories:
    1. Create repositories using the default registry ID:

        ```bash
        BENTO_REPO=yatai-bentos
        MODEL_REPO=yatai-models

        aws ecr create-repository --repository-name $BENTO_REPO
        aws ecr create-repository --repository-name $MODEL_REPO

        # If the repositories are created using the default registry
        read ENDPOINT < <(echo $(aws ecr get-authorization-token | jq '.authorizationData[0].proxyEndpoint'))
        ```

    2. Create repositories using a specific registry ID for Yatai:

        ```bash
        BENTO_REPO=yatai-bentos
        MODEL_REPO=yatai-models
        REGISTRY_ID=yatai
        REGISTRY_USERNAME=AWS

        aws ecr create-repository --registry-id $REGISTRY_ID --repository-name $BENTO_REPO
        aws ecr create-repository --registry-id $REGISTRY_ID --repository-name $MODEL_REPO

        # If the repositories are created use a different registry id from the default
        read REGISTRY_ENDPOINT < <(echo $(aws ecr get-authorization-token --regsitry-ids $REGISTRY_ID | jq '.authorizationData[0].proxyEndpoint'))
        ```

2. Get ECR registry password info

    ```bash
    PASSWORD=$(aws ecr get-login-password)
    ```

3. Create Kubernetes secrets

    ```bash
    kubectl create secret generic yatai-docker-registry-credentials \
        --from-literal=password=$PASSWORD \
        -n yatai-system
    ```

#### DockerHub:

1. Create a Kubernetes secrets

    ```bash
    REGISTRY_ENDPOINT="https://index.docker.r.io/v1"
    REGISTRY_USERNAME="my_dockerhub_user_name"

    kubectl create secret generic yatai-docker-registry-credentials \
        --from-literal=password=MY_DOCKER_USER_PASSWORD \
        -n yatai-system
    ```

#### GCP Container Registry:

*TODO*

#### Azure Container Registry:

*TODO*

### Install Yatai


1. Create `my_value.yaml` file with the information from the previous steps

    ```yaml
    postgresql:
        enabled: false

    externalPostgresql:
        host: $DB_ENDPOINT
        port: $DB_PORT
        user: $DB_USER_NAME
        database: $DB_NAME
        sslmode: disable
        existingSecret: rds-password
        existingSecretPasswordKey: password

    externalS3:
        enabled: true
        endpoint: $BLOB_ENDPOINT
        region: $BLOB_REGION
        bucketName: $BUCKET_NAME
        secure: true
        existingSecret: 'yatai-s3-credentials'
        existingSecretAccessKeyKey: 'accessKeyId'
        existingSecretSecretKeyKey: 'secretAccessKey'

    externalDockerRegistry:
        enabled: true
        server: $REGISTRY_ENDPOINT
        username: $REGISTRY_USERNAME
        secure: true
        bentoRepositoryName: $BENTO_REPO
        modelRepositoryName: $MODEL_REPO
        existingSecret: 'yatai-docker-registry-credentials'
        existingSecretPasswordKey: password

    # Due to an issue with GKE https://github.com/moby/buildkit/issues/879,
    # please enable `dockerImageBuilder.privileged` to `true` when
    # installing Yatai.
    dockerImageBuilder:
        privileged: false
    ```

See all available helm chart configuration options [here](./helm-configuration.md)

2. Run Helm install command:

    ```bash
    helm install yatai yatai/yatai -n yatai-system -f my_value.yaml
    ```

### Update external IP address

By default, the host IP address that the Yatai ingress is initialized with is 127.0.0.1. To access Yatai, you will need to update the host parameter in the ingress spec.

This command will give you the external IP for the Yatai ingress:
```bash
kubectl -n yatai-components get svc yatai-ingress-controller-ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

If the previous command does not return anything, use the following command:
```bash
dig +short `kubectl -n yatai-components get svc yatai-ingress-controller-ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'` | head -n 1
```

Then replace "127.0.0.1" in the generated Yatai domain name at this path with the external ip:
```bash
.spec.rules[0].host
```

Using this command:
```bash
kubectl -n yatai-system edit ing yatai
```



## Verify installation


Check installation status with Helm

```bash
helm status yatai -n yatai-system
```

Visit the link listed in the post Helm installation notes to access Yatai Web UI.


## Debugging

If Yatai is not correctly installed, you can use the following commands to debug the installation:

```bash
kubectl -n yatai-operators logs -f deploy/deployment-yatai-deployment-comp-operator
```

If Yatai is unable to create a deployment, you can use the following commands to debug the installation:

```bash
kubectl -n yatai-components logs -f deploy/yatai-yatai-deployment-operator
```

If the Kubernetes pod created by Yatai cannot mount the volume, you can use the following commands to debug the installation:

```bash
kubectl -n yatai-components logs -f ds/yatai-csi-driver-image-populator -c image
```