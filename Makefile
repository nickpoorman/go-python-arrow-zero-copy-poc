# You may need to update this to reflect your PYTHONPATH.
PKG_CONFIG_PATH=${CONDA_PREFIX}/lib/pkgconfig
LD_LIBRARY_PATH=${CONDA_PREFIX}/lib/python3.7:${CONDA_PREFIX}/lib
PYTHONPATH=${CONDA_PREFIX}/lib/python3.7/site-packages:${PWD}/__python__
GO_PREFIX=PKG_CONFIG_PATH=${PKG_CONFIG_PATH} LD_LIBRARY_PATH=${LD_LIBRARY_PATH} PYTHONPATH=${PYTHONPATH}
GO_CMD=${GO_PREFIX} go

GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_TEST?=$(GO_CMD) test
GO_GET=$(GO_CMD) get
GO_RUN=${GO_CMD} run

DIST_DIR=bin

GO_SOURCES := $(shell find . -path -prune -o -name '*.go' -not -name '*_test.go')

.PHONY: default clean test build fmt bench run clean-cache

#
# Our default target, clean up, do our install, test, and build locally.
#
default: clean build

# Clean up after our install and build processes. Should get us back to as
# clean as possible.
#
clean:
	@for d in ./bin/*; do \
		if [ -f $$d ] ; then rm $$d ; fi \
	done
	rm -rf ./__python__/*.pyc

clean-cache: clean
	go clean -cache -testcache -modcache

#
# Do what we need to do to run our tests.
#
test: clean $(GO_SOURCES)
	$(GO_TEST) -count=1 -v $(GO_TEST_ARGS) ./...

#
# Build/compile our application.
#
build:
	echo "Building ${DIST_DIR}/poc"
	${GO_BUILD} -o ${DIST_DIR}/poc poc.go

#
# Run the benchmarks for the tools.
#
bench: $(GO_SOURCES)
	$(GO_TEST) $(GO_TEST_ARGS) -bench=. -run=- ./...

#
# Most of this is setup with telling python c-api where the python modules are.
#
run: clean build
	${GO_PREFIX} ./bin/poc
