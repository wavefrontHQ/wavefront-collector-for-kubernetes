## Wavefront Collector images for multi-platform

The [manifest-tool](https://github.com/estesp/manifest-tool) is created for the purpose of viewing, creating, and pushing the new manifests list object type in the Docker registry. The main purpose of manifest list is supporting multi-architecture and/or multi-platform images within a Docker registry.

Refer to [Manifest tool sample usage](https://github.com/estesp/manifest-tool#sample-usage) document for basic commands

### Usage

1. Build platform specific wavefront-collector docker image & push them to docker registry.
	For Linux platform:
	```
	make container
	docker tag wavefronthq/wavefront-kubernetes-collector:1.2.4 wavefronthq/wavefront-kubernetes-collector:1.2.4-linux
	docker push wavefronthq/wavefront-kubernetes-collector:1.2.4-linux
	```

	For Windows platform:
	```
	make container_win
	docker tag wavefronthq/wavefront-kubernetes-collector:1.2.4 wavefronthq/wavefront-kubernetes-collector:1.2.4-windows
	docker push wavefronthq/wavefront-kubernetes-collector:1.2.4-windows
	```

2. Create the manifest list with the manifest-tool.
    Manifest definition:
	```
	image: wavefronthq/wavefront-kubernetes-collector:1.2.4
	manifests:
	  -
	    image: wavefronthq/wavefront-kubernetes-collector:1.2.4-linux
	    platform:
	      architecture: amd64
	      os: linux
	  -
	    image: wavefronthq/wavefront-kubernetes-collector:1.2.4-windows
	    platform:
	      architecture: amd64
	      os: windows
	```

    Command to create manifest list (Windows platform):
	```
	.\manifest-tool-windows-amd64.exe push from-spec .\wavefront-collector-manifest.yml
	```

3. Check on the docker registry for multi-platform supported tag.
