FROM alpine:3.4

COPY . /source/

RUN apk add --update bash \
  && ls -lta /source/ \
  && apk --update add git go ca-certificates \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/list-notifications-rw/" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && cp -r /source/* $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && go get ./... \
  && go install \
  && touch /config.yml \
  && mv ${GOPATH}/bin/list-notifications-rw / \
  && apk del go git \
  && rm -rf $GOPATH /var/cache/apk/*

EXPOSE 8080

CMD [ "/list-notifications-rw" ]
