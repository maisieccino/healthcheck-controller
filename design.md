# Design

The idea behind the Health Operator is to have a system that provides a clear view of the health of a system. Probes are a 
good way to health check containers but they're not good enough to validate the state or how healthy an application is.
Health checking an application is not just checking that a service responds (200 OK), but it might need a deeper analysis. For
example, you might want to define that an application is healthy if certain conditions are met, this operator is the foundation
of how to define and implement those conditions and how to expose them.

Health checks are basically a set of CronJobs that validate certain conditions at certain intervals. This effort is about
automating the lifecycle of these checks.

There are 3 different use cases:

* Kubernetes service that has to be health checked (`annotation healthcheck=true`)
* Ad-hoc check that executes one or more than one health check (kind: HealthCheck)
* Probes result gathering

There are two approaches: 

* From Dashboard to CronJobs
* From CronJobs to Dashboard

[ Matt Diagram ] <-- modify

Let's get the Dashboard to CronJobs approach. The idea of the Dashboard is to have a place where you can see the health of a
system. This means gathering a set of results in a regular basis and expose them.

What does it mean, a system is healthy?

Is it that the pods are repsonding as expected? (--> gathering probes result)
Is it that services have endpoints and can fullfill a response? (--> creating cronjobs that check services)
Is it that we can assert that one or more conditions are true? `yes, users can log in into our system`

## Services

The idea behind this service-intrsopection that generates CronJobs is very simple. We can have a list of common ports that are
tested based on some pre-defined rules:

* 80: HTTP request/response
* 443: HTTP request/response
* 8000: HTTP request/response
* 8080: HTTP request/response
* 5433: TCP request/response

and so on. Other services using other ports might need to use an annotation to define the type of check needed.

## Labels and Grouping
