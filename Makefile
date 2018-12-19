docker-build.datahub:
	cd datahub && $(MAKE) docker-build

docker-build.operator:
	cd operator && $(MAKE) docker-build