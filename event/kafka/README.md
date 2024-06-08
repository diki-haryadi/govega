# Kafka

Event Adapter for Kafka

## Sender

Available config for kafka event sender adapter.

## Config

| Config          | Type                 | Required | Description                                                                                                                                                                                                                                                                                                                                                        |
|-----------------|----------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| brokers         | Array of string      | Yes      | List of kafka brokers                                                                                                                                                                                                                                                                                                                                              |
| batch_size      | int                  | No       | Limit on how many messages will be buffered before being sent to a partition.<br><br>Default: 100 messages                                                                                                                                                                                                                                                         |
| batch_timeout   | string               | No       | Time limit on how often incomplete message batches will be flushed to kafka.<br><br>Default: 1s                                                                                                                                                                                                                                                                    |
| key_file        | string               | No       | The path to the key file location (required to connect to kafka brokers with tls)                                                                                                                                                                                                                                                                                  |
| cert_file       | string               | No       | The path to the certification file location (required to connect to kafka brokers with tls)                                                                                                                                                                                                                                                                        |
| ca_cert         | string               | No       | The path to the CA (certificate authority) file location (required to connect to kafka brokers with tls)                                                                                                                                                                                                                                                           |
| auth_type       | string               | No       | The authentication type to connect with kafka brokers using username / password mechanism.<br><br>Valid value:<br>- plain<br>- scram<br><br>Default: plain                                                                                                                                                                                                         |
| username        | string               | No       | The username to be used to connect to the brokers (required for plain and scram mechanism)                                                                                                                                                                                                                                                                         |
| password        | string               | No       | The password to be used to connect to the brokers (required for plain and scram mechanism)                                                                                                                                                                                                                                                                         |
| max_attempts    | int                  | No       | Limit on how many attempts will be made to deliver a message.<br><br>Default: 10                                                                                                                                                                                                                                                                                   |
| balancer        | string               | No       | The balancer used to distribute messages across partitions.<br><br>Please refer to [Balancer Section](#balancer) for available balancer.<br><br>Default: least_bytes                                                                                                                                                                                               |
| balancer_config | map[string]interface | No       | Additional config for balancer                                                                                                                                                                                                                                                                                                                                     |
| print_log_level | string               | No       | Log level for all log message other than error message.<br><br>Due to the implementation using golib log library, this will also follow the golib log minimum log level,<br>So if you set the value higher than the minimum level it will not be printed in the log.<br><br>Please refer to [Log Level](#log-level) for the available level.<br><br>Default: debug |
| error_log_level | string               | No       | Log level for error message.<br><br>Due to the implementation using golib log library, this will also follow the golib log minimum log level,<br>So if you set the value higher than the minimum level it will not be printed in the log.<br><br>Please refer to [Log Level](#log-level) for the available level.<br><br>Default: error                            |

## Listener

Available config for kafka event listener adapter.

| Config                   | Type            | Required | Description                                                                                                                                                                                                                                                                                  |
|--------------------------|-----------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| brokers                  | Array of string | Yes      | List of kafka brokers                                                                                                                                                                                                                                                                        |
| key_file                 | string          | No       | The path to the key file location (required to connect to kafka brokers with tls)                                                                                                                                                                                                            |
| cert_file                | string          | No       | The path to the certification file location (required to connect to kafka brokers with tls)                                                                                                                                                                                                  |
| ca_cert                  | string          | No       | The path to the CA (certificate authority) file location (required to connect to kafka brokers with tls)                                                                                                                                                                                     |
| auth_type                | string          | No       | The authentication type to connect with kafka brokers using username / password mechanism.<br><br>Valid value: <br>- plain<br>- scram<br><br><br>Default: plain                                                                                                                              |
| username                 | string          | No       | The username to be used to connect to the brokers (required for plain and scram mechanism)                                                                                                                                                                                                   |
| password                 | string          | No       | The password to be used to connect to the brokers (required for plain and scram mechanism)                                                                                                                                                                                                   |
| print_log_level          | string          | No       | Log level for all log message other than error message.<br><br>Due to the implementation using golib log library, this will also follow the golib log minimum log level,<br>So if you set the value higher than the minimum level it will not be printed in the log.  <br><br>Default: debug |
| error_log_level          | string          | No       | Log level for error message.<br><br>Due to the implementation using golib log library, this will also follow the golib log minimum log level,<br>So if you set the value higher than the minimum level it will not be printed in the log.  <br><br>Default: error                            |
| queue_capacity           | int             | No       | The capacity of the internal message queue.<br><br>Default: 100                                                                                                                                                                                                                              |
| min_bytes                | int             | No       | Indicates to the broker the minimum batch size that the consumer<br>will accept. Setting a high minimum when consuming from a low-volume topic<br>may result in delayed delivery when the broker does not have enough data to<br>satisfy the defined minimum.<br><br>Default: 1              |
| max_bytes                | int             | No       | Indicates to the broker the maximum batch size that the consumer<br>will accept. The broker will truncate a message to satisfy this maximum, so<br>choose a value that is high enough for your largest message size.<br><br>Default: 1MB                                                     |
| max_wait                 | string          | No       | Maximum amount of time to wait for new data to come when fetching batches<br>of messages from kafka.<br><br>Default: 10s                                                                                                                                                                     |
| max_attempts             | int             | No       | Limit of how many attempts will be made before delivering the error.<br><br>Default: 3                                                                                                                                                                                                       |
| read_lag_interval        | string          | No       | Sets the frequency at which the reader lag is updated.<br>Setting this field to a negative value disables lag reporting.<br><br>Default: 1m                                                                                                                                                  |
| heartbeat_interval       | string          | No       | sets the optional frequency at which the reader sends the consumer<br>group heartbeat update.<br><br>Default: 3s                                                                                                                                                                             |
| commit_interval          | string          | No       | Indicates the interval at which offsets are committed to<br>the broker.  If 0, commits will be handled synchronously.<br><br>Default: 0                                                                                                                                                      |
| partition_watch_interval | string          | No       | Indicates how often a reader checks for partition changes.<br>If a reader sees a partition change (such as a partition add) it will rebalance the group<br>picking up new partitions.<br><br>Default: 5s                                                                                     |
| watch_partition_changes  | bool            | No       | Is used to inform kafka-go that a consumer group should be<br>polling the brokers and rebalancing if any partition changes happen to the topic.                                                                                                                                              |
| session_timeout          | string          | No       | Optionally sets the length of time that may pass without a heartbeat<br>before the coordinator considers the consumer dead and initiates a rebalance.<br><br>Default: 30s                                                                                                                    |
| rebalance_timeout        | string          | No       | Optionally sets the length of time the coordinator will wait<br>for members to join as part of a rebalance.  For kafka servers under higher<br>load, it may be useful to set this value higher.<br><br>Default: 30s                                                                          |
| join_group_backoff       | string          | No       | Optionally sets the length of time to wait between re-joining<br>the consumer group after an error.<br><br>Default: 5s                                                                                                                                                                       |
| retention_time           | string          | No       | RetentionTime optionally sets the length of time the consumer group will be saved<br>by the broker<br><br>Default: 24h                                                                                                                                                                       |
| start_offset             | int             | No       | Determines from whence the consumer group should begin<br>consuming when it finds a partition without a committed offset.  If<br>non-zero, it must be set to one of FirstOffset (-2) or LastOffset (-1).<br><br>Default: FirstOffset (-2)                                                    |
| read_backoff_min         | string          | No       | Optionally sets the smallest amount of time the reader will wait before<br>polling for new messages<br><br>Default: 100ms                                                                                                                                                                    |
| read_backoff_max         | string          | No       | Optionally sets the maximum amount of time the reader will wait before<br>polling for new messages<br><br>Default: 1s                                                                                                                                                                        |

## Additional Config

### Balancer

| Type        | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | Additional Config                            |
|-------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------|
| hash        | Hash is a Balancer that uses the provided hash function to determine which<br>partition to route messages to.  This ensures that messages with the same key<br>are routed to the same partition.<br><br>The logic to calculate the partition is:<br><br>		hasher.Sum32() % len(partitions) => partition<br><br>By default, Hash uses the FNV-1a algorithm.  This is the same algorithm used<br>by the Sarama Producer and ensures that messages produced by kafka-go will<br>be delivered to the same topics that the Sarama producer would be delivered to                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | -                                            |
| round_robin | RoundRobin is an Balancer implementation that equally distributes messages<br>across all available partitions.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | -                                            |
| crc32       | CRC32Balancer is a Balancer that uses the CRC32 hash function to determine<br>which partition to route messages to.  This ensures that messages with the<br>same key are routed to the same partition.  This balancer is compatible with<br>the built-in hash partitioners in librdkafka and the language bindings that<br>are built on top of it, including the<br>github.com/confluentinc/confluent-kafka-go Go package.<br><br>With the Consistent field false (default), this partitioner is equivalent to<br>the "consistent_random" setting in librdkafka.  When Consistent is true, this<br>partitioner is equivalent to the "consistent" setting.  The latter will hash<br>empty or nil keys into the same partition.<br><br>Unless you are absolutely certain that all your messages will have keys, it's<br>best to leave the Consistent flag off.  Otherwise, you run the risk of<br>creating a very hot partition.                                                                                                                                                                                                                                                                                                                                                                    | [crc32 config](#crc32-additional-config)     |
| murmur2     | Murmur2Balancer is a Balancer that uses the Murmur2 hash function to<br>determine which partition to route messages to.  This ensures that messages<br>with the same key are routed to the same partition.  This balancer is<br>compatible with the partitioner used by the Java library and by librdkafka's<br>"murmur2" and "murmur2_random" partitioners. /<br><br>With the Consistent field false (default), this partitioner is equivalent to<br>the "murmur2_random" setting in librdkafka.  When Consistent is true, this<br>partitioner is equivalent to the "murmur2" setting.  The latter will hash<br>nil keys into the same partition.  Empty, non-nil keys are always hashed to<br>the same partition regardless of configuration.<br><br>Unless you are absolutely certain that all your messages will have keys, it's<br>best to leave the Consistent flag off.  Otherwise, you run the risk of<br>creating a very hot partition.<br><br>Note that the librdkafka documentation states that the "murmur2_random" is<br>functionally equivalent to the default Java partitioner.  That's because the<br>Java partitioner will use a round robin balancer instead of random on nil<br>keys.  We choose librdkafka's implementation because it arguably has a larger<br>install base. | [murmur2 config](#murmur2-additional-config) |
| least_bytes | LeastBytes is a Balancer implementation that routes messages to the partition<br>that has received the least amount of data.<br><br>Note that no coordination is done between multiple producers, having good<br>balancing relies on the fact that each producer using a LeastBytes balancer<br>should produce well balanced messages.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | -                                            |


#### crc32 Additional Config

| Config     | Type | Required | Description                                                                                                                                                                   |
|------------|------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| consistent | bool | No       | When consistent is false, message with empty key will be send into random partition,<br>but when is set to true, message with empty key will be hash into the same partition. |

#### murmur2 Additional Config

| Config     | Type | Required | Description                                                                                                                                                                   |
|------------|------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| consistent | bool | No       | When consistent is false, message with empty key will be send into random partition,<br>but when is set to true, message with empty key will be hash into the same partition. |

### Log Level
| Level   | Description                                           |
|---------|-------------------------------------------------------|
| discard | No log message will be printed                        |
| debug   | Only log with level debug and above will be printed   |
| info    | Only log with level info and above will be printed    |
| warning | Only log with level warning and above will be printed |
| error   | Only log with level error and above will be printed   |
| fatal   | Only log with level fatal and above will be printed   |
| panic   | This will work exactly as fatal level                 |