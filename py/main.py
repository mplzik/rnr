#!/usr/bin/env python

import subprocess
import argparse
import sys
import time
import random

import task
import tasks_pb2
import webserver

def hello_task(idx):
    t = task.Nested("Hello %s" % idx)
    t.add_task(task.Threaded("Wait a moment", lambda _: time.sleep(1)))
    t.add_task(task.Callback("Hello %d" % idx, lambda x: print("Hello from '%s'" % x.pb.name) or True))
    t.add_task(task.Callback("Sometimes fails", lambda _: 1/0 if random.randint(0, 2) == 0 else True ))

    return t

def main(argv):
    parser = argparse.ArgumentParser(description="Runs a pipeline of commands")    
    args = parser.parse_args(argv[1:])
    print(args)

    t = task.Nested("Greetings", parallelism=2)
    for i in range(0, 100):
        t.add_task(hello_task(i))

    webserver.init(t)
    t.pb.state = tasks_pb2.RUNNING

    while True:
        if task.TaskSchedulingState.from_state(t.pb.state) == task.TaskSchedulingState.PENDING:
            t.pb.state = task.TaskState.RUNNING
        if task.TaskSchedulingState.from_state(t.pb.state) != task.TaskSchedulingState.DONE:
            t.poll()
        
        time.sleep(1)

main(sys.argv)
sys.exit(0)