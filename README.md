## d-operators
D-operators defines various declarative patterns to write kubernetes controllers. This uses [metac](https://github.com/AmitKumarDas/metac/) under the hood. Users will be able to create, delete, update, assert & patch one or more kubernetes resources _(native as well as custom)_ using a yaml file. This yaml file is itself a kubernetes custom resource that exposes these CRUD operations as building blocks to build a higher order controller.

This declarative approach is a pure intent based model instead of having specifications that are cluttered with scripts, kubectl, loops, conditions, templating and so on.

### Programmatic vs. Declarative
It is important to understand that these declarative patterns are built upon programmatic ones. The low level constructs _(read native k8s resources & custom resources)_ might be implemented in programming languages of one's choice. Use d-controller's YAMLs to aggregate these low level resources in a particular way to build a completely new kubernetes controller.

### When to use d-operators
D-operators is not meant to build complex controllers like Deployment, StatefulSet or Pod. However, if one needs to use these resources to build a new k8s native feature then d-operators should be considered to build this feature operating as a k8s controller.