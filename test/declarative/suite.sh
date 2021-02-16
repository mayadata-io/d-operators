#!/bin/bash

cleanup() {
  set +e

  echo -e "\n Display $ctrlbin-0 logs"
  k3s kubectl logs -n $ctrlbin $ctrlbin-0 || true

  echo -e "\n Display status of experiment with name 'inference'"
  k3s kubectl -n $ns get $group inference -ojson | jq .status || true

  echo -e "\n Display all experiments"
  k3s kubectl -n $ns get $group || true

  echo ""
  echo "--------------------------"
  echo "++ Clean up started"
  echo "--------------------------"

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

# Name of the targeted controller binary under test
ctrlbin="dope"

# Name of the daction controller binary
dactionctrlbin="daction"

# group that defines the Recipe custom resource
group="recipes.dope.mayadata.io"

# Namespace used by inference Recipe custom resource
ns="d-testing"

echo -e "\n Remove locally cached image $ctrlbin:e2e"
docker image remove $ctrlbin:e2e || true

echo -e "\n Remove locally cached image localhost:5000/$ctrlbin"
docker image remove localhost:5000/$ctrlbin || true

echo -e "\n Remove locally cached image localhost:5000/$dactionctrlbin"
docker image remove localhost:5000/$dactionctrlbin || true

echo -e "\n Run local docker registry at port 5000"
docker run -d -p 5000:5000 --restart=always --name e2eregistry registry:2

echo -e "\n Build $ctrlbin image as $ctrlbin:e2e"
docker build -t $ctrlbin:e2e ./../../

echo -e "\n Build $dactionctrlbin image as $dactionctrlbin:e2e"
docker build -t $dactionctrlbin:e2e ./../../tools/d-action

echo -e "\n Tag $ctrlbin:e2e image as localhost:5000/$ctrlbin"
docker tag $ctrlbin:e2e localhost:5000/$ctrlbin

echo -e "\n Tag $dactionctrlbin:e2e image as localhost:5000/$dactionctrlbin"
docker tag $dactionctrlbin:e2e localhost:5000/$dactionctrlbin

echo -e "\n Push the image to local registry running at localhost:5000"
docker push localhost:5000/$ctrlbin
docker push localhost:5000/$dactionctrlbin

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

echo -e "\n Apply d-operators based ci to K3s cluster"
k3s kubectl apply -f ci.yaml

echo -e "\n Apply test experiments to K3s cluster"
k3s kubectl apply -f ./experiments/command-creation-deletion.yaml

echo -e "\n Apply ci inference to K3s cluster"
k3s kubectl apply -f inference.yaml

echo -e "\n List configmaps if any in namespace $ns"
k3s kubectl get configmaps -n $ns

echo -e "\n Retry 50 times until inference experiment gets executed"
date
phase=""
for i in {1..50}
do
    ## TODO Remove these debug lines
    echo -e "\n List of available commands"
    k3s kubectl get command -A -o yaml
    deletionTime=$(k3s kubectl get command -n recipe-integration-cmd-testing testing-command -o=jsonpath='{.metadata.deletionTimestamp}')
    echo "-e \n Command deletion timestamp $deletionTime"
    echo -e "\n List of available jobs"
    k3s kubectl get job -A
    echo -e "\n List of available pods"
    k3s kubectl get po -A
    phase=$(k3s kubectl -n $ns get $group inference -o=jsonpath='{.status.phase}')
    echo -e "Attempt $i: Inference status: status.phase='$phase'"
    if [[ "$phase" == "" ]] || [[ "$phase" == "NotEligible" ]]; then
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
