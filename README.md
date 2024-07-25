# Kafka trigger
A CLI application for subscribing to Apache Kafka topics and triggering Datamin pipelines using messages as input in real-time streaming mode.

![GitHub branch check runs](https://img.shields.io/github/check-runs/datamin-io/kafka-trigger/main?color=green)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/datamin-io/kafka-trigger?color=blue)
<a href="https://github.com/datamin-io/kafka-trigger?tab=Apache-2.0-1-ov-file">![Static Badge](https://img.shields.io/badge/license-Apache%202.0-blue)</a>
<a href="https://datamin.io" target="_blank">![Static Badge](https://img.shields.io/badge/website-datamin.io-blue)</a>
<a href="https://docs.datamin.io" target="_blank">![Static Badge](https://img.shields.io/badge/documentation-docs.datamin.io-blue)</a>
<a href="https://join.slack.com/t/datamincommunity/shared_invite/zt-2nawzl6h0-qqJ0j7Vx_AEHfnB45xJg2Q" target="_blank">![Static Badge](https://img.shields.io/badge/community-join%20Slack-blue)</a>

## How it works
Kafka-trigger listens to Kafka topics and calls the [pipeline run endpoint](https://docs.datamin.io/datamin-api/api-endpoints#run-workflow) for each message, passing the message body as pipeline input.

In the pipeline, the input can be received by the task "External trigger". When the pipeline is run, this task gets the input data and passes it further to the next pipeline task.

## Configuration
Kafka-trigger is configured with environment variables. 
Besides the conventional way, the config variables can also be specified in the `.env` or `.env.local` file.

Main variables:
- **DTMN_KT_API_URL** - Datamin's API URL. If you use the cloud version of Datamin, it is https://api.datamin.io, othervise add URL of your custom Datamin API instance here.
- **DTMN_KT_API_CLIENT_ID** — OAuth client ID.
- **DTMN_KT_API_CLIENT_SECRET** — OAuth client secret.
- **DTMN_KT_KAFKA_VERSION=3.1.0** — version of the Kafka server.
- **DTMN_KT_KAFKA_BOOTSTRAP_SERVERS="127.0.0.1:9092"** — a comma-separated list of Kafka bootstrap servers.
- **DTMN_KT_KAFKA_TOPIC_MAPPING="topic_1:pipeline_uuid_1,topic_1:pipeline_uuid_2,topic_2:pipeline_uuid_2"** — topic-to-pipeline mapping, a comma-separated list of `<topic name>:<pipeline uuid>` pairs.

Other supported env variables are in the [.env.dist file](https://github.com/datamin-io/kafka-trigger/blob/main/.env.dist)

More information about how to obtain your OAuth credentials is in our [documentation](https://docs.datamin.io/datamin-api/oauth-clients)

## Usage

### 1. Create a new topic in Kafka

Let's call it `test_kafka_trigger`

### 2. Create a new pipeline in Datamin

Let's create a simple pipeline that contains only two tasks: External_trigger and Aggregator.

<img width="1402" alt="Screenshot 2024-07-25 at 21 40 55" src="https://github.com/user-attachments/assets/0f20ce9b-cf7e-46fd-8159-b036400370d8">

An Aggregator will just send what it receives as input from Kafka to the output.

<img width="1294" alt="Screenshot 2024-07-25 at 21 41 05" src="https://github.com/user-attachments/assets/60c9d414-d2b9-4d74-810c-2766b08758de">

### 3. Configure the Kafka trigger service

Create the new `.env` and copy the content from `.env.dist` to it and then customize values in the list.

Let's assume, you run this application in a Docker container, with your Kafka set up on the hostmachine port 9092 and you want to trigger pipelines in the Cloud version of Datamin. In Kafka you created a topic called `test_kafka_trigger` and you want to subscribe to the new messages there and trigger the pipeline UUID `8feaae75-7234-4f2f-9e4e-0b491e4a4331`

The .env variables will look like this:

```
DTMN_KT_API_CLIENT_ID=%%REPLACE_IT_WITH_YOUR_CLIENT_ID%%
DTMN_KT_API_CLIENT_SECRET=%%REPLACE_IT_WITH_YOUR_CLIENT_SECRET%%
DTMN_KT_API_URL=https://api.datamin.io
DTMN_KT_KAFKA_BOOTSTRAP_SERVERS="host.docker.internal:9092"
DTMN_KT_KAFKA_TOPIC_MAPPING="test_kafka_trigger:8feaae75-7234-4f2f-9e4e-0b491e4a4331"
```

### 4. Run the container

`docker-compose up --build`

If everything goes well, the service will be able to subscribe to your newly created topic and in the CLI output it will print something like that:

<img width="917" alt="Screenshot 2024-07-25 at 22 12 54" src="https://github.com/user-attachments/assets/650be831-8157-4b7b-9a87-6045be661f60">

### 5. Produce the new message

If you now produce a simple new message to the topic:

<img width="441" alt="Screenshot 2024-07-25 at 22 14 15" src="https://github.com/user-attachments/assets/946128fd-3a50-4aeb-929d-7e7d5e7a32bb">

You will see in the CLI that Kafka trigger successfully consumed it and triggered the pipeline:


<img width="1180" alt="Screenshot 2024-07-25 at 22 15 21" src="https://github.com/user-attachments/assets/0774515f-134d-4961-83eb-d5533c39fe37">


And the pipeline logs also contain a new item stating that pipeline was successfully executed and each of two tasks returned a correct output:

<img width="1335" alt="Screenshot 2024-07-25 at 21 48 22" src="https://github.com/user-attachments/assets/4d092760-821b-4cd1-930e-39361a7c6d9f">

<img width="1330" alt="Screenshot 2024-07-25 at 21 48 32" src="https://github.com/user-attachments/assets/dca0ed64-6aa9-4e55-b379-3d4b6907c7f0">


