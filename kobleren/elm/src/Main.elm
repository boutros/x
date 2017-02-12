module Main exposing (..)

import App exposing (..)
import Html


main =
    Html.program
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        }
