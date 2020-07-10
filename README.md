## d-operators
D-operators define various declarative patterns to write kubernetes controllers. This uses [metac](https://github.com/AmitKumarDas/metac/) under the hood. Users can _create_, _delete_, _update_, _assert_, _patch_, _clone_, & _schedule_ one or more kubernetes resources _(native as well as custom)_ using a yaml file. D-operators expose a bunch of kubernetes custom resources that provide the building blocks to implement a higher order controller.

D-operators follow a pure intent based approach to writing specifications **instead of** having to deal with yamls that are cluttered with scripts, kubectl, loops, conditions, templating and so on.

### Programmatic vs. Declarative
It is important to understand that these declarative patterns are built upon programmatic ones. The low level constructs _(read native k8s resources & custom resources)_ might be implemented in programming language(s) of one's choice. Use d-controller's YAMLs to aggregate these low level resources in a particular way to build a completely new kubernetes controller.

### When to use d-operators
D-operators is not meant to build complex controllers like Deployment, StatefulSet or Pod in a declarative yaml. However, if one needs to use Deployment, StatefulSet, Pod, etc. to build new k8s controller(s) then d-operators' declarative constructs _(read custom resources)_ should be considered to build one. 

However, any controller which is complex and at the same time is **generic** enough to be used along with other kubernetes resources, can be implemented as a core d-operator custom resource.

### Available controllers
- [x] kind: Recipe
- [ ] kind: HTTP
- [ ] kind: Chef
- [ ] kind: Command
- [ ] kind: DaemonJob
- [ ] kind: MasterChef
