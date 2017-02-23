module E_Init exposing (init)

import A_Model exposing (Model)
import B_Message exposing (..)
import F_Update exposing (parseLocation)
import Navigation exposing (Location)


init : Location -> ( Model, Cmd Msg )
init location =
    ( { route = (parseLocation location), error = "", query = "", results = Nothing }, Cmd.none )
