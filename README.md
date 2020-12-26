# dd Test - Metrics 

ddMetrics (dd comes from the linux os dd command `man dd`) it is a simple disk write/read testing tool.

The only cool part is that it can be easily run in a docker container and metrics are going to be exposed so a prometheus server can scrape them.

*CAVEAT*: this is not a comprehensive disk deviece benckmarking/testing tool, please do not use it like that.

## Exposed Metrics

* dd_writes_total_v2(counter): total of writes, with bs (BYTES) - count (input blocks) - result (err, ok , timeout) as labels.
* dd_writes_duration_seconds_v2: durations in seconds for all the `ok` writes, with bs (BYTES) - count (input blocks) as labels.
* dd_writes_duration_seconds_v2: durations in seconds for all the `ok` writes, with bs (BYTES) - count (input blocks) as labels.

## How To Run

TOTO(jproig): add instructions on how to run it locally and also in a container.