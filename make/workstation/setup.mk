WORKSPACE=$(HOME)/workspace
REPOS=$(WORKSPACE)/docs \
	$(WORKSPACE)/integrations \
	$(WORKSPACE)/wavefront-collector-for-kubernetes \
	$(WORKSPACE)/wavefront-kubernetes-adapter \
	$(WORKSPACE)/wavefront-operator-for-kubernetes \
	$(WORKSPACE)/wavefront-proxy \
	$(WORKSPACE)/prometheus-storage-adapter \
	$(WORKSPACE)/helm \
	$(WORKSPACE)/wavefront-kubernetes \
	$(WORKSPACE)/tmc-wavefront-operator

K8S_AND_GO_TOOLING=/usr/local/bin/wget \
    /usr/local/bin/k9s \
    /usr/local/bin/kind \
	/usr/local/Cellar/go@1.15/1.15.14/libexec/bin/go \
    /usr/local/bin/git \
    /usr/local/bin/kubectl \
    /usr/local/bin/helm \
    /usr/local/bin/kustomize \
    /usr/local/bin/npm

DEV_TOOLS=/usr/local/bin/git-duet \
	/usr/local/bin/github \
	/usr/local/bin/jq \
	/usr/local/bin/jump \
	/Applications/Flycut.app \
	$(HOME)/google-cloud-sdk/bin/gcloud \
	/usr/local/bin/aws \
	/Applications/GoLand.app \
	/Applications/Docker.app \
	/Applications/iTerm.app \
	/Applications/Tuple.app

BREW_BIN=/usr/local/bin/brew

DOTFILES=$(HOME)/.git-authors

OH_MY_ZSH=$(HOME)/.oh-my-zsh

setup-workstation: \
	$(BREW_BIN) \
	$(K8S_AND_GO_TOOLING) \
	$(DEV_TOOLS) \
	$(DOCKER_GCLOUD_STUFF) \
	$(REPOS) \
	$(DOTFILES)

$(HOME)/.%: make/workstation/.%.templ
	cp $^ $@

$(BREW_BIN):
	# Install Homebrew
	ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
	echo $(BREW_BIN)

/usr/local/bin/%:
	brew install $(shell basename $@)
	brew link $(shell basename $@) || true

/Applications/%.app:
	@BREW_TARGET=$(shell basename $@ | cut -f 1 -d '.' | awk '{print tolower($0)}');\
		brew install $$BREW_TARGET &&\
		echo "Congrats, you just installed $${BREW_TARGET}! Opening '$${BREW_TARGET}' for you now.";\
		open $@

/usr/local/bin/git-duet:
	brew install git-duet/tap/git-duet
	brew link git-duet/tap/git-duet || true

/usr/local/bin/aws:
	brew install awscli
	brew link awscli || true

/usr/local/Cellar/go@1.15/1.15.14/libexec/bin/go:
	brew install go@1.15
	brew link go@1.15 || true

$(HOME)/google-cloud-sdk/bin/gcloud:
	brew install google-cloud-sdk
	gcloud init # setup google cloud creds

$(OH_MY_ZSH):
	curl -fsSL 'https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh' | sh

oh-my-zsh: $(OH_MY_ZSH)

# BEGIN: Repos
clone-repos: $(REPOS)

$(WORKSPACE)/integrations:
	git clone git@github.com:sunnylabs/integrations.git $(WORKSPACE)/integrations

$(WORKSPACE)/docs:
	git clone git@github.com:/wavefrontHQ/docs.git $(WORKSPACE)/docs

$(WORKSPACE)/wavefront-collector-for-kubernetes:
	git clone git@github.com:/wavefrontHQ/wavefront-collector-for-kubernetes.git $(WORKSPACE)/wavefront-collector-for-kubernetes

$(WORKSPACE)/wavefront-kubernetes-adapter:
	git clone git@github.com:/wavefrontHQ/wavefront-kubernetes-adapter.git $(WORKSPACE)/wavefront-kubernetes-adapter

$(WORKSPACE)/wavefront-operator-for-kubernetes:
	git clone git@github.com:/wavefrontHQ/wavefront-operator-for-kubernetes.git $(WORKSPACE)/wavefront-operator-for-kubernetes

$(WORKSPACE)/wavefront-proxy:
	git clone git@github.com:/wavefrontHQ/wavefront-proxy.git $(WORKSPACE)/wavefront-proxy

$(WORKSPACE)/prometheus-storage-adapter:
	git clone git@github.com:/wavefrontHQ/prometheus-storage-adapter.git $(WORKSPACE)/prometheus-storage-adapter

$(WORKSPACE)/helm:
	git clone git@github.com:/wavefrontHQ/helm.git $(WORKSPACE)/helm

$(WORKSPACE)/wavefront-kubernetes:
	git clone git@github.com:/wavefrontHQ/wavefront-kubernetes.git $(WORKSPACE)/wavefront-kubernetes

$(WORKSPACE)/tmc-wavefront-operator:
	git clone git@gitlab.eng.vmware.com:tobs-k8s-group/tmc-wavefront-operator.git $(WORKSPACE)/tmc-wavefront-operator
# END: Repos
