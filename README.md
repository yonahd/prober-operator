# prober-operator
## Overview
The prober-operator is a Kubernetes operator designed to automate the monitoring of Ingress resources within a Kubernetes cluster. It creates and manages Prometheus black box probes for Ingress resources based on a specified label.

## Description
The prober-operator simplifies the process of monitoring Ingress resources by dynamically creating Prometheus black box probes for domains specified in the Ingress labels. When an Ingress is labeled with `monitor: true`, the operator automatically generates a probe configuration and deploys it to the Prometheus black box exporter. This ensures that the specified domains are regularly probed for health and performance metrics.

### Features
- Automatic creation of Prometheus black box probes for labeled Ingress resources.
- Seamless integration with Prometheus black box exporter.
- Automatic cleanup of probes when Ingress resources are deleted or the monitoring label is removed.

### How it Works
1. Label Ingress resources with `monitor: true` to enable monitoring.
2. The prober-operator detects labeled Ingress resources and dynamically generates Prometheus black box probe configurations.
3. Probes are deployed to the Prometheus black box exporter, which regularly checks the specified domains.
4. Probes are automatically cleaned up when Ingress resources are deleted or the monitoring label is removed.

## Getting Started

### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

#### Configuration
ConfigMap: `prober-blackbox-config`
The prober-blackbox-config ConfigMap is used to configure the blackbox prober. It includes the prober URL and labels.

- **proberURL**: Specifies the prober URL of the blckbox exporter
- **labels**: Defines additional labels for the blackbox prober if specific labels are required for scraping


#### Helm
```sh
helm install prober-operator charts/prober-operator
```

#### Local
Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/prober-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. 
And it is required to have access to pull the image from the working environment. 
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/prober-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin 
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/prober-operator:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/prober-operator/<tag or branch>/dist/install.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024 yonahd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

