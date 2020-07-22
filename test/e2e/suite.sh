#!/bin/bash

cleanup() {
  set +e

  echo ""
  echo "--------------------------"
  echo "++ Clean up started"
  echo "--------------------------"

  echo -e "\n Display $ctrlbin-0 logs"
  k3s kubectl logs -n $ctrlbin $ctrlbin-0

  echo -e "\n Display status of experiment with name 'inference'"
  k3s kubectl -n $ns get $group inference -ojson | jq .status || true

  echo -e "\n Uninstall K3s"
  /usr/local/bin/k3s-uninstall.sh > uninstall-k3s.txt 2>&1 || true

  echo -e "\n Stop local docker registry container"
  docker container stop e2eregistry || true

  echo -e "\n Remove local docker registry container"
  docker container rm -v e2eregistry || true
  
  echo ""
  echo "--------------------------"
  echo "++ Clean up completed"
  echo "--------------------------"
}

# Comment below if you donot want to invoke cleanup 
# after executing this script
#
# This is helpful if you might want to do some checks manually
# & verify the state of the Kubernetes cluster and resources
trap cleanup EXIT

# Uncomment below if debug / verbose execution is needed
#set -ex

echo ""
echo "--------------------------"
echo "++ E to E suite started"
echo "--------------------------"

# Name of the Kubernetes controller binary under test
ctrlbin="dope"

# kind: Recipe custom resource is used to build the test cases
group="recipes.dope.metacontroller.io"

# Namespace used by all Recipe custom resources
ns="d-testing"

echo -e "\n Remove locally cached image $ctrlbin:e2e"
docker image remove $ctrlbin:e2e || true

echo -e "\n Remove locally cached image localhost:5000/$ctrlbin"
docker image remove localhost:5000/$ctrlbin || true

echo -e "\n Run local docker registry at port 5000"
docker run -d -p 5000:5000 --restart=always --name e2eregistry registry:2

echo -e "\n Build $ctrlbin image as $ctrlbin:e2e"
docker build -t $ctrlbin:e2e ./../../

echo -e "\n Tag $ctrlbin:e2e image as localhost:5000/$ctrlbin"
docker tag $ctrlbin:e2e localhost:5000/$ctrlbin

echo -e "\n Push the image to local registry running at localhost:5000"
docker push localhost:5000/$ctrlbin

echo -e "\n Setup K3s registries path"
mkdir -p "/etc/rancher/k3s/"

echo -e "\n Copy registries.yaml to K3s registries path"
cp registries.yaml /etc/rancher/k3s/

echo -e "\n Download K3s if not available"
if true && k3s -v ; then
    echo ""
else
    curl -sfL https://get.k3s.io | sh -
fi

echo -e "\n Verify if K3s is up and running"
k3s kubectl get node

echo -e "\n Apply $ctrlbin manifests to K3s cluster..."
k3s kubectl apply -f ./../../manifests/

echo -e "\n Apply $ctrlbin as a statefulset to K3s cluster..."
cat <<EOF | k3s kubectl apply -f -
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app.mayadata.io/name: dope
  name: dope
  namespace: dope
spec:
  replicas: 1
  selector:
    matchLabels:
      app.mayadata.io/name: dope
  serviceName: ""
  template:
    metadata:
      labels:
        app.mayadata.io/name: dope
    spec:
      serviceAccountName: dope
      containers:
      - name: dope
        image: localhost:5000/dope
        command: ["/usr/bin/dope"]
        args:
        - --logtostderr
        - --run-as-local
        - -v=1
        - --discovery-interval=30s
        - --cache-flush-interval=240s
EOF

echo -e "\n Apply test namespace to K3s cluster"
k3s kubectl apply -f namespace.yaml

echo -e "\n Apply test experiments to K3s cluster"
k3s kubectl apply -f ./../experiments/

echo -e "\n Retry 25 times until inference experiment gets executed"
date
phase=""
for i in {1..25}
do
    phase=$(k3s kubectl -n $ns get $group inference -o=jsonpath='{.status.phase}')
    echo -e "Attempt $i: Inference status: status.phase='$phase'"
    if [[ "$phase" == "" ]]; then
        sleep 5 # Sleep & retry since experiment is in-progress
    else
        break # Abandon this loop since phase is set
    fi
done
date

if [[ "$phase" != "Completed" ]]; then
    echo ""
    echo "--------------------------"
    echo -e "++ E to E suite failed: status.phase='$phase'"
    echo "--------------------------"
    exit 1 # error since inference experiment did not succeed
fi

echo ""
echo "--------------------------"
echo "++ E to E suite passed"
echo "--------------------------"