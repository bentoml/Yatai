=======================================
Model Deployment at scale on Kubernetes
=======================================

|github_stars| |actions_status| |documentation_status| |join_slack|

----

`Yatai(屋台, food cart) <https://github.com/bentoml/Yatai>`_ lets you deploy, operate and scale Machine Learning services on Kubernetes.

It supports deploying any ML models via `BentoML <https://github.com/bentoml/BentoML>`_, the unified model serving framework.


Why Yatai?
----------

🍱 Made for BentoML, deploy at scale

- Scale `BentoML <https://github.com/bentoml/BentoML>`_ to its full potential on a distributed system, optimized for cost saving and performance.
- Manage deployment lifecycles to deploy, update, or roll back via API or Web UI.
- Centralized registry providing the **foundation for CI/CD** via artifact management APIs, labeling, and WebHooks for custom integration.

🚅 Cloud native & DevOps friendly

- Kubernetes-native workflow via :doc:`BentoDeployment CRD <concepts/bentodeployment_crd>` (Custom Resource Definition), which can easily fit into an existing GitOps workflow.
- Native :doc:`integration with Grafana <observability/metrics>` stack for observability.
- Support for traffic control with Istio.
- Compatible with all major cloud platforms (AWS, Azure, and GCP).


Learn Yatai
-----------

.. grid:: 1 2 2 2
    :gutter: 3
    :margin: 0
    :padding: 3 4 0 0

    .. grid-item-card:: :doc:`💻 Installation Guide <installation/index>`
        :link: installation/index
        :link-type: doc

        A hands-on tutorial for installing Yatai

    .. grid-item-card:: :doc:`🔭 Observability <observability/index>`
        :link: observability/index
        :link-type: doc

        Learn how to monitor and debug your BentoDeployment

    .. grid-item-card:: :doc:`📖 Main Concepts <concepts/index>`
        :link: concepts/index
        :link-type: doc

        Explain the main concepts of Yatai

    .. grid-item-card:: `💬 BentoML Community <https://l.linklyhq.com/l/ktOX>`_
        :link: https://l.linklyhq.com/l/ktOX
        :link-type: url

        Join us in our Slack community where hundreds of ML practitioners are contributing to the project, helping other users, and discuss all things MLOps.



Staying Informed
----------------

The `BentoML Blog <http://modelserving.com>`_ and `@bentomlai <https://twitt
er.com/bentomlai>`_ on Twitter are the official source for
updates from the BentoML team. Anything important, including major releases and announcements, will be posted there. We also frequently
share tutorials, case studies, and community updates there.

To receive release notification, star & watch the `Yatai project on GitHub <https://github.com/bentoml/Yatai>`_. For release
notes and detailed changelog, see the `Releases <https://github.com/bentoml/Yatai/releases>`_ page.

----

Getting Involved
----------------

Yatai has a thriving open source community where hundreds of ML practitioners are
contributing to the project, helping other users and discuss all things MLOps.
`👉 Join us on slack today! <https://l.linklyhq.com/l/ktOX>`_


.. toctree::
   :hidden:

   installation/index
   observability/index
   concepts/index
   Community <https://l.linklyhq.com/l/ktOX>
   GitHub <https://github.com/bentoml/Yatai>
   Blog <https://modelserving.com>


.. spelling::

.. |actions_status| image:: https://github.com/bentoml/Yatai/workflows/Lint/badge.svg
   :target: https://github.com/bentoml/Yatai/actions
.. |documentation_status| image:: https://readthedocs.org/projects/yatai/badge/?version=latest&style=flat-square
   :target: https://docs.bentoml.org/projects/yatai
.. |join_slack| image:: https://badgen.net/badge/Join/BentoML%20Slack/cyan?icon=slack&style=flat-square
   :target: https://l.linklyhq.com/l/ktOX
.. |github_stars| image:: https://img.shields.io/github/stars/bentoml/Yatai?color=%23c9378a&label=github&logo=github&style=flat-square
   :target: https://github.com/bentoml/Yatai
