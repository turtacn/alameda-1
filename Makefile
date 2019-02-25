docker-build.datahub:
	cd datahub && $(MAKE) docker-build

docker-build.operator:
	cd operator && $(MAKE) docker-build

docker-build.admission-controller:
	cd admission-controller && $(MAKE) docker-build