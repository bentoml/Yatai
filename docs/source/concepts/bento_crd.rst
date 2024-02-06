=========
Bento CRD
=========

:code:`Bento` is a Kubernetes `CRD <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_ defined by :ref:`yatai-image-builder <concepts/architecture:yatai-image-builder>` component.

It is primarily used to describe bento's information. :ref:`yatai-deployment <concepts/architecture:yatai-deployment>` to get bento information via Bento CR.

Bento CRs are often generated through the :ref:`BentoRequest CR <concepts/bentorequest_crd:BentoRequest CRD>`, but you can create a Bento CR manually, and :ref:`yatai-deployment <concepts/architecture:yatai-deployment>` relies on the Bento CR to get the Bento information.

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
      - The type of the resource. ``Bento``

    * - :code:`metadata`
      - :code:`object`
      - The metadata of the resource. Refer to the Kubernetes API documentation for the fields of the ``metadata`` field

    * - :code:`spec.bentoTag`
      - :code:`string`
      - The tag of bento. **required**

    * - ``spec.image``
      - ``string``
      - The OCI image URI of bento. **required**

    * - :code:`spec.runners`
      - :code:`array`
      - The runners information. **required**

    * - :code:`spec.runners[].name`
      - :code:`string`
      - The name of the runner. **required**


Example of a Bento
------------------

.. code:: yaml

  apiVersion: resources.yatai.ai/v1alpha1
  kind: Bento
  metadata:
    name: my-bento
    namespace: my-namespace
  spec:
    tag: iris:1
    image: my-registry.com/my-repository/iris:1
    runners:
    - name: runner1
      runnableType: SklearnRunnable
      modelTags:
      - iris:1
    - name: runner2
