FROM golang:1.16.6 as base

ARG YQ_VERSION=3.4.1

RUN apt-get update
RUN apt-get --yes install jq default-jre

# install yq
RUN wget https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64
RUN mv yq_linux_amd64 yq
RUN chmod +x yq
RUN mv yq /usr/bin/

# install gcloud
RUN echo "deb https://packages.cloud.google.com/apt cloud-sdk-bionic main" \
  | tee /etc/apt/sources.list.d/google-cloud-sdk.list
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN apt-get -y update
RUN apt-get -y install --no-install-recommends google-cloud-sdk
RUN apt-get clean
RUN apt-get autoremove --yes
