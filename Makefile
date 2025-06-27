default: build

GO ?= $(shell which go)
APP = xiaozhi-gogo
BUILD_DIR = build
TAGS =  # -tags=release
GCFLAGS =
LD_FLAGS =


build: ensure_build_dir
	@echo "Building the project..."
	${GO} build ${TAGS} ${GCFLAGS} ${LD_FLAGS} -o ${BUILD_DIR}/${APP} ./main.go

clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf ${BUILD_DIR}/${APP}

ensure_build_dir:
	@echo "Ensuring build directory exists..."
	@echo mkdir -p ${BUILD_DIR}

install: build
	@echo "Installing the application..."
	@cp ${BUILD_DIR}/${APP} /usr/local/bin/${APP}
	@chmod +x /usr/local/bin/${APP}

