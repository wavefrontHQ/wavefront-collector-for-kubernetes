K8S_AND_GO_TOOLING=/usr/local/bin/wget \
    /usr/local/bin/k9s \
    /usr/local/bin/kind \
	/usr/local/Cellar/go@1.15/1.15.14/libexec/bin/go \
    /usr/local/bin/git \
    /usr/local/bin/kubectl \
    /usr/local/bin/helm \
    /usr/local/bin/kustomize \
    /usr/local/bin/npm

DEV_EFFICIENCY_TOOLS=/usr/local/bin/git-duet \
	/usr/local/bin/github \
	/usr/local/bin/jq \
	/usr/local/bin/jump \
	/Applications/Flycut.app

DOCKER_GCLOUD_STUFF=$(HOME)/google-cloud-sdk/bin/gcloud \
	/Applications/GoLand.app \
	/Applications/Docker.app \
	/Applications/iTerm.app \
	/Applications/Tuple.app

BREW_BIN=/usr/local/bin/brew

DOTFILES=~/.nonsense \
	~/.zshrc \
	~/.git-authors \
	~/workspace/.envrc

OH_MY_ZSH=$(HOME)/.oh-my-zsh

~/*/.%: dotfiles/.%.templ
	./install-dotfile.sh $^ $@

~/.%: dotfiles/.%.templ
	./install-dotfile.sh $^ $@

#kind-create-cluster-and-run-integration-test:
	# do we even use this?
	#echo "kind-create-cluster-and-run-integration-test ain't do nuthin"
#	# setup kind cluster
#	kind create cluster
#
#	# build and deploy our code to kind
#	cd workspace/wavefront-collector-for-kubernetes
#	WAVEFRONT_API_KEY="<wavefront-api-key>" make integration-test

all: \
	$(BREW_BIN) \
	$(K8S_AND_GO_TOOLING) \
	$(DEV_EFFICIENCY_TOOLS) \
	$(DOCKER_GCLOUD_STUFF) \
	$(OH_MY_ZSH) \
	tweak-defaults \
	$(DOTFILES)
#	kind-create-cluster-and-run-integration-test \

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

/usr/local/Cellar/go@1.15/1.15.14/libexec/bin/go:
	brew install go@1.15
	brew link go@1.15 || true

$(HOME)/google-cloud-sdk/bin/gcloud:
	brew install google-cloud-sdk
	gcloud init # setup google cloud creds

#j:
#	#if ! command -v j; then
#	echo 'eval "$(jump shell zsh)"' >> ~/.zshrc
#	#fi

$(OH_MY_ZSH):
	curl -fsSL 'https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh' | sh
	rm $(HOME)/.zshrc

tweak-defaults:
	@## Improve scroll speed on mac
	@defaults write -g InitialKeyRepeat -int 10 # normal minimum is 15 (225 ms)
	@defaults write -g KeyRepeat -int 2 # normal minimum is 2 (30 ms)
