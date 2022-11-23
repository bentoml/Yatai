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

BentoRequest CRD
################

:ref:`BentoRequest CRD <concepts/bentorequest_crd:BentoRequest CRD>` is a `Kubernetes Custom Resource Definition (CRD) <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ added to the Kubernetes cluster by :ref:`yatai-image-builder <concepts/architecture:yatai-image-builder>`. Each BentoRequest CR will generate a :ref:`Bento <concepts/bento_crd:Bento CRD>` CR with the same name after the OCI image is built. The CRD describes Bento image build information and runners information. For a full list of the possible descriptive fields and an example CRD, see :ref:`BentoRequest CRD <concepts/bentorequest_crd:BentoRequest CRD>`.

Bento CRD
#########

:ref:`Bento CRD <concepts/bento_crd:Bento CRD>` is a `Kubernetes Custom Resource Definition (CRD) <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ added to the Kubernetes cluster by :ref:`yatai-image-builder <concepts/architecture:yatai-image-builder>`. Bento CRs are often generated through the :ref:`BentoRequest CR <concepts/bentorequest_crd:BentoRequest CRD>`, but you can create a Bento CR manually, and :ref:`yatai-deployment <concepts/architecture:yatai-deployment>` relies on the Bento CR to get the Bento information. The CRD describes Bento image information and Bento runners information. For a full list of the possible descriptive fields and an example CRD, see :ref:`Bento CRD <concepts/bento_crd:Bento CRD>`.

BentoDeployment CRD
###################

:ref:`BentoDeployment CRD <concepts/bentodeployment_crd:BentoDeployment CRD>` is a `Kubernetes Custom Resource Definition (CRD) <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ added to the Kubernetes cluster by :ref:`yatai-deployment <concepts/architecture:yatai-deployment>`. The CRD describes Bento deployment. For a full list of the possible descriptive fields and an example CRD, see :ref:`BentoDeployment CRD <concepts/bentodeployment_crd:BentoDeployment CRD>`.
