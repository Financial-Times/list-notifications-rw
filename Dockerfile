FROM alpine:3.4

COPY . /source/

RUN apk add --update bash \
  && ls -lta /source/ \
  && apk --update add git go ca-certificates \
  && cd /source/ \
  && BUILDINFO_PACKAGE="github.com/Financial-Times/service-status-go/buildinfo." \
  && VERSION="version=$(git describe --tag --always 2> /dev/null)" \
  && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
  && REPOSITORY="repository=$(git config --get remote.origin.url)" \
  && REVISION="revision=$(git rev-parse HEAD)" \
  && BUILDER="builder=$(go version)" \
  && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
  && cd - \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/list-notifications-rw/" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && cp -r /source/* $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && echo $LDFLAGS \
  && go get ./... \
  && go build -ldflags="${LDFLAGS}" \
  && mv ./list-notifications-rw / \
  && mv ./api/api.yml / \
  && touch /config.yml \
  && apk del go git \
  && rm -rf $GOPATH /var/cache/apk/*

EXPOSE 8080

CMD [ "/list-notifications-rw" ]
