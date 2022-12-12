DEBUG?=false
COVER?=false
IMAGE_MODE_POSTFIX?=

ifeq ($(DEBUG), true)
	IMAGE_MODE_POSTFIX:=debug-
else ifeq ($(COVER), true)
	IMAGE_MODE_POSTFIX:=cover-
endif

container:
	echo "beep boop I'm a container"

debug-container:
	echo "we r now debuggering"

cover-container:
	echo "woah now cowboy better cover them horsies"

integration-test: $(IMAGE_MODE_POSTFIX)container
