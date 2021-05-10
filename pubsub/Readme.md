# GCP Pub/Sub Tracing Example

### Run Subscribers
set following environment variables
```sh
GOOGLE_CLOUD_PROJECT=<gcp project name>
PUBSUB_TOPIC=<topic name>
GOOGLE_APPLICATION_CREDENTIALS=<gcp-credentials>
```

```sh
go run ./subscriber
```

### Run Publisher

set following environment variables
```sh
GOOGLE_CLOUD_PROJECT=<gcp project name>;
PUBSUB_TOPIC=<topic name>;
GOOGLE_APPLICATION_CREDENTIALS=<gcp-credentials>
```
```sh
go run ./publisher
```

#### Publish a Message
```sh
curl localhost:8080/pubsub/publish
```

