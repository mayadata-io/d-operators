module mayadata.io/d-operators

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-resty/resty/v2 v2.2.0
	github.com/google/go-cmp v0.4.0
	github.com/pkg/errors v0.9.1
	k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/klog/v2 v2.0.0
	openebs.io/metac v0.5.0
)

replace (
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	openebs.io/metac => github.com/AmitKumarDas/metac v0.5.0
)
