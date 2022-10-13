============
Architecture
============

Yatai
#####

- Dashboard interface for all components.
- Registry for managing models and bentos with version control.
- APIs for programmatic access of Yatai functionality, e.g. bento push and pull.

The ``yatai`` component is the foundation of all other components in Yatai. It offers a dashboard user interface for viewing and managing bentos and models, deployments, clusters, and events. It also provides APIs for programmatic access to various Yatai functionality.

The ``yatai`` component installs and runs in a Kubernetes cluster and requires a relational database service (e.g. PostgreSQL) and an object storage service (e.g. AWS S3) to function.

Relational Database Service
***************************

A relational database used to store metadata, such as models, bentos, and users, required by the ``yatai`` component. Any PostgreSQL compatible relational databases can be used with provided user name and password. A database named ``yatai`` will be created to store tables of various metadata. We recommend using one of the following based on availability.

- `PostgreSQL <https://www.postgresql.org/>`_
- `AWS Fully Managed Relational Database (RDS) <https://aws.amazon.com/rds/>`_
- `Google Cloud SQL <https://cloud.google.com/sql>`_

Object Storage Service
**********************

Bucket storage for storing model and bento objects. Any S3 compatible services can be used providing the bucket name and endpoint address. We recommend using one of the following based on availability.

- `MinIO <https://min.io/>`_
- `AWS S3 <https://aws.amazon.com/s3/>`_
- `Google Cloud Storage (GCS) <https://cloud.google.com/storage>`_

Yatai Deployment
################

- Builds Docker images from bentos.
- Manages Docker images of bentos in an image registry.
- Deploy bentos in Kubernetes.

The ``yatai-deployment`` component is an add-on on top of ``yatai`` for building Docker images from bentos and deploying bentos to Kubernetes. One ``yatai-deployment`` component is required to be installed for every Kubernetes cluster ``yatai`` manages.

The ``yatai-deployment`` component installs and runs in a Kubernetes cluster and adds the ``BentoDeployment``  resource type. It requires a certificate manager, metrics server, image registry, and an ingress controller to function.

Certificate Manager
*******************

`Certificate manager <https://cert-manager.io/docs/>`_ adds certificate and certificate issuers resource types in the Kubernetes cluster. These resources are required by the ``BentoDeployment`` CRD conversion webhook.

Metrics Server
**************

`Metrics server <https://github.com/kubernetes-sigs/metrics-server>`_ collects resource metrics from Kubelets and exposes them in Kubernetes apiserver `horizontal pod autoscaling <https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#how-does-a-horizontalpodautoscaler-work>`_.

Docker Registry
***************

The Docker registry stores the OCI images built from bentos. Yatai Deployment fetches images from the registry during deployment.

Ingress Controller
******************

`Ingress controller <https://kubernetes.io/docs/concepts/services-networking/ingress/>`_ provides access and load balances external requests to the ``BentoDeployment`` . Yatai Deployment is agnostic about the `Ingress Class <https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class) or the ingress controller [implementations](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/>`_.

Observability
#############

- Stores and visualize metrics and logs from bento deployment in Grafana.
