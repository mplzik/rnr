# Rnr

## Overview

`rnr` is a library that tries to address the most common pains of writing workflow-based ops tooling -- scheduling, control and visualization.

Ops tooling is often implemented as a set of shellscripts and small utilities that can be mostly controlled and inspected using CLI. While this works for small and simple tools, it does not scale for larger workflows that take hours or even days to execute -- these can no longer be safely ran on local workstations or left alone to do their thing with only crude means of control (a.k.a. `Ctrl-C`).

`rnr` tries to work around this by providing a tooling to create tasks that can be componsed into jobs. Behind the scenes, `rnr` manages state of the tasks, exposes API to retrieve and modify it and ensures correct scheduling. The human operator can leverage the full-fledged programming language, libraries and tests, while `rnr` handles some of the boring bits.

## Why not CI/CD ?

CI/CD is definitely has its place in automation, but comes short in some areas:

- CI/CD requires statically defined pipelines. There is no easy way to dynamically adjust pipelines based on i.e. currently running Kubernetes pods or VMs running workloads.
- CI/CD often takes a docker container as a minimum running unit. Although this makes sense in some cases, it also introduces a lot of boilerplace for environment setup just to create another step in the workflow. If few lines of code could perform the same workflow, it might be a more viable option.
- CI/CD do not offer much control over the pipeline execution. The execution can be paused, any closer fiddling (such as skipping a step or pausing a certain branch of execution) is often a no-go.
- CI/CD don't facilitate testing of any kind -- that is often left up to the end-user who should validate the steps.

There's no strict boundary, though -- some workflows are better implemented as regular code, the other might be better off as containerized steps in CI/CD. There's also no strict requirement to not cross the boundary -- in some cases it might be possible to rewrite a `rnr` workflow to a CI/CD pipeline (or vice versa), launch CI/CD from `rnr` program, or use `rnr` binaries as a part of CI/CD steps.

## The concepts

### Task

`Task` is the smallest building block in `rnr`. It represents a unit of work that can be stopped, running, succeed, failed, ... . Tasks can have their own child tasks, formning hierarchies. Parent task is responsible for scheduling child tasks. State of each task is represtented by a [protobuf](proto3/rnr.proto).

It is possible to change task's state externally using HTTP API, and thus the task should not make any assumptions on the state itself.

Currently, there are at least these _task states_ defined in the protobuf: `UNKNOWN`, `PENDING`, `RUNNING`, `SUCCESS`, `FAILED`, `SKIPPED`, `ACTION_PENDING`. For scheduling purposes, these states are translated to three _scheduling states_ -- `PENDING` (waits to become running), `RUNNING` (currently running), `DONE` (excluded from scheduling).

### Job

`Job` represents a root data structure that holds a reference to a root task.

### Polling

Polling is the main mechanism of refreshing state of a job's progress in `rnr`. Internally, tasks are being periodically polled and are expected to update their state accordingly. The choice of polling comes as a conservative and simple decision. This by no means discourages the use of any more complex mechanisms if they're more suitable.

In general, there is no guarantee on how often `rnr` will poll the tasks. If a job depends on being polled in a constant interval, ensure it by its own means.

### Local statelessness

`rnr` jobs should be resilient to restarts of the binary as much as possible; `rnr` library itself currently doesn't persist the state of the job in any way. A recommended pattern to work around this is to store the state externally -- as close to the source of truth as possible.

*Example:* when upgrading a package on a virtual machine, the recommended steps to execute each poll would be:

1. Retrieve the current version of the package; if it's correct, mark task as successfuly completed
2. Verify whether the package upgrade process is already running.
3. If upgrade process is not yet running, launch it.

The first step ensures that a task will succeed if the upgrade succeeded previously. The second makes sure that if there's already an upgrade that we lost track of running, the process will learn about it. The third step ensures that the upgrade will get launched if needed.

## Types of tasks

The `Task` type contains a fair amount of `rnr`-internal implementation details and is harder to work with. To simplify the development, there are two wrappers around this type -- `CallbackTask` and `NestedTask`.

### CallbackTask

A task that calls the provided callback handler with each poll. This provides some shortcuts that use callback's return value to configure task's state appropriately.

### NestedTask

Nested tasks are used to schedule multiple child tasks. With each Poll, all the children that have either changed their state or are running will getd `Poll`-ed, ensuring that at most `parallelism` tasks is running at once. If more tasks is running i.e. due to manual changes, new tasks won't get scheduled until a sufficient number of tasks terminates.

## Example

See i.e. [the example golang code](golang/main.go) .