FROM alpine:3.5
MAINTAINER DavyJ0nes <davy.jones@me.com>
ENV UPDATED_ON 13-02-2017

RUN apk add ca-certificates --update-cache
RUN mkdir -p /srv/app
WORKDIR /srv/app
ADD twittervotes /srv/app
CMD ["./twittervotes"]
