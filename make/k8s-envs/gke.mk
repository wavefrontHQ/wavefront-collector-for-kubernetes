GCP_PROJECT=wavefront-gcp-dev
GCP_REGION=us-central1
GCP_ZONE?=c
GCR_ENDPOINT=us.gcr.io
GCR_PREFIX=$(GCP_PROJECT)
NUMBER_OF_NODES?=3

target-gke: connect-to-gke gke-connect-to-cluster

connect-to-gke:
	gcloud config set project $(GCP_PROJECT)
	gcloud auth configure-docker --quiet

gke-cluster-name-check:
	@if [ -z ${GKE_CLUSTER_NAME} ]; then echo "Need to set GKE_CLUSTER_NAME" && exit 1; fi

gke-connect-to-cluster: gke-cluster-name-check
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GCP_REGION)-$(GCP_ZONE) --project $(GCP_PROJECT)

delete-gke-cluster: gke-cluster-name-check connect-to-gke
	echo "Deleting GKE K8s Cluster: $(GKE_CLUSTER_NAME)"
	gcloud container clusters delete $(GKE_CLUSTER_NAME) --zone $(GCP_REGION)-$(GCP_ZONE) --quiet

create-gke-cluster: gke-cluster-name-check connect-to-gke
	echo "Creating GKE K8s Cluster: $(GKE_CLUSTER_NAME)"
	gcloud container clusters create $(GKE_CLUSTER_NAME) --machine-type=e2-standard-2 --zone=$(GCP_REGION)-$(GCP_ZONE) --enable-ip-alias --create-subnetwork range=/21 --num-nodes=$(NUMBER_OF_NODES)
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GCP_REGION)-$(GCP_ZONE) --project $(GCP_PROJECT)
	kubectl create clusterrolebinding --clusterrole cluster-admin \
		--user $$(gcloud auth list --filter=status:ACTIVE --format="value(account)") \
		clusterrolebinding
