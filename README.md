# Kafka trigger
A CLI application for running Datamin pipelines using Kafka messages as input.

## How it works
Kafka-trigger acts as an OAuth client for Datamin API. It listens to Kafka topics and calls the [pipeline run endpoint](https://docs.datamin.io/datamin-api/api-endpoints#run-workflow) for each message, passing message body as pipeline input.
In the pipeline, the input can be received by the task "External trigger". When the pipeline is run, this task receives the input data and passes it further to the next task of the pipeline.

## Configuration
Kafka-trigger is configured with environment variables. 
Besides the conventional way, the config variables can also be specified in the `.env` or `.env.local` file.
Main variables:

| Variable | Description |
|--|-|
| **DTMN_API_CLIENT_ID** | OAuth client ID. Create a new client here: https://app.datamin.io/api-clients |
| **DTMN_API_CLIENT_SECRET** | OAuth client secret |
| **DTMN_KAFKA_VERSION** | Kafka server version. Example: `3.1.0` |
| **DTMN_KAFKA_BOOTSTRAP_SERVERS** | Comma-separated list of Kafka bootstrap servers. Example: `127.0.0.1:9092` |
| **DTMN_KAFKA_TOPIC_MAPPING** | Topic-to-pipeline mapping, a comma-separated list of `<topic name>:<pipeline uuid>` pairs. Example: `topic_1:pipeline_uuid_1,topic_1:pipeline_uuid_2,topic_2:pipeline_uuid_2` |


More information about how to obtain your OAuth credentials is in our [documentation](https://docs.datamin.io/datamin-api/oauth-clients).

## Usage
1. Create a pipeline starting with [External trigger](https://docs.datamin.io/workflows-and-actions/tasks-ip#external-trigger) task.
2. Configure Kafka-trigger to read from a topic and initiate a pipeline run.

### Example:
  * **Topic name:** `test_topic`
  * **Pipeline UUID:** `e87ddc79-8e3f-4dae-92a8-8fff57ca81d3`
  * **Topic-to-pipeline mapping:** `DTMN_KAFKA_TOPIC_MAPPING="test_topic:e87ddc79-8e3f-4dae-92a8-8fff57ca81d3"`

---

# S3 Listener Lambda

An AWS lambda for listening to events from S3 and running Dataming pipelines using the uploaded file content as input.

## How it works

After installation, the lambda listens to S3 events about created objects (i.e. uploaded files). If the file path matches the configured expression, the lambda reads their contents and runs the configured pipelines for each file.

## Installation

### Pre-requisites

Create an OAuth client for the lambda here: https://app.datamin.io/api-clients and copy the client ID and client secret key.

### Method 1: From Zip archive

Follow [this guide](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html) to install the lambda from a zip archive.

### Method 2: From Datamin container registry

1. Navigate to AWS Lambda → Functions section.
2. Click "Create function".
3. Choose "Container image".
4. Enter function name.
5. Enter container image URI: `241485570393.dkr.ecr.eu-central-1.amazonaws.com/datamin-integration:latest`
6. Architecture: `x86_64`.
7. Execution role: pick `Create a new role with basic Lambda permissions`.

## Configuration

First, make sure that the lambda has read access to objects in all S3 buckets you plan to use.

Second, follow [this guide](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html) to set up a S3 trigger for the lambda.

Then, you need to configure several environment variables for the lamba to work.

Navigate to Configuration → Environment variables section of the lambda.

| Variable | Description |
|--|-|
| **DTMN_API_CLIENT_ID** | OAuth client ID. Create a new client here: https://app.datamin.io/api-clients |
| **DTMN_API_CLIENT_SECRET** | OAuth client secret |
| **DTMN_S3_MAPPING** | Mapping of path expressions to pipeline UUIDs. See below. |

### Mapping

The format is:

`<path expression #1>`:`<pipeline UUID #1>,<pipeline UUID #2>`;`<path expression #2>`:`<pipeline UUID #3>`

**Path expression**: A Glob-like expression to match the full object path against, including the bucket name.

If the object path matches the expression, lambda will read the object contents and send them as input to all pipelines which UUIDs are enumerated in the config.

**Pipeline UUIDs:** a comma-separated list of pipelines to run if the object path matches the expression.

#### Mapping example
- `my-bucket-name/logs/*.json:18d27bb8-f886-4072-9165-485baa3e332d;my-bucket-name/*/daily.json:2f0cfe9e-3571-49a4-9103-cbd9d3d58c2b`