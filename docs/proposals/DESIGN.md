# KubeETL Design Principles & Design Goals

This document lays down the design principles & design goals of the KubeETL project.

## Design Goals

KubeETL aims to simplify the Data Engineering lifecycle. In Data Engineering the same steps are often executed over and over again for each data workflow:

- Extract: extracting data is often the start of a data workflow. For each data source, a data workflow requires data source information, data source credentials & information on the protocol used to communicate with the data source. Often this information is replicated for each data workflow, especially when multiple different teams have data engineering initiatives. Another aspect of extracting data involves data source profiling, assessing the quality of the data source. If this information is not captured centrally and maintained regularly, data quality issues will arise.
- Transform: After the data is extracted at a certain point it needs to be transformed to serve it's purpose. Transformation can aim to clean the data, restructure the data in a more tangible form (e.g. a dimensional model) or prepare it for machine learning workflows. Often transformations contain a lot of business logic that is hard to capture and hard to reuse. This means the same transformation are often duplicated in multiple data workflows and lineage information of specific transformation are hard to track. Additionally transformations often require schema information of the data source that is manually replicated in different locations. Besides that transformations often have complicated infrastructure requirements. Compute resources need to be available to run the transformations on, scaling with the size of the data. Transformations often also require a specific order, where later stages of a workflow build on the data produced in earlier stages. This means that there are dependencies between stages and this needs to be maintained properly. If an earlier stage fails, later stages should not run until the earlier stages succeed.
- Load: loading data has similar issues as extracting data. Information on the data sink & its protocols are needed before data can be loaded into the data sink. The format of the data loaded into the data sink also needs to adhere to the expected schema in the data sink.
- Metadata: Along the way, relevant metadata is produced that is rarely captured but often really valuable for a multitude of reasons. Schemas of the data produced by workflows could be captured and reused in other workflows. Metadata on data transformations & the data itself can be used to handle compliance & data quality issues. Workflow executions can be tracked to gather information on resource usage & prevent operational issues. Capturing all this information correctly is hard.

## Design Principles

- **KubeETL is opinionated, but extensible:** KubeETL should provide sane defaults wherever possible. Configuration should be possible, but only needed if there is an exceptional case.
- **Common Cases should be easy:** standard ETL workflows, such as common ELT flows with Singer, DBT & a Data Warehouse should be easily build using KubeETL.
- **Exceptional Cases should be possible:** while we focus on the common cases, we should still allow for unforeseen use cases of KubeETL by exposing lower level primitives that can be used to construct new workflows.
- **Libraries over YAML:** KubeETL should expose it's functionality through libraries as opposed to just exposing YAML interfaces.
- **High reusability:** workflows & tasks within those workflows should be reusable for other workflows. KubeETL should also make it easy to share intermediate data stages to create reusable data sources.
- **Make metadata accessible:** KubeETL should provide functionality to capture the metadata needed to cover common data quality measurement & lineage tracking requirements.
- **Reuse existing tools where possible:** the goal is to not reinvent the wheel, but build on existing open source tools & components.
