# Twitter Voting

## Overview
This is a tutorial application from [Go Blueprints](https://www.packtpub.com/application-development/go-programming-blueprints-second-edition) that builds a distributed voting application using twitter as the source of votes.

I've made some small changes to help me learn better as well as dockerizing the system.

### Architecture Diagram
![Arch Diagram](https://www.packtpub.com/graphics/9781786468949/graphics/image_05_001.jpg)

## Dev Set up 
```
# Setup MongoDB Container
docker run -d --name mongo1 -v data:/data/db -p 27017:27017 mongo

# Start Mongo Shell
docker run -it --link mongo1:mongo --rm mongo sh -c 'exec mongo "$MONGO_PORT_27017_TCP_ADDR:$MONGO_PORT_27017_TCP_PORT/test"'

# Create test poll
use ballots
db.polls.insert({"title":"Test poll","options":["happy","sad","fail","win"]})

# Setup NSQ Service
docker run -d --name lookupd -p 4160:4160 -p 4161:4161 nsqio/nsq /nsqlookupd
docker run -d --name nsqd -p 4150:4150 -p 4151:4151 nsqio/nsq /nsqd --broadcast-address=<docker-host> --lookupd-tcp-address=<docker-host>:4160
docker run -d --name nsqadmin -p 4171:4171 nsqio/nsq /nsqadmin --nsqd-http-address=<docker-host>:4151
```

## Build Twitter Votes
```
cd twittervotes && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o twittervotes .
docker build --no-cache -t twittervotes .
```
