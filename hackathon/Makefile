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

~/*/.%: dotfiles/.%.templ
	./install-dotfile.sh $^ $@

~/.%: dotfiles/.%.templ
	./install-dotfile.sh $^ $@

all: \
	$(DOTFILES) \
	$(BREW_BIN) \
	$(K8S_AND_GO_TOOLING) \
	$(DEV_EFFICIENCY_TOOLS) \
	$(DOCKER_GCLOUD_STUFF) \
	tweak-defaults

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

install:
	## Install powerlevel10k bash prompt
	brew install romkatv/powerlevel10k/powerlevel10k
	echo "source $(brew --prefix)/opt/powerlevel10k/powerlevel10k.zsh-theme" >>~/.zshrc

	# h*cking gcloud can't just have a normal install... or can it? TODO
	if [ ! -d ${HOME}/google-cloud-sdk ]; then
	curl https://sdk.cloud.google.com > install.sh; bash install.sh --disable-prompts;
	fi

	if ! command -v gcloud; then
	echo "source ~/google-cloud-sdk/path.zsh.inc" >> ~/.zshrc
	echo "source ~/google-cloud-sdk/completion.zsh.inc" >> ~/.zshrc
	fi

	# setup kind cluster
	kind create cluster

	# build and deploy our code to kind
	cd workspace/wavefront-collector-for-kubernetes
	WAVEFRONT_API_KEY="<wavefront-api-key>" make integration-test

tweak-defaults:
	## Improve scroll speed on mac
	defaults write -g InitialKeyRepeat -int 10 # normal minimum is 15 (225 ms)
	defaults write -g KeyRepeat -int 2 # normal minimum is 2 (30 ms)
