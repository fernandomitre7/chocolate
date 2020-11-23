GOPATH:=$(shell pwd)
GO:=go
GOFLAGS:=-v -p 1
WORKSPACE:=chocolate
OUT_EXEC=chocolate
SERVICE:=service #$(shell basename `pwd`)
TODAY := $(shell date +%s%N | cut -b1-13)
default: rebuild 


build:
	@echo "========== Building $@ =========="
	@echo "GOPATH = ${GOPATH}"
	sh -c 'export GOPATH=${GOPATH}; $(GO) build $(GOFLAGS) -o ${GOPATH}/bin/${OUT_EXEC} ${WORKSPACE}/${SERVICE}'
install: 
	@echo "========== Compiling $@ =========="
	sh -c 'export GOPATH=${GOPATH}; $(GO) install $(GOFLAGS) ${WORKSPACE}/${SERVICE}'
utils:
	@echo "========== Building DB Utils $@ =========="
	sh -c 'export GOPATH=${GOPATH}; $(GO) build $(GOFLAGS) -o ${GOPATH}/bin/dbutils ${WORKSPACE}/utils/db'
clean:
	@echo "Deleting binary files ..."; sh -c 'if [ -f bin/${OUT_EXEC} ]; then rm -f bin/${OUT_EXEC} && echo ../bin/${OUT_EXEC} ;fi;'
	@echo "Deleting db utils binary files ..."; sh -c 'if [ -f bin/dbutils ]; then rm -f bin/dbutils && echo ../bin/dbutils ;fi;'
	@echo "Moving log files "; sh -c 'if [ -f logs/${OUT_EXEC}.log ]; then mv logs/${OUT_EXEC}.log logs/${OUT_EXEC}.${TODAY}.log; fi;'
rebuild: clean build