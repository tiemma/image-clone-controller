# Image Clone Controller

This controller clones docker images from other repositories
into another specified docker repository via the REPO_URL variable.


# Environment configuration

Find below a list of environment variables, some of them are required to successfully start the controller.

| Key                | Required | Default           | Function                                                                                                               |
|--------------------|----------|-------------------|------------------------------------------------------------------------------------------------------------------------|
| NAMESPACES_TO_SKIP | false    | kube-system       | Comma separated list of namespaces to ignore e.g "default, another-namespace"                                          |
| DELAY_PERIOD       | false    | 5                 | Time in minutes to wait before queuing a failed operation                                                              |
| IS_DEV_ENV         | false    | false             | Use dev configurations, good for local development only!                                                               |
| KUBECONFIG         | false    | in-cluster config | Specifies path to kubeconfig file, only used when IS_DEV_ENV is true                                                   |
| REPO_URL           | true     |                   | REQUIRED: Link to the "cache" repository e.g docker.io/k8s/ etc                                                        |
| DOCKER_CONFIG      | true     |                   | REQUIRED: Folder where Docker configuration used to authenticate to registry can be found. This is a folder path and the file can be mounted from a secret. |

For the DOCKER_CONFIG env, you can find a sample file to create it by running the commands below locally:
```bash
    docker login
    cat $HOME/.docker/config.json
```

You can then create a secret using the following command on the file previously obtained:
```bash
    # Change secret_name and config_path as necessary, config_path should be `$HOME/.docker/config.json`
    # secret_name should be dockercred as seen in the deployment spec
    kubectl create secret generic dockercred -n image-clone-controller-system --from-file=.dockerconfigjson=<config_path> --type=kubernetes.io/dockerconfigjson
```

Follow the instructions [here](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod) to mount the secrets as a file
in the configuration and set the environment variable to the folder mount path you specified 


# How to run it locally

The controller can be executed using the following command locally, set environment variables to required configuration

```bash
    git clone github.com/Tiemma/image-clone-controller
    cd image-clone-controller
    make run IS_DEV_ENV=true REPO_URL= DOCKER_CONFIG=
```


# Deploying to a kubernetes cluster

> NOTE: Requires kustomize

You can run the controller within a Kubernetes cluster by running the following commands:

```bash
    make deploy
```

It would parse all the configurations with kustomize and apply.


# Running tests

```bash
    make test
```


