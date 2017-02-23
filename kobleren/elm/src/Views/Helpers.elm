module Views.Helpers exposing (..)

import B_Message exposing (..)
import Html exposing (Html, Attribute, a)
import Html.Events exposing (onInput, onClick, onWithOptions, defaultOptions)
import Json.Decode as Json


link : String -> List (Attribute Msg) -> List (Html Msg) -> Html Msg
link route attributes children =
    let
        clickHandler =
            onWithOptions "click" defaultOptions <| Json.succeed <| NavigateTo route

        attrs =
            clickHandler :: attributes
    in
        a
            attrs
            children
