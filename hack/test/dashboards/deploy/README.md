
# First time setup instructions to use once a VM instance of image "debian-10-buster-v20220118" is created like this -
# https://console.cloud.google.com/compute/instancesDetail/zones/us-central1-a/instances/k8po-kind-stable-env-vm?project=wavefront-gcp-dev

gcloud auth login
gcloud config set project wavefront-gcp-dev
gcloud compute ssh --zone "us-central1-a" "k8po-kind-stable-env-vm"  --project "wavefront-gcp-dev"

# get root access and setup folders
sudo su -
mkdir -p /home/k8po/bin
cd /home/k8po/bin

# install docker
# Follow steps in https://docs.docker.com/engine/install/debian/#install-using-the-repository
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io

# install kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
chmod +x ./kind
# verify if the .kube file is generated in /home/k8po. If not copy it from the user directory where it might have been put at.

# install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl

echo "export PATH=$PATH:/home/k8po/bin" >> ~/.profile
echo "export KUBECONFIG=/home/k8po/.kube/config" >> ~/.profile
echo "export WAVEFRONT_TOKEN=<your-api-token-for-demo>" >> ~/.profile #Remember to set the demo env token here
echo "export WF_CLUSTER=demo" >> ~/.profile

cd ../
git clone https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes.git
cd wavefront-collector-for-kubernetes
git checkout K8SSAAS-773-stable-env # This branch is not meant to be merged to master
./hack/test/dashboards/deploy/deploy-demo.sh


#########################
# How to start everyday #
#########################
gcloud compute ssh --zone "us-central1-a" "k8po-kind-stable-env-vm"  --project "wavefront-gcp-dev"
sudo su -

cd /home/k8po/wavefront-collector-for-kubernetes/

## Set your values for WAVEFRONT_TOKEN and WF_CLUSTER in ~/.profile
#vi ~/.profile
#source ~/.profile

## Do the below when you are trying to pull in changes and/or do a new deploy
#git checkout -- .
#git pull

./hack/test/dashboards/deploy/deploy-demo.sh
