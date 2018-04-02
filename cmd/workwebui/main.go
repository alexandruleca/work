package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/alexandruleca/work/webui"
	"strings"
)

var (
	redisHostPort  = flag.String("redis", ":6379", "redis hostport")
	redisDatabase  = flag.String("database", "0", "redis database")
	redisNamespace = flag.String("ns", "work", "redis namespace")
	webHostPort    = flag.String("listen", ":5040", "hostport to listen for HTTP JSON API")
)

func resolveNamespaceInput() (result []string) {
	splitString := strings.Split(*redisNamespace, ",")
	for _, s := range splitString {
		result = append(result, strings.TrimSpace(s))
	}
	return
}

func main() {
	flag.Parse()

	namespaces := resolveNamespaceInput()

	fmt.Println("Starting workwebui:")
	fmt.Println("redis = ", *redisHostPort)
	fmt.Println("database = ", *redisDatabase)
	fmt.Printf("namespace = %+v", namespaces)
	fmt.Println("")
	fmt.Println("listen = ", *webHostPort)

	database, err := strconv.Atoi(*redisDatabase)
	if err != nil {
		fmt.Printf("Error: %v is not a valid database value", *redisDatabase)
		return
	}

	pool := newPool(*redisHostPort, database)

	server := webui.NewServer(namespaces, pool, *webHostPort)
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c

	server.Stop()

	fmt.Println("\nQuitting...")
}

func newPool(addr string, database int) *redis.Pool {
	return &redis.Pool{
		MaxActive:   3,
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(addr, redis.DialDatabase(database))
		},
		Wait: true,
	}
}
