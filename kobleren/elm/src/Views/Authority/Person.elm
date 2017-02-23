module Views.Authority.Person exposing (view)

import Html exposing (Html, div, text, p, main_, nav, h2)
import A_Model exposing (..)
import B_Message exposing (..)


view : Model -> Html Msg
view model =
    div []
        [ nav []
            [ h2 [] [ text "Vedlikeholde autoritet: person" ]
            ]
        , main_
            []
            [ p
                []
                [ text "her " ]
            ]
        ]
