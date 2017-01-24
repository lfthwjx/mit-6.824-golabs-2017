package mapreduce

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

// doReduce manages one reduce task: it reads the intermediate
// key/value pairs (produced by the map phase) for this task, sorts the
// intermediate key/value pairs by key, calls the user-defined reduce function
// (reduceF) for each key, and writes the output to disk.
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	//
	// You will need to write this function.
	//
	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTaskNumber) yields the file
	// name from map task m.
	//
	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.
	//
	// You may find the first example in the golang sort package
	// documentation useful.
	//
	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.
	//
	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
	var kvs []KeyValue
	for i := 0; i < nMap; i++ {
		f := reduceName(jobName, i, reduceTaskNumber)
		file, err := os.Open(f)
		if err != nil {
			log.Printf("os.Open(%s) Failed: %s", f, err)
			return
		}
		defer file.Close()
		d := json.NewDecoder(file)
		var kv KeyValue
		err = d.Decode(&kv)
		for err == nil {
			kvs = append(kvs, kv)
			err = d.Decode(&kv)
		}
	}
	if len(kvs) == 0 {
		log.Printf("0 len")
		return
	}
	sort.Sort(ByKey(kvs))
	var rst []KeyValue
	j := 1
	for j <= len(kvs) {
		var values []string
		for j < len(kvs) && kvs[j].Key == kvs[j-1].Key {
			values = append(values, kvs[j-1].Value)
			j++
		}
		values = append(values, kvs[j-1].Value)
		v := reduceF(kvs[j-1].Key, values)
		kv := KeyValue{Key: kvs[j-1].Key, Value: v}
		rst = append(rst, kv)
		j++
	}
	file, err := os.Create(outFile)
	if err != nil {
		log.Printf("os.Open(%s) Failed: %s", outFile, err)
		return
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	for _, kv := range rst {
		enc.Encode(kv)
	}
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }
