============
Architecture
============

.. _yatai-architecture:

.. image:: /_static/img/architecture.png
   :width: 100%
   :alt: Architecture Diagram
   :class: only-light

.. image:: /_static/img/architecture-dark.png
   :width: 100%
   :alt: Architecture Diagram
   :class: only-dark

The versions of the ``Yatai`` and ``yatai-image-builder`` and ``yatai-deployment`` components need to be consistent in the minor version.

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

yatai-image-builder
###################

- Builds OCI images for bentos.
- Generate Bento CR.

The ``yatai-image-builder`` component is an add-on on top of ``yatai`` for building OCI images for bentos.

``yatai-image-builder`` runs in Kubernetes, it is the operator of BentoRequest CRD, it is responsible for reconcile BentoRequest CR and then build the image for Bento, after the image is built Bento CR is generated, ``yatai-deployment`` component will depend on Bento CR to deploy Bento in Kubernetes.

Docker Registry
***************

The Docker registry stores the OCI images built from bentos. yatai-deployment fetches images from the OCI image registry during deployment.

yatai-deployment
################

- Deploy bentos in Kubernetes.

The ``yatai-deployment`` component is an add-on on top of ``yatai`` for deploying bentos to Kubernetes.

``yatai-deployment`` runs in Kubernetes, it is the operator of BentoDeployment CRD, it is responsible for reconcile BentoDeployment CR and then deploying bentos to Kubernetes. It relies on Bento CR to get the OCI image and runners information, so it should install after the ``yatai-image-builder`` component installation.

cert-manager
************

`cert-manager <https://cert-manager.io/docs/>`__ adds certificate and certificate issuers resource types in the Kubernetes cluster. These resources are required by the ``BentoDeployment`` CRD conversion webhook.

metrics-server
**************

`metrics-server <https://github.com/kubernetes-sigs/metrics-server>`_ collects resource metrics from kubelet and exposes them in Kubernetes apiserver `horizontal pod autoscaling <https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#how-does-a-horizontalpodautoscaler-work>`_.

Ingress Controller
******************

`Ingress controller <https://kubernetes.io/docs/concepts/services-networking/ingress/>`_ provides access and load balances external requests to the ``BentoDeployment`` . Yatai Deployment is agnostic about the `Ingress Class <https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class>`_ or the ingress controller `implementations <https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/>`_.

Observability
#############

Yatai uses a group of tools to source and visualize metrics that helps pod observability and narrow down root causes when issues occur.

Prometheus
**********

`Prometheus <https://prometheus.io/>`_ is a monitoring system and time series database. It collects metrics exposed by ``BentoDeployment`` such as request count, duration, and in-progress.

Grafana
*******

`Grafana <https://grafana.com/>`_ is used for visualizing metrics and collecting logs. Grafana works with the Prometheus stack to turn pod metrics into time series graphs to provide analytic insight. Grafana offers the user a customizable dashboard with clean, organized graphs.

Promtail
********

`Promtail <https://grafana.com/docs/loki/latest/send-data/promtail/>`_ is a Grafana agent that acquires logs, turns logs into streams, and pushes the streams to Loki through an HTTP API. Promtail can be deployed as a DaemonSet on every Kubernetes node to collect logs from the node.

Loki
****

`Loki <https://grafana.com/oss/loki/>`_ is a stack that is used for log aggregation. It receives logs and streams and indexes them for efficient querying. Along with its highly efficient memory usage and scalability, Loki is designed to integrate seamlessly with Grafana for effective log storage and retrieval.
