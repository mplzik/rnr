from enum import Enum
from typing import List, Callable, Optional
import copy
import threading
import sys

import tasks_pb2
from google.protobuf.json_format import MessageToJson

class TaskSchedulingState(Enum):
    PENDING = 1
    RUNNING = 2
    DONE = 3

    @classmethod
    def from_state(cls, state):
        return {
            tasks_pb2.PENDING: TaskSchedulingState.PENDING,
            tasks_pb2.RUNNING: TaskSchedulingState.RUNNING,
            tasks_pb2.SUCCESS: TaskSchedulingState.DONE,
            tasks_pb2.FAILED: TaskSchedulingState.DONE,
            tasks_pb2.SKIPPED: TaskSchedulingState.DONE,
        }[state]


class Task:
    def __init__(self, name: str):
        self.pb = tasks_pb2.Task()
        self.pb.name = name
        self.children: List[Task] = []
        self.pb.state = tasks_pb2.PENDING
        self.pb.message = ""
    
    def poll(self):
        pass

    def as_proto(self):
        ret = copy.deepcopy(self.pb)
        ret.children.extend(list([x.as_proto() for x in self.children]))
        return ret


class Callback(Task):
    def __init__(self, name, callback: Callable[[Task],bool]):
        super().__init__(name)
        self.callback = callback

    def poll(self):
        ret = self.callback(self)

        if ret == True:
            self.pb.state = tasks_pb2.SUCCESS

class Threaded(Task):
    def __init__(self, name, callback: Callable[[Task],bool]):
        super().__init__(name)
        self.callback = callback
        self.thread:Optional[threading.Thread] = None

    def poll(self):
        if not self.thread:
            self.thread = threading.Thread(target=lambda: self.callback(self))
            self.thread.start()

        if not self.thread.is_alive():
            self.thread.join()
            self.pb.state = tasks_pb2.SUCCESS


class Nested(Task):
    def __init__(self, name: str, parallelism = 1):
        super().__init__(name)
        self.parallelism = parallelism

    def add_task(self, task: Task):
        self.children.append(task)

    def poll(self):
        running = 0
        pending_tasks = []

        if TaskSchedulingState.from_state(self.pb.state) != TaskSchedulingState.RUNNING:
            # This might happen if the root task is paused
            return

        # Poll existing tasks
        for child in self.children:
            sched_state = TaskSchedulingState.from_state(child.pb.state)
            if sched_state == TaskSchedulingState.RUNNING:
                try:
                    child.poll()
                except Exception as e:
                    # got an exception, print out error and force task as faled
                    child.pb.state = tasks_pb2.FAILED
                    child.pb.message = str(e)
                    print("Exception!")
                else:
                    if TaskSchedulingState.from_state(child.pb.state) == TaskSchedulingState.RUNNING:
                        running += 1
            if sched_state == TaskSchedulingState.PENDING:
                pending_tasks.append(child)
        
        # Schedule any pending tasks, assuming slots are available
        while running < self.parallelism and len(pending_tasks) > 0:
            child = pending_tasks.pop(0)
            child.pb.state = tasks_pb2.RUNNING
            running += 1

        success_count = len([x for x in self.children if x.pb.state == tasks_pb2.SUCCESS])
        if success_count < len(self.children):
            self.pb.message = "%d/%d" % (success_count, len(self.children))
        else:
            self.pb.message = ""

        # Check if all children are finished
        if all([TaskSchedulingState.from_state(x.pb.state) == TaskSchedulingState.DONE for x in self.children]):
            if all([task.pb.state == tasks_pb2.SUCCESS for task in self.children]):
                self.pb.state = tasks_pb2.SUCCESS
            else:
                self.pb.state = tasks_pb2.FAILED
            return True
        return False
