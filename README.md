## d-operators
D-operators defines various declarative patterns to write kubernetes controllers. This uses [metac](https://github.com/AmitKumarDas/metac/) under the hood. Users will be able to create, delete, update, assert & patch one or more kubernetes resources _(native as well as custom)_ using a yaml file. This yaml file is a kubernetes custom resource that exposes above CRUD operations as building blocks to implement a higher order controller.

D-operators follow a pure intent based approach instead of having specifications that are cluttered with scripts, kubectl, loops, conditions, templating and so on.

### Programmatic vs. Declarative
It is important to understand that these declarative patterns are built upon programmatic ones. The low level constructs _(read native k8s resources & custom resources)_ might be implemented in programming languages of one's choice. Use d-controller's YAMLs to aggregate these low level resources in a particular way to build a completely new kubernetes controller.

### When to use d-operators
D-operators is not meant to build complex controllers like Deployment, StatefulSet or Pod. However, if one needs to use these resources to build a new k8s controller then d-operators should be considered to build the same.