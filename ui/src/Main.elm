module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Time
import Proto exposing (..)
import Html.Events exposing (onClick)

-- MAIN


main =
  Browser.element
    { init = init
    , update = update
    , subscriptions = subscriptions
    , view = view
    }

-- MODEL


type Model
  = Failure String
  | Loading
  | Loaded Job


init : () -> (Model, Cmd Msg)
init _ =
  ( Loading
  , updateTasks
  )



-- UPDATE


type Msg
  = GotJob (Result Http.Error Job)
  | Tick Time.Posix
  | PostTaskRequest (List String) String
  | TaskRequestPosted (Result Http.Error ())

update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    GotJob result ->
      case result of
        Ok task ->
          (Loaded task, Cmd.none)

        Err errmsg ->
          (Failure (Debug.toString errmsg), Cmd.none)
    Tick _ -> (model, updateTasks)
    PostTaskRequest path state -> (model, Http.post
      { url = "/tasks"
      , body = Http.jsonBody (Proto.taskRequestEncoder { path = path, state = (taskStateFromString state |> Maybe.withDefault Proto.Unknown)} )
      , expect = Http.expectWhatever TaskRequestPosted })
    TaskRequestPosted _ -> (model, Cmd.none)

updateTasks : Cmd Msg
updateTasks = Http.get
      { url = "/tasks"
      , expect = Http.expectJson GotJob jobDecoder
      }

-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
  Time.every 1000 Tick



-- VIEW


view : Model -> Html Msg
view model =
  case model of
    Failure msg ->
      Html.text ("Error communicating with backend: " ++ msg)

    Loading ->
      Html.text "Loading..."

    Loaded job -> viewTask [] job.root

viewTask : List String -> Task -> Html Msg
viewTask path task = 
  let 
    (Children children) = task.children
  in
    if List.length children > 0 then
      details [] ([ 
        summary [] [ viewTaskHeadline path task ], 
        text task.state,
        ul [attribute "style" "list-style-type: none"] (List.map (\child -> li [] [viewTask (path ++ [child.name]) child]) children)
      ])
    else
      viewTaskHeadline path task

taskStyle : Task -> List (Attribute Msg)
taskStyle task = case task.state of
  "PENDING" -> [ attribute "style" "color: grey" ]
  "RUNNING" -> [ attribute "style" "font-weight: bold" ]
  "SUCCESS" -> [ attribute "style" "color: green" ]
  "FAILED" -> [ attribute "style" "color: darkred" ]
  "ACTION_NEEDED" -> [ attribute "style" "color: orange" ]
  _ -> []

viewTaskHeadline : List String -> Task -> Html Msg
viewTaskHeadline path task = span [] [ 
  span (taskStyle task) [ viewTaskState path task, text " ", text task.name ]
  , text " ", i [] [text task.message]
  ]
  
viewTaskState : List String -> Task -> Html Msg
viewTaskState path task = select [ Html.Events.onInput (PostTaskRequest path) ] (
  List.map (\(ts, s) -> option [Html.Attributes.selected (task.state == s) ] [ text s ]) 
  taskStateStrings)