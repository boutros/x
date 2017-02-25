module Views.Authority.Person exposing (view)

import Html exposing (Html, div, text, p, main_, nav, h2, fieldset)
import Html.Attributes exposing (class)
import A_Model exposing (..)
import B_Message exposing (..)
import Views.Inputs as Inputs


view : Model -> String -> Html Msg
view model uri =
    div []
        [ nav []
            [ h2 [] [ text "Person" ]
            , div [] [ text "uri" ]
            ]
        , main_
            []
            [ div
                [ class "resource-edit" ]
                [ fieldset []
                    [ (Inputs.singleString "Navn")
                    , (Inputs.singleNumber "Fødselsår" 4)
                    , (Inputs.singleNumber "Dødsår" 4)
                    , (Inputs.singleSearchSelect "Nasjonalitet" [ "Norge", "Sverige", "Danmark", "Senegal" ])
                    , (Inputs.singleString "Kjønn")
                    , (Inputs.singleText "Forklarende tilføyelse" 3)
                    , (Inputs.singleString "Alternativt navn")
                    , (Inputs.singleNumber "Nummer" 4)
                    ]
                ]
            ]
        ]
