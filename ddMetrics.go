package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"fmt"
	"regexp"
	"strconv"
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)
const (
	cmdTimeout = 2
	loopCount = 10
	writeOnDir = "/tmp"
)
var (
	appVersion string
	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": appVersion,
		},
	})

	ddWriteTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dd_writes_total_v2",
		Help: "Count of all DD writes",
	}, []string{"bs", "count", "result"})

	ddWriteDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dd_writes_duration_seconds_v2",
		Help: "Duration of all dd writes",
	}, []string{"bs", "count"})

	ddWriteThroughput = prometheus.NewGaugeVec(prometheus.GaugeOpts{
                Name: "dd_writes_throughput_MBs_v2",
                Help: "Throughput of all dd writes",
        }, []string{"bs", "count"})
)

func runDD(bs,count,destDir string) {
	// TODO(jproig): improve strconv error handling
	bsArg := fmt.Sprintf("bs=%s", bs)
	countArg := fmt.Sprintf("count=%s", count)
	ofArg := fmt.Sprintf("of=%s/test_file", destDir)
	

	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout*time.Minute)
	defer cancel() // The cancel should be deferred so resources are cleaned up
    cmd := exec.CommandContext(ctx, "dd", "if=/dev/zero", ofArg, bsArg, countArg, "conv=fsync")

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Command timed out")
		ddWriteTotal.WithLabelValues(bs, count, "timeout").Inc()
		return
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		ddWriteTotal.WithLabelValues(bs, count, "err").Inc()
	    log.Printf("cmd.Run() failed with %s\n", err)
		return
	}

	var throughputReMB = regexp.MustCompile(`, ([0-9]*\.[0-9]+|[0-9]+) MB/s`)
    throughputMB := throughputReMB.FindStringSubmatch(string(out))
	if len(throughputMB) > 0 {
		th, _ := strconv.ParseFloat(throughputMB[1],64)
		ddWriteThroughput.WithLabelValues(bs, count).Add(th)
    }
    if len(throughputMB) == 0 {
      var throughputReGB = regexp.MustCompile(`, ([0-9]*\.[0-9]+|[0-9]+) GB/s`)
      throughputGB := throughputReGB.FindStringSubmatch(string(out))
      th, _ := strconv.ParseFloat(throughputGB[1],64)
	  ddWriteThroughput.WithLabelValues(bs, count).Add(th*1000) 
    }

	var durationRe = regexp.MustCompile(`, ([0-9]*\.[0-9]+|[0-9]+) s,`)
    duration := durationRe.FindStringSubmatch(string(out))
    if len(duration) > 0 {
        d, _ := strconv.ParseFloat(duration[1],64)
		ddWriteDuration.WithLabelValues(bs, count).Add(d)
    }

	ddWriteTotal.WithLabelValues(bs, count, "ok").Inc()
	log.Printf("Wrote bs=%s count %s", bs, count)

}
func main() {
	version.Set(1)
	bind := ""
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&bind, "bind", ":8080", "The socket to bind to.")
	flagset.Parse(os.Args[1:])

	r := prometheus.NewRegistry()
	r.MustRegister(ddWriteTotal)
	r.MustRegister(ddWriteDuration)
	r.MustRegister(ddWriteThroughput)
	r.MustRegister(version)

    go func(){
    	time.Sleep(20*time.Second)
    	for n:=1; n <= loopCount; n++ {
		    runDD("1G","1",writeOnDir)
		    runDD("64M","1",writeOnDir)
		    runDD("1M","256",writeOnDir)
		    runDD("8k","10k",writeOnDir)
		    runDD("512","1000",writeOnDir)
		}
		fmt.Printf("Done dding in the disk %s, will stay up so our metrics can be scraped.", writeOnDir)
	}()

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(bind, nil))
}
