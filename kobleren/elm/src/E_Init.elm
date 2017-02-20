module E_Init exposing (init)

import A_Model exposing (Model)
import B_Message exposing (Msg)


init : ( Model, Cmd Msg )
init =
    ( { error = "", query = "", results = Nothing }, Cmd.none )
