module mayadata.io/d-operators

go 1.13

require (
	github.com/go-resty/resty/v2 v2.2.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/tidwall/gjson v1.6.0
	k8s.io/apimachinery v0.17.3
	openebs.io/metac v0.1.1-0.20200207112147-5bfc4b3f4af9
)

replace openebs.io/metac => github.com/AmitKumarDas/metac v0.1.1-0.20200207112147-5bfc4b3f4af9
