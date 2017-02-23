module Main exposing (..)

import B_Message exposing (..)
import E_Init exposing (init)
import F_Update exposing (update)
import G_View exposing (view)
import Navigation


main =
    Navigation.program
        LocationChange
        { init = init
        , view = view
        , update = update
        , subscriptions = (\_ -> Sub.none)
        }
