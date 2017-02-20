module Main exposing (..)

import E_Init exposing (init)
import F_Update exposing (update)
import G_View exposing (view)
import Html


main =
    Html.program
        { init = init
        , view = view
        , update = update
        , subscriptions = (\_ -> Sub.none)
        }
