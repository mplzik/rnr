from google.protobuf.json_format import MessageToJson
from flask import Flask, jsonify, render_template, request, send_from_directory, Response
import threading
import task
import tasks_pb2
import uuid

app = Flask(__name__)
app.config['JSONIFY_PRETTYPRINT_REGULAR'] = True
root_task = None
run_uuid = uuid.uuid4()

@app.route('/')
def index():
    return send_from_directory("../ui", "index.html")

@app.route('/tasks', methods=["GET", "POST"])
def tasks():
    if request.method == "POST":
        data = request.get_json()
        print(data)
        
        # Look up task
        t = root_task
        for idx in data["path"]:
            t = t.children[idx]
        
        # process the data
        if "state" in data:
            state = task.TaskState[data["state"]]
            t.state = state


    return Response(MessageToJson(
        tasks_pb2.Job(
            version = 1,
            uuid = str(run_uuid),
            root= root_task.as_proto(),
        ), including_default_value_fields=True
    ), mimetype="application/json")

def init(t: task.Task):
    global root_task
    root_task = t
    
    threading.Thread(target=lambda: app.run(port=8080)).start()
