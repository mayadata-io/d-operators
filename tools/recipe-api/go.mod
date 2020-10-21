module mayadata.io/recipe-api

go 1.13

require (
	k8s.io/apimachinery v0.17.3
	openebs.io/metac v0.4.0
)

replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	openebs.io/metac => github.com/AmitKumarDas/metac v0.4.0
)
