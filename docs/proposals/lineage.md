# Data lineage

To track data lineage across pipelines, KubeETL provides a metadata-tracking service.
This service works by attaching a sidecar to containers
and receiving signals from said container conforming to a yet to be determined data lineage spec.

## Questions

* How do containers provide data lineage information?
* What is the data lineage specification?
* How do we present this information?
* Does the lineage service only provide "high level" lineage about which processing steps happened, or do we track individual batches of data?
