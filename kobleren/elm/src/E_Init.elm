module E_Init exposing (init)

import A_Model exposing (Model)
import B_Message exposing (..)
import F_Update exposing (parseLocation, routeAction)
import Navigation exposing (Location)
import RDF.Graph as Graph


init : Location -> ( Model, Cmd Msg )
init location =
    let
        route =
            parseLocation location
    in
        ( { route = route
          , error = ""
          , query = ""
          , results = Nothing
          , graph = Graph.empty
          }
        , routeAction route
        )
