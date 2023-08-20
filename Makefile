BUILD_ENV := CGO_ENABLED=0
LDFLAGS=-a -ldflags '-s -w' -trimpath -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}"

.PHONY: all package setup api clean
all: setup darwin linux windows
package: setup package_darwin package_darwin_worker package_linux package_linux_worker package_windows package_windows_worker clean

setup:
	mkdir -p release
	rm -rf release/*
	echo > log/runtime.log
	echo > log/access.log

darwin:
	${BUILD_ENV} GOARCH=amd64 GOOS=darwin go build ${LDFLAGS} -o server_darwin_amd64 cmd/server/main.go
	${BUILD_ENV} GOARCH=amd64 GOOS=darwin go build ${LDFLAGS} -o worker_darwin_amd64 cmd/worker/main.go
	${BUILD_ENV} GOARCH=amd64 GOOS=darwin go build ${LDFLAGS} -o daemon_worker_darwin_amd64 cmd/daemon_worker/main.go

linux:
	${BUILD_ENV} GOARCH=amd64 GOOS=linux go build ${LDFLAGS} -o server_linux_amd64 cmd/server/main.go
	${BUILD_ENV} GOARCH=amd64 GOOS=linux go build ${LDFLAGS} -o worker_linux_amd64 cmd/worker/main.go
	${BUILD_ENV} GOARCH=amd64 GOOS=linux go build ${LDFLAGS} -o daemon_worker_linux_amd64 cmd/daemon_worker/main.go

windows:
	${BUILD_ENV} GOARCH=amd64 GOOS=windows go build ${LDFLAGS} -o server_windows_amd64.exe cmd/server/main.go
	${BUILD_ENV} GOARCH=amd64 GOOS=windows go build ${LDFLAGS} -o worker_windows_amd64.exe cmd/worker/main.go
	${BUILD_ENV} GOARCH=amd64 GOOS=windows go build ${LDFLAGS} -o daemon_worker_windows_amd64.exe cmd/daemon_worker/main.go

package_darwin: setup darwin
	tar -cvzf release/nemo_darwin_amd64.tar \
      --exclude=thirdparty/xray/xray_linux_amd64 \
      --exclude=thirdparty/xray/xray_windows_amd64.exe \
      --exclude=thirdparty/nuclei/nuclei_linux_amd64 \
      --exclude=thirdparty/nuclei/nuclei_windows_amd64.exe \
      --exclude=thirdparty/fingerprinthub/observer_ward_linux_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_linux_amd64 \
      --exclude=thirdparty/httpx/httpx_windows_amd64.exe \
      --exclude=thirdparty/httpx/httpx_linux_amd64 \
      --exclude=thirdparty/goby/goby-cmd.exe \
      --exclude=thirdparty/goby/goby-cmd-linux \
      --exclude=thirdparty/massdns/massdns_windows_amd64.exe \
      --exclude=thirdparty/massdns/cygwin1.dll \
      --exclude=thirdparty/massdns/massdns_linux_amd64 \
      server_darwin_amd64 worker_darwin_amd64 daemon_worker_darwin_amd64 version.txt \
      conf log thirdparty web

package_linux: setup linux
	tar -cvzf release/nemo_linux_amd64.tar \
      --exclude=thirdparty/xray/xray_darwin_amd64 \
      --exclude=thirdparty/xray/xray_windows_amd64.exe \
      --exclude=thirdparty/nuclei/nuclei_darwin_amd64 \
      --exclude=thirdparty/nuclei/nuclei_windows_amd64.exe \
      --exclude=thirdparty/fingerprinthub/observer_ward_darwin_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_darwin_amd64 \
      --exclude=thirdparty/httpx/httpx_windows_amd64.exe \
      --exclude=thirdparty/httpx/httpx_darwin_amd64 \
      --exclude=thirdparty/goby/goby-cmd.exe \
      --exclude=thirdparty/goby/goby-cmd \
      --exclude=thirdparty/massdns/massdns_windows_amd64.exe \
      --exclude=thirdparty/massdns/cygwin1.dll \
      --exclude=thirdparty/massdns/massdns_darwin_amd64 \
      server_linux_amd64 worker_linux_amd64 daemon_worker_linux_amd64 version.txt \
      conf log thirdparty web docker* Dockerfile*

package_windows: setup windows
	tar -cvzf release/nemo_windows_amd64.tar \
      --exclude=thirdparty/xray/xray_darwin_amd64 \
      --exclude=thirdparty/xray/xray_linux_amd64 \
      --exclude=thirdparty/nuclei/nuclei_darwin_amd64 \
      --exclude=thirdparty/nuclei/nuclei_linux_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_darwin_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_linux_amd64 \
      --exclude=thirdparty/subfinder/subfinder_darwin_amd64 \
      --exclude=thirdparty/subfinder/subfinder_linux_amd64 \
      --exclude=thirdparty/httpx/httpx_darwin_amd64 \
      --exclude=thirdparty/httpx/httpx_linux_amd64 \
      --exclude=thirdparty/goby/goby-cmd \
      --exclude=thirdparty/goby/goby-cmd-linux \
      --exclude=thirdparty/massdns/massdns_darwin_amd64 \
      --exclude=thirdparty/massdns/massdns_linux_amd64 \
      server_windows_amd64.exe worker_windows_amd64.exe daemon_worker_windows_amd64.exe version.txt \
      conf log thirdparty web

package_darwin_worker: setup darwin
	tar -cvzf release/worker_darwin_amd64.tar \
      --exclude=thirdparty/xray/xray_linux_amd64 \
      --exclude=thirdparty/xray/xray_windows_amd64.exe \
      --exclude=thirdparty/nuclei/nuclei_linux_amd64 \
      --exclude=thirdparty/nuclei/nuclei_windows_amd64.exe \
      --exclude=thirdparty/fingerprinthub/observer_ward_linux_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_linux_amd64 \
      --exclude=thirdparty/httpx/httpx_windows_amd64.exe \
      --exclude=thirdparty/httpx/httpx_linux_amd64 \
      --exclude=conf/app.conf --exclude=server.yml \
      --exclude=thirdparty/goby/goby-cmd.exe \
      --exclude=thirdparty/goby/goby-cmd-linux \
      --exclude=thirdparty/massdns/massdns_windows_amd64.exe \
      --exclude=thirdparty/massdns/cygwin1.dll \
      --exclude=thirdparty/massdns/massdns_linux_amd64 \
      worker_darwin_amd64 daemon_worker_darwin_amd64 conf log thirdparty version.txt

package_linux_worker: setup linux
	tar -cvzf release/worker_linux_amd64.tar \
      --exclude=thirdparty/xray/xray_darwin_amd64 \
      --exclude=thirdparty/xray/xray_windows_amd64.exe \
      --exclude=thirdparty/nuclei/nuclei_darwin_amd64 \
      --exclude=thirdparty/nuclei/nuclei_windows_amd64.exe \
      --exclude=thirdparty/fingerprinthub/observer_ward_darwin_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_windows_amd64.exe \
      --exclude=thirdparty/subfinder/subfinder_darwin_amd64 \
      --exclude=thirdparty/httpx/httpx_windows_amd64.exe \
      --exclude=thirdparty/httpx/httpx_darwin_amd64 \
      --exclude=conf/app.conf --exclude=server.yml \
      --exclude=thirdparty/goby/goby-cmd.exe \
      --exclude=thirdparty/goby/goby-cmd \
      --exclude=thirdparty/massdns/massdns_windows_amd64.exe \
      --exclude=thirdparty/massdns/cygwin1.dll \
      --exclude=thirdparty/massdns/massdns_darwin_amd64 \
      worker_linux_amd64 daemon_worker_linux_amd64 conf log thirdparty version.txt

package_windows_worker: setup windows
	tar -cvzf release/worker_windows_amd64.tar \
      --exclude=thirdparty/xray/xray_darwin_amd64 \
      --exclude=thirdparty/xray/xray_linux_amd64 \
      --exclude=thirdparty/nuclei/nuclei_darwin_amd64 \
      --exclude=thirdparty/nuclei/nuclei_linux_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_darwin_amd64 \
      --exclude=thirdparty/fingerprinthub/observer_ward_linux_amd64 \
      --exclude=thirdparty/subfinder/subfinder_darwin_amd64 \
      --exclude=thirdparty/subfinder/subfinder_linux_amd64 \
      --exclude=thirdparty/httpx/httpx_darwin_amd64 \
      --exclude=thirdparty/httpx/httpx_linux_amd64 \
      --exclude=conf/app.conf --exclude=server.yml \
      --exclude=thirdparty/goby/goby-cmd \
      --exclude=thirdparty/goby/goby-cmd-linux \
      --exclude=thirdparty/massdns/massdns_darwin_amd64 \
      --exclude=thirdparty/massdns/massdns_linux_amd64 \
      worker_windows_amd64.exe daemon_worker_windows_amd64.exe conf log thirdparty version.txt

api:
	# generate router and build server
	bee generate routers -ctrlDir=pkg/webapi/controllers -routersFile=pkg/webapi/routers/commentsRouter_controllers.go -routersPkg=routers
	# build serverapi
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o serverapi_darwin_amd64 cmd/serverapi/main.go
	# generate swagger docs
	export GOROOT=/usr/local/opt/go/libexec && cd pkg/webapi && bee generate docs && mv swagger/* ../../swagger/ && rm -rf swagger && cd ../../

clean:
	rm -f server_darwin_amd64 worker_darwin_amd64 daemon_worker_darwin_amd64 \
    	server_linux_amd64 worker_linux_amd64 daemon_worker_linux_amd64 \
    	server_windows_amd64.exe worker_windows_amd64.exe daemon_worker_windows_amd64.exe \
    	serverapi_darwin_amd64