module E_Init exposing (init)

import A_Model exposing (Model)
import B_Message exposing (..)
import Navigation exposing (Location)


init : Location -> ( Model, Cmd Msg )
init location =
    ( { route = HomeRoute, error = "", query = "", results = Nothing }, Cmd.none )
