FROM alpine:3.5

COPY . .git /methode-content-placeholder-mapper/

RUN apk --update add git go libc-dev \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/methode-content-placeholder-mapper" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && mv /methode-content-placeholder-mapper/* $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && BUILDINFO_PACKAGE="github.com/Financial-Times/service-status-go/buildinfo." \
  && VERSION="version=$(git describe --tag --always 2> /dev/null)" \
  && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
  && REPOSITORY="repository=$(git config --get remote.origin.url)" \
  && REVISION="revision=$(git rev-parse HEAD)" \
  && BUILDER="builder=$(go version)" \
  && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
  && echo $LDFLAGS \
  && go get -u github.com/kardianos/govendor \
  && $GOPATH/bin/govendor sync \
  && go get -t -v ./... \
  && go build -ldflags="${LDFLAGS}" \
  && mv methode-content-placeholder-mapper /methode-content-placeholder-mapper-app \
  && rm -rf /methode-content-placeholder-mapper \
  && mv /methode-content-placeholder-mapper-app /methode-content-placeholder-mapper \
  && apk del go git libc-dev \
  && rm -rf $GOPATH /var/cache/apk/*

CMD [ "/methode-content-placeholder-mapper" ]