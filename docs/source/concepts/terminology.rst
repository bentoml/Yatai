===========
Terminology
===========

Model
#####

A trained ML model instance needs to be saved with BentoML API. A model can be pushed to and pulled from Yatai. See the BentoML documentation for a more detailed explanation of `Model <https://docs.bentoml.org/en/latest/concepts/model.html>`_.

Model Registry
##############

The model registry is a hub for storing, versioning, and sharing models for collaboration. The relationship between :code:`model registry` and :code:`models` is analogous to :code:`Docker registry` and :code:`Docker images`.

Bento
#####

Bento 🍱 is a file archive with all the source code, models, data files and dependency configurations required for running a user-defined bentoml.Service, packaged into a standardized format. See the BentoML documentation for a more detailed explanation of `Bento <https://docs.bentoml.org/en/latest/concepts/bento.html>`_.

Bento Registry
##############

The bento registry is a hub for storing, versioning, and sharing :code:`Bento` for collaboration. The relationship between :code:`Bento registry` and :code:`Bentos` is analogous to :code:`Docker registry` and :code:`Docker images`.

Bento Deployment
################

:ref:`BentoDeployment <concepts/bentodeployment_crd:BentoDeployment CRD>` is a `Kubernetes Custom Resource Definition (CRD) <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ added to the Kubernetes cluster by Yatai Deployment. The CRD describes Bento deployments with a yaml file that can be queried. For a full list of the possible descriptive fields and an example CRD, see BentoDeployment CRD.
