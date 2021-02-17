#!/bin/bash

cleanup() {
  set +e

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
echo "++ Integration test suite started"
echo "--------------------------"

# Name of the targeted controller binary suitable for 
# running integration tests
ctrlbinary="dopeit"

echo -e "\n Delete previous integration test manifests if available"
k3s kubectl delete -f it.yaml || true

echo -e "\n Remove locally cached image $ctrlbinary:it"
docker image remove $ctrlbinary:it || true

echo -e "\n Remove locally cached image localhost:5000/$ctrlbinary"
docker image remove localhost:5000/$ctrlbinary || true

echo -e "\n Run local docker registry at port 5000"
docker run -d -p 5000:5000 --restart=always --name e2eregistry registry:2

echo -e "\n Build $ctrlbinary image as $ctrlbinary:it"
docker build -t $ctrlbinary:it ./../../ -f ./../../Dockerfile.testing

echo -e "\n Tag $ctrlbinary:it image as localhost:5000/$ctrlbinary"
docker tag $ctrlbinary:it localhost:5000/$ctrlbinary

echo -e "\n Push $ctrlbinary:it image to local registry running at localhost:5000"
docker push localhost:5000/$ctrlbinary

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

echo -e "\n Apply integration manifests to K3s cluster"
k3s kubectl apply -f it.yaml

echo -e "\n Will retry 50 times until integration test job gets completed"

echo -e "\n Start Time"
date
echo -e "\n"

phase=""
for i in {1..50}
do
    succeeded=$(k3s kubectl get job inference -n dit -o=jsonpath='{.status.succeeded}')
    failed=$(k3s kubectl get job inference -n dit -o=jsonpath='{.status.failed}')

    echo -e "Attempt $i: status.succeeded='$succeeded' status.failed='$failed'"

    if [[ "$failed" == "1" ]]; then
        break # Abandon this loop since job has failed
    fi

    if [[ "$succeeded" != "1" ]]; then
        sleep 15 # Sleep & retry since experiment is in-progress
    else
        break # Abandon this loop since succeeded is set
    fi
done

echo -e "\n End Time"
date
echo -e "\n"

echo -e "\n Display status of inference job"
k3s kubectl get job inference -n dit -ojson | jq .status || true

echo -e "\n Display test logs & coverage"
k3s kubectl -n dit logs -ljob-name=inference --tail=-1 || true

if [ "$succeeded" != "1" ] || [ "$failed" == "1" ]; then
    echo ""
    echo "--------------------------"
    echo -e "++ Integration test suite failed:"
    echo -e "+++ status.succeeded='$succeeded'"
    echo -e "+++ status.failed='$failed'"
    echo "--------------------------"
    exit 1 # error since inference experiment did not succeed
fi

echo ""
echo "--------------------------"
echo "++ Integration test suite passed"
echo "--------------------------"
