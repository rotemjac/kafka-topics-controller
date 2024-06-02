package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"strconv"
	"time"
)

type Topic struct {
	Name              string `yaml:"Name"`
	Partitions        int    `yaml:"Partitions"`
	Replicas          int    `yaml:"Replicas"`
	MinInSyncReplicas string `yaml:"MinInSyncReplicas"`
	RetentionBytes    string `yaml:"RetentionBytes"`
	RetentionMs       string `yaml:"RetentionMs"`
}

var logger = logrus.New()

func main() {
	// ENV vars
	tenantId := os.Getenv("TENANT_ID")
	bootstrapServer := os.Getenv("BOOTSTRAP_SERVER")
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	configFileName := os.Getenv("CONFIG_FILE_NAME")
	//logLevel := os.Getenv("LOG_LEVEL")
	debugSleep := os.Getenv("DEBUG_SLEEP")

	// Init logging configuration
	logger.SetOutput(os.Stdout)       // Output to stdout instead of the default stderr , Can be any io.Writer, see below for File example
	logger.SetLevel(logrus.InfoLevel) // Only log the warning severity or above.

	// Connect to Kafka Brokers
	logger.Info("Connecting to Kafka broker")
	conn, err := kafka.Dial("tcp", bootstrapServer)
	if err != nil {
		panic(err.Error())
		logger.Error(err.Error())
	}
	defer conn.Close()

	// Read Topics Config and Create Topics
	topics, _ := readTopicsFromFile(configFilePath, configFileName)
	for _, topic := range topics {
		topicName := tenantId + "-" + topic.Name
		createTopic(conn, Topic{
			topicName,
			topic.Partitions,
			topic.Replicas,
			topic.MinInSyncReplicas,
			topic.RetentionBytes,
			topic.RetentionMs,
		})
	}

	// Sleep if needed
	value, err := strconv.ParseBool(debugSleep)
	if err != nil && value {
		logger.Error("Error parsing environment variable:", err)
	} else {
		logger.Info("Sleeping..")
		time.Sleep(3600000)
	}
}

func createTopic(conn *kafka.Conn, topic Topic) {
	controller, err := conn.Controller()
	if err != nil {
		panic(err.Error())
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err.Error())
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic.Name,
			NumPartitions:     topic.Partitions,
			ReplicationFactor: topic.Replicas,
			ConfigEntries: []kafka.ConfigEntry{
				{ConfigName: "min.insync.replicas", ConfigValue: fmt.Sprint(topic.MinInSyncReplicas)},
				{ConfigName: "retention.bytes", ConfigValue: fmt.Sprint(topic.RetentionBytes)},
				{ConfigName: "retention.ms", ConfigValue: fmt.Sprint(topic.RetentionMs)},
			},
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}
}

func readTopicsFromFile(configFilePath string, configFileName string) ([]Topic, error) {

	// Construct the full path to the YAML file
	yamlPath := filepath.Join(configFilePath, configFileName)

	// Read the YAML file
	dataInYaml, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	} else {
		logger.Info("Succeed to read YAML")
	}

	// Create a new Config object to store the parsed data
	var topics []Topic
	err = yaml.Unmarshal(dataInYaml, &topics) // Output will be placed in topics
	if err != nil {
		logger.Info("Failed to unmarshal YAML:", err)
	} else {
		logger.Info("Succeed to unmarshal YAML.")
		logger.Info("Printing topics according to the fields: ")
		logger.Info("Name, Partitions, Replicas, MinInSyncReplicas, RetentionBytes, RetentionMs")
		logger.Info(topics)
	}
	return topics, nil
}

func getAllTopics(conn *kafka.Conn) map[string]struct{} {
	partitions, err := conn.ReadPartitions()
	if err != nil {
		panic(err.Error())
	}
	m := map[string]struct{}{}
	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	//for k := range m {fmt.Println(k)}
	return m
}
