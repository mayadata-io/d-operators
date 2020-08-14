## D-operators
D-operators define various declarative patterns to write kubernetes controllers. This uses [metac](https://github.com/AmitKumarDas/metac/) under the hood. Users can _create_, _delete_, _update_, _assert_, _patch_, _clone_, & _schedule_ one or more kubernetes resources _(native as well as custom)_ using a yaml file. D-operators expose a bunch of kubernetes custom resources that provide the building blocks to implement higher order controller(s).

D-operators follow a pure intent based approach to writing specifications **instead of** having to deal with yamls that are cluttered with scripts, kubectl, loops, conditions, templating and so on.

### A sample declarative intent
```yaml
apiVersion: dope.mayadata.io/v1
kind: Recipe
metadata:
  name: crud-ops-on-pod
  namespace: d-testing
spec:
  tasks:
  - name: apply-a-namespace
    apply: 
      state: 
        kind: Namespace
        apiVersion: v1
        metadata:
          name: my-ns
  - name: create-a-pod
    create: 
      state: 
        kind: Pod
        apiVersion: v1
        metadata:
          name: my-pod
          namespace: my-ns
        spec:
          containers:
          - name: web
            image: nginx
  - name: delete-the-pod
    delete: 
      state: 
        kind: Pod
        apiVersion: v1
        metadata:
          name: my-pod
          namespace: my-ns
  - name: delete-the-namespace
    delete: 
      state: 
        kind: Namespace
        apiVersion: v1
        metadata:
          name: my-ns
```

### Programmatic vs. Declarative
It is important to understand that these declarative patterns are built upon programmatic ones. The low level constructs _(read native Kubernetes resources & custom resources)_ might be implemented in programming language(s) of one's choice. Use d-controller's YAMLs to aggregate these low level resources in a particular way to build a completely new kubernetes controller.

### When to use D-operators
D-operators is not meant to build complex controller logic like Deployment, StatefulSet or Pod in a declarative yaml. However, if one needs to use available Kubernetes resources to build new k8s controller(s) then d-operators should be considered to build one. D-operators helps implement the last mile automation needed to manage applications & infrastructure in Kubernetes clusters.

### E to E testing
D-operators make use of its custom resource(s) to test its controllers. One can imagine these custom resources acting as the building blocks to implement a custom CI framework. One of the primary advantages with this approach, is to let custom resources remove the need to write code to implement test cases.

_NOTE: One can make use of these YAMLs (kind: Recipe) to test their own Kubernetes controllers declaratively_

Navigate to test/experiments to learn more on these YAMLs.

```sh
# Following runs the e2e test suite
#
# NOTE: test/e2e/suite.sh does the following:
# - d-operators' image known as 'dope' is built
# - a docker container is started & acts as the image registry
# - dope image is pushed to this image registry
# - k3s is installed with above image registry
# - d-operators' manifests are applied
# - experiments _(i.e. test code written as YAMLs)_ are applied
# - experiments are asserted
# - if all experiments pass then e2e is a success else it failed
# - k3s is un-installed
# - local image registry is stopped
sudo make e2e-test
```

### Available Kubernetes controllers
- [x] kind: Recipe
- [ ] kind: RecipeClass
- [ ] kind: RecipeGroupReport
- [ ] kind: RecipeDebug
- [ ] kind: Blueprint
- [-] kind: Validation
- [ ] kind: CodeCov
- [-] kind: HTTP
- [ ] kind: HTTPFlow
- [-] kind: Command
- [ ] kind: DaemonJob
- [ ] kind: UberLoop
