package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/couchbase/gocb/v2"
)

func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%s:\n", label)
	fmt.Printf("  Alloc = %v MB\n", m.Alloc/1024/1024)
	fmt.Printf("  TotalAlloc = %v MB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("  Sys = %v MB\n", m.Sys/1024/1024)
	fmt.Printf("  NumGC = %v\n", m.NumGC)
	fmt.Println()
}

func main() {
	runtime.GC()
	printMemStats("Before Connect")

	// Connect to cluster
	cluster, err := gocb.Connect("couchbase://localhost", gocb.ClusterOptions{
		Username: "Administrator",
		Password: "password",
	})
	if err != nil {
		panic(err)
	}
	defer cluster.Close(nil)

	time.Sleep(2 * time.Second) // 연결이 완전히 확립되도록 대기
	runtime.GC()
	printMemStats("After Connect")

	// Open bucket
	bucket := cluster.Bucket("default")
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)
	runtime.GC()
	printMemStats("After Bucket Open")

	collection := bucket.DefaultCollection()

	// Perform some operations
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		_, err := collection.Upsert(key, map[string]interface{}{
			"value": i,
			"data":  "test data",
		}, nil)
		if err != nil {
			fmt.Printf("Upsert error: %v\n", err)
		}
	}

	runtime.GC()
	printMemStats("After 1000 Upserts")

	fmt.Println("\n=== Connection Info ===")
	// 클러스터에 대한 진단 정보
	diagnostics, err := cluster.Diagnostics(nil)
	if err == nil {
		fmt.Printf("Cluster ID: %s\n", diagnostics.ID)
		fmt.Printf("Number of KV endpoints: %d\n", len(diagnostics.State["kv"]))
		for _, endpoint := range diagnostics.State["kv"] {
			fmt.Printf("  - %s: %s\n", endpoint.Remote, endpoint.State)
		}
	}

	time.Sleep(5 * time.Second)
}
