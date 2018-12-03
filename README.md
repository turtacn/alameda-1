[![Go Report Card](https://goreportcard.com/badge/github.com/containers-ai/Alameda)](https://goreportcard.com/report/github.com/containers-ai/Alameda)
[![CircleCI](https://circleci.com/gh/containers-ai/alameda.svg?style=shield)](https://circleci.com/gh/containers-ai/alameda)
[![codecov](https://codecov.io/gh/containers-ai/alameda/branch/master/graph/badge.svg)](https://codecov.io/gh/containers-ai/alameda)


# What is Alameda

### The Brain of Resources Orchestrator for Kubernetes

Alameda is a prediction engine that foresees future resource usage of your Kubernetes cluster down to the pod level. We use machine learning technology to provide resource predictions that enable dynamic scaling and scheduling of your containers, effectively making us the “brain” of Kubernetes resource orchestration. By providing full foresight of resource availability, demand, health, and impact, we enable cloud strategies that involve changing provisioned resources in real time.

Alameda agents in your cluster collect compute and I/O metrics, and send it to our engine, which will learn the continually changing resource demands and generate configuration recommendations that can be used by other container and storage orchestrators. We aim to help create a solution that automates pod scaling and scheduling, persistent volume provisioning, etc. to replace all manual configuration and orchestration tasks.

Automated orchestration (pod scaling and scheduling, persistent volume provisioning, etc.) means your cluster’s time spent reactively addressing resource failure and unavailability is reduced to a minimum. With Alameda, container and storage orchestrators can proactively make cluster-wide resource optimizations and reallocations before those problems arise.

You’re welcome to join and contribute to our community. We will be continually adding more support based on community demand and engagement in future releases.

### Contact
 
Please use the following to reach members of the community:

Slack: [Join our slack channel](https://join.slack.com/t/alameda-ai/signup)

Email: [Click](mailto:alameda@prophetstor.com)

Meeting: 

### Getting Started and Documentation

See our [Documentation](./docs/).

### Contributing

We welcome contributions. See [Contributing](CONTRIBUTING.md) to get started.

### Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please open an issue.
