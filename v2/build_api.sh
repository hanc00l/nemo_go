#!/usr/bin/env bash

echo "build nemo for api server..."

# generate router and build server
bee generate routers -ctrlDir=pkg/webapi/controllers -routersFile=pkg/webapi/routers/commentsRouter_controllers.go -routersPkg=routers
# build serverapi
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o serverapi_darwin_amd64 cmd/serverapi/main.go
# generate swagger docs
cd pkg/webapi
bee generate docs
# move to swagger path
mv swagger/* ../../swagger/
rm -rf swagger
cd ../../

echo "build done."