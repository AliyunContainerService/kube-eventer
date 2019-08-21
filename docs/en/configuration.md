Configuring sources
===================
### Kubernetes
To use the kubernetes source add the following flag:

	--source=kubernetes:<KUBERNETES_MASTER>[?<KUBERNETES_OPTIONS>]

If you're running kube-eventer in a Kubernetes pod you can use the following flag:

	--source=kubernetes:https://kubernetes.default

If you don't want to setup inClusterConfig, you can still use kube-eventer! To run without auth, use the following config:

	--source=kubernetes:http://<address-of-kubernetes-master>:<http-port>?inClusterConfig=false

This requires the apiserver to be setup completely without auth, which can be done by binding the insecure port to all interfaces (see the apiserver `--insecure-bind-address` option) but *WARNING* be aware of the security repercussions. Only do this if you trust *EVERYONE* on your network. Or you can setup proxy with command `kubectl proxy --port=8080` in local when debugging kube-eventer.

The following options are available:
* `inClusterConfig` - Use kube config in service accounts (default: true)
