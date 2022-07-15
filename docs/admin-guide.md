# Yatai Administrator's Guide

This guide helps you to install and configure Yatai on a Kubernetes Cluster for your machine
learning team, using the official [Yatai Helm chart](https://github.com/bentoml/yatai-chart). Note
that Helm chart is the official supported method of installing Yatai.

By default, Yatai helm chart will install Yatai and its dependency services in the target Kubernetes cluster. Those dependency services include PostgreSQL, Minio, Docker registry, and Nginx Ingress Controller. Users can configure those services with existing infrastructure or cloud-based services via the Helm chart configuration yaml file.


- [System Overview](#system-overview)
- [Local Minikube Installation](#local-minikube-installation)
- [Production Installation](#production-installation)
  - Custom PostgreSQL database
    - [AWS RDS](#aws-rds)
  - Custom Docker Registry
    - [Docker hub](#docker-hub)
    - [AWR ECR](#aws-ecr)
  - Custom Blob Storage
    - [AWS S3](#aws-s3)
  - [Verify Yatai installation](#verify-yatai-installation)

## System Overview

### Namespaces:

When deploying Yatai with Helm,  `yatai-system`, `yatai-components`,  `yatai-operators`, `yatai-builder`, and `yatai` in the Kubernetes cluster.

* **Yatai-system:**

    All the services that run the yatai platform are group under `yatai-system` namespace.  These services include Yatai application and the default PostgreSQL database.

* **Yatai-components:**

    Yatai groups dependency services into components for easy management. These dependency services are deployed in the `yatai-components` namespace. Yatai installs the default deployment component after the service start.

    *Components and their managed services:*

    * *Deployment*: Nginx Ingress Controller, Minio, Docker Registry

    * *Logging*: Loki, Grafana

    * *Monitoring*: Prometheus, Grafana


* **Yatai-operators:**

    Yatai uses controllers to manage the lifecycle of Yatai component services. The controllers deployed in the `yatai-operators` namespace.

* **Yatai-builders:**

    All the automated jobs such as build docker images for models and bentos are executed in the `yatai-builder` namespace.

* **Yatai:**

    By default Yatai server will create a `yatai` namespace on the Kuberentes cluster for managing all the user created bento deployments. User can configure this namespace value in the Yatai web UI.


### Default dependency services installed:

* *PostgreSQL*:

    Yatai uses Postgres database to store model and bento’s metadata , deployment configuration, user activities and other information. Users can use their existing database or cloud provider’s services such as AWS RDS.  By default, Yatai will create a Postgres service in the Kubernetes cluster. For production usage, we recommend users to setup external Postgres database from cloud provider such as AWS RDS. This setup provides reliability, high performance and reliability, while persist the data, in case the Kuberentes cluster goes down.

* *Minio datastore*:

    Yatai uses the datastore as persistence layer for storing bentos. By default Yatai will start a Minio service. Users can configure to use cloud-based object store such as AWS S3. Cloud based object stores provide scalability, high performance and reliability at a desirable cost.  They are recommended for production usage.

* *Docker registry*:

    Yatai uses an internal docker registry to provide docker images access for deployments. For users who want to access the built images for other system, they can configure to use DockerHub or other cloud based docker registry services.

* *Nginx Ingress controller*:

    Yatai uses Nginx ingress controller to facilitates access to deployments and canary deployments.


See all available helm chart configuration options [here](./helm-configuration.md)

## Local Minikube Installation

Minikube is recommended for development and testing purpose only.

**prerequisites**
- Minikube version 1.20 or newer. Please follow the [official installation guide](https://minikube.sigs.k8s.io/docs/start/) to install Minikube.
- Recommend system with 4 CPUs and 4GB of RAM or more


**Step 1. Start a new minikube cluster**

If you have an existing minikube cluster, make sure to delete it first: `minikube delete`

```bash
# Start a new minikube cluster
minikube start --cpus 4 --memory 4096
```

**Step 2. Add and update Yatai helm chart**

```bash
helm repo add yatai https://bentoml.github.io/yatai-chart
helm repo update
```

**Step 3. Install Yatai chart**

This will create a new namespace `yatai-system` in the Minikube cluster, install Yatai and all its dependency services.

```bash
helm install yatai yatai/yatai -n yatai-system --create-namespace
```


### Verify installation

Check installation status:

```bash
helm status yatai -n yatai-system
```

Check Yatai containers status in the Minikube cluster:

```bash
# If kubectl is installed, run the following command:
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

Use minikube tunnel to expose Yatai Web UI locally::

```bash
# this requires enter your system password
sudo minikube tunnel
```

Once the minikube tunnel established, you can access the Yatai Web UI: http://localhost:8001/setup?token=<token>. You can find the URL link and the token again using `helm get notes yatai -n yatai-system` command.

You can also retrieve the token using `kubectl` command:

```bash
kubectl get pods --selector=app.kubernetes.io/name=yatai -n yatai-system \
    -o jsonpath='{.items[0].spec.containers[0].env[?(@.name=="YATAI_INITIALIZATION_TOKEN")].value}'
```



## Production Installation

To install and operate Yatai in production, we generally recommend using a dedicated database service(e.g. AWS RDS) and a managed blob storage (AWS S3 or managed MinIO cluster). The following demonstrates how to customize a Yatai deployment for production use.

**Prerequisite**

- Kubernetes cluster with version 1.20 or newer
- LoadBalancer (If you are using AWS EKS, Google GKS, your cluster is likely already configured with a working LoadBalancer. If you are using Kubernetes in a private data center, contact your system admin)

**Install Yatai with default configuration.**

1. Download and update Helm repo

    ```bash
    helm repo add yatai https://bentoml.github.io/yatai-chart

    helm repo update
    
    ```
    Create yatai-system namespace
    ```bash
    kubectl create namespace yatai-system
    ```

2. Install Yatai helm chart

    ```bash
    helm install yatai yatai/yatai -n yatai-system
    ```

3. Update Ingress to reference external ip

    By default, the host ip address that the Yatai ingress is initialized with is 127.0.0.1. In order to access yatai, you will need to update the host parameter in the ingress spec.

    This command will give you the external ip for the Yatai ingress:
    ```bash
    kubectl -n yatai-components get svc yatai-ingress-controller-ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
    ```

    If the previous command does not return anything, use the following command:
    ```bash
    dig +short `kubectl -n yatai-components get svc yatai-ingress-controller-ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'` | head -n 1
    ```
    
    Then replace "127.0.0.1" in the generated yatai domain name at this path with the external ip:
    ```bash
    .spec.rules[0].host
    ```
    
    Using this command:
    ```bash
    kubectl -n yatai-system edit ing yatai
    ```

### Custom PostgreSQL database

#### AWS RDS

Prerequisites:

- `jq`  command line tool. Follow the [official installation guide](https://stedolan.github.io/jq/download/) to install `jq`
- AWS CLI with RDS permission. Follow the [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) to install AWS CLI

1. Create an RDS db instance and secret

    ```bash
    USER_PASSWORD=secret99
    USER_NAME=admin
    DB_NAME=yatai
    INSTANCE_IDENTIFIER=yatai-postgres

    aws rds create-db-instance \
        --db-name $DB_NAME \
        --db-instance-identifier $INSTANCE_IDENTIFIER \
        --db-instance-class db.t3.micro \
        --engine postgres \
        --master-username $USER_NAME \
        --master-user-password $USER_PASSWORD \
        --allocated-storage 20
    
    kubectl create secret generic rds-password \
        --from-literal=password=$USER_PASSWORD \
        -n yatai-system
    ```

2. Get the RDS instance’s endpoint and port information

    ```bash
    read ENDPOINT PORT < <(echo $(aws rds describe-db-instances --db-instance-identifier $INSTANCE_IDENTIFIER | jq '.DBInstances[0].Endpoint.Address, .DBInstances[0].Endpoint.Port'))
    ```


1. Install Yatai chart with the RDS configuration

    ```bash
    DB_NAME=yatai

    helm install yatai yatai/yatai \
    	--set postgresql.enabled=false \
    	--set externalPostgresql.host=$host \
    	--set externalPostgresql.port=$port \
    	--set externalPostgresql.user=$USER_NAME \
    	--set externalPostgresql.existingSecret=rds-password \
        --set externalPostgresql.existingSecretPasswordKey=password \
    	--set externalPostgresql.database=$DB_NAME \
    	-n yatai-system
    ```


### Custom Docker registry

#### Docker hub

```bash
helm install yatai yatai/yatai \
	--set config.docker_registry.server='https://index.docker.io/v1' \
	--set config.docker_registry.username='MY_DOCKER_USER' \
	--set config.docker_registry.password='MY_DOCKER_USER_PASSWORD' \
	--set config.docker_registry.secure=true \
	-n yatai-system
```

#### AWS ECR

Prerequisites:

- `jq`  command line tool. Follow the [official installation guide](https://stedolan.github.io/jq/download/) to install `jq`
- AWS CLI with AWS ECR permission. Follow the [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) to install AWS CLI


1. Create AWS ECR repositories:
    1. Create repositories using default registry ID:

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

        aws ecr create-repository --registry-id $REGISTRY_ID --repository-name $BENTO_REPO
        aws ecr create-repository --registry-id $REGISTRY_ID --repository-name $MODEL_REPO

        # If the repositories are created use a different registry id from the default
        read ENDPOINT < <(echo $(aws ecr get-authorization-token --regsitry-ids $REGISTRY_ID | jq '.authorizationData[0].proxyEndpoint'))
        ```

2. Get ECR registry password info

    ```bash
    PASSWORD=$(aws ecr get-login-password)
    ```

3. Create Kubernetes secrets

    ```bash
    kubectl create secret generic yatai-docker-registry-credentials \
        --from-literal=password=$PASSWORD
        -n yatai-system
    ```

3. Install Yatai chart

    ```bash
    helm install yatai yatai/yatai \
    	--set externalDockerRegistry.enabled=true \
    	--set externalDockerRegistry.server=$ENDPOINT \
    	--set externalDockerRegistry.username=AWS \
    	--set externalDockerRegistry.secure=true \
    	--set externalDockerRegistry.bentoRepositoryName=$BENTO_REPO \
    	--set externalDockerRegistry.modelRepositoryName=$MODEL_REPO \
    	--set externalDockerRegistry.existingSecret=yatai-docker-registry-credentials \
    	--set externalDockerRegistry.existingSecretPasswordKey=password \
    	-n yatai-system
    ```


### Custom blob storage

#### AWS S3

Prerequisites:

- AWS CLI with AWS S3 permission. Follow the [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) to install AWS CLI


1. Configure AWS S3 bucket for Yatai

    ```bash
    BUCKET_NAME=my_yatai_bucket
    MY_REGION=us-west-1
    ENDPOINT='https://s3.amazonaws.com'

    aws s3 create-bucket --bucket $BUCKET_NAME --region MY_REGION
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

3. Install Yatai chart

    ```bash
    helm install yatai yatai/yatai \
        --set externalS3.enabled=true \
        --set externalS3.endpoint=$ENDPOINT \
    	--set externalS3.region=$MY_REGION \
    	--set externalS3.bucketName=$BUCKET_NAME \
    	--set externalS3.secure=true \
        --set externalS3.existingSecret="yatai-s3-credentials" \
    	--set externalS3.existingSecretAccessKeyKey=accessKeyId \
    	--set externalS3.existingSecretSecretKeyKey=secretAccessKey \
    	-n yatai-system
    ```


#### Verify Yatai installation


Check installation status with Helm

```bash
helm status yatai -n yatai-system
```

Run the `kubectl get svc` command to list out all of the services deployed by Yatai.

```bash
kubectl get svc --all-namespaces

# With default Yatai installation, the following services should be running.
yatai-components   console                                                       ClusterIP      10.108.153.223   <none>        9090/TCP,9443/TCP            16m
yatai-components   minio                                                         ClusterIP      10.96.13.19      <none>        80/TCP                       15m
yatai-components   operator                                                      ClusterIP      10.100.170.75    <none>        4222/TCP                     16m
yatai-components   yatai-docker-registry                                         ClusterIP      10.103.89.42     <none>        5000/TCP                     15m
yatai-components   yatai-ingress-controller-ingress-nginx-controller             LoadBalancer   10.106.230.175   127.0.0.1     80:32320/TCP,443:32351/TCP   17m
yatai-components   yatai-ingress-controller-ingress-nginx-controller-admission   ClusterIP      10.110.83.157    <none>        443/TCP                      17m
yatai-components   yatai-minio-console                                           ClusterIP      10.111.44.219    <none>        9090/TCP                     15m
yatai-components   yatai-minio-hl                                                ClusterIP      None             <none>        9000/TCP                     15m
yatai-system       yatai                                                         NodePort       10.99.111.8      <none>        80:30080/TCP                 20m
yatai-system       yatai-postgresql                                              ClusterIP      10.111.156.46    <none>        5432/TCP                     20m
yatai-system       yatai-postgresql-headless                                     ClusterIP      None             <none>        5432/TCP                     20m
```

Visit the link listed in the post Helm installation notes to access Yatai Web UI.
