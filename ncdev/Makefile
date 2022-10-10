build:
	docker buildx build --platform linux/amd64 -t ghcr.io/barkardk/ncsandbox:latest .
	docker push ghcr.io/barkardk/ncsandbox:latest
package:
	helm package kubernetes/ncdev
	helm upgrade --install ncdev ncdev-*