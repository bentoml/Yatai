================
BentoRequest CRD
================

:code:`BentoRequest` is a Kubernetes `CRD <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ defined by :ref:`yatai-image-builder <concepts/architecture:yatai-image-builder>` component.

It is primarily used to describe bento's OCI image build request and to describe how to generate a :ref:`Bento <concepts/bento_crd:Bento CRD>` CR.

.. list-table:: Specification
    :widths: 25 25 50
    :header-rows: 1


    * - Field
      - Type
      - Description

    * - :code:`apiVersion`
      - :code:`string`
      - The version of the schema. Current version is ``resources.yatai.ai/v1alpha1``

    * - :code:`kind`
      - :code:`string`
      - The type of the resource. ``BentoRequest``

    * - :code:`metadata`
      - :code:`object`
      - The metadata of the resource. Refer to the Kubernetes API documentation for the fields of the ``metadata`` field

    * - :code:`spec.bentoTag`
      - :code:`string`
      - The tag of bento. **required**

    * - ``spec.downloadUrl``
      - ``string``
      - The url to download the bento tar file. If not specified, yatai-image-builder will fetch this information from yatai. **optional**

    * - :code:`spec.runners`
      - :code:`array`
      - The runners information. If not specified, yatai-image-builder will fetch this information from yatai. **optional**

    * - :code:`spec.runners[].name`
      - :code:`string`
      - The name of the runner. **required**


Example of a BentoRequest
-------------------------

.. code:: yaml

  apiVersion: resources.yatai.ai/v1alpha1
  kind: BentoRequest
  metadata:
    name: my-bento
    namespace: my-namespace
  spec:
    bentoTag: iris:1
    downloadUrl: s3://my-bucket/bentos/iris.tar.gz
    runners:
    - name: runner1
      runnableType: SklearnRunnable
      modelTags:
      - iris:1
    - name: runner2
