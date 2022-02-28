module Proto exposing (..)

import Json.Decode exposing (..)
import Json.Encode as Encode
import Json.Decode.Extra exposing (..)

type alias Task = { name : String, state : String, message : String, children: Children }
type Children = Children (List Task)
type alias Job = { version: Int, uuid : String, root : Task }

type TaskState = Unknown | Pending | Running | Success | Failed | Skipped | ActionNeeded

taskStateStrings : List (TaskState, String)
taskStateStrings = [ (Unknown, "UNKNOWN"), (Pending, "PENDING"), (Running, "RUNNING"), (Success, "SUCCESS"), (Failed, "FAILED"), (Skipped, "SKIPPED"), (ActionNeeded, "ACTION_NEEDED") ]

taskStateToString : TaskState -> String
taskStateToString ts = List.filter (\(xts, _) -> xts == ts) taskStateStrings |> List.map Tuple.second |> List.head |> Maybe.withDefault ""
taskStateFromString : String -> Maybe TaskState
taskStateFromString s = List.filter (\(_, x) -> x == s) taskStateStrings |> List.map Tuple.first |> List.head

jobDecoder : Decoder Job
jobDecoder = 
    map3 Job
      (field "version" parseInt)
      (field "uuid" string)
      (field "root" taskDecoder)

childrenDecoder : Decoder Children
childrenDecoder = 
  Json.Decode.map Children
    (Json.Decode.list (lazy (\_ -> taskDecoder)))

taskDecoder : Decoder Task
taskDecoder =
    map4 Task
      (field "name" string)
      (field "state" string)
      (field "message" string)
      (field "children" childrenDecoder)

type alias TaskRequest = { path: List String, state: TaskState }

taskRequestEncoder : TaskRequest -> Encode.Value
taskRequestEncoder td = Encode.object
    [ ("path", Encode.list Encode.string td.path)
    , ("state", Encode.string <| taskStateToString td.state)]