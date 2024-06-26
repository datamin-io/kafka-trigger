# Kafka trigger
A CLI application for running Datamin pipelines using Kafka messages as input.

## How it works
Kafka-trigger acts as an OAuth client for Datamin API. It listens to Kafka topics and calls the [pipeline run endpoint](https://docs.datamin.io/datamin-api/api-endpoints#run-workflow) for each message, passing message body as pipeline input.
In the pipeline, the input can be received by the task "External trigger". When the workflow is run, this task receives the input data and passes it further to the next task of the pipeline.

## Configuration
Kafka-trigger is configured with environment variables. 
Besides the conventional way, the config variables can also be specified in the `.env` or `.env.local` file.
Main variables:
- **DTMN_KT_API_CLIENT_ID** — OAuth client ID.
- **DTMN_KT_API_CLIENT_SECRET** — OAuth client secret.
- **DTMN_KT_KAFKA_VERSION=3.1.0** — version of the Kafka server.
- **DTMN_KT_KAFKA_BOOTSTRAP_SERVERS="127.0.0.1:9092"** — a comma-separated list of Kafka bootstrap servers.
- **DTMN_KT_KAFKA_TOPIC_MAPPING="topic_1:pipeline_uuid_1,topic_1:pipeline_uuid_2,topic_2:pipeline_uuid_2"** — topic-to-pipeline mapping, a comma-separated list of `<topic name>:<pipeline uuid>` pairs.

More information about how to obrain your OAuth credentials is in our [documentation](https://docs.datamin.io/datamin-api/oauth-clients)

## Usage
1. Create a pipeline starting with [External trigger](https://docs.datamin.io/workflows-and-actions/tasks-ip#external-trigger) task.
2. Configure Kafka-trigger to read from a topic and initiate a pipeline run.

Example:
  * **Topic name:** test_topic
  * **Pipeline UUID:** e87ddc79-8e3f-4dae-92a8-8fff57ca81d3
  * **Topic-to-pipeline mapping:** `DTMN_KT_KAFKA_TOPIC_MAPPING="test_topic:e87ddc79-8e3f-4dae-92a8-8fff57ca81d3"`
