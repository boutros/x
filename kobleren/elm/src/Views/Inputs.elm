module Views.Inputs exposing (..)

import Html exposing (Html, div, text, input, label, textarea, datalist, option, span)
import Html.Attributes exposing (class, value, type_, maxlength, style, rows, list, id)
import A_Model exposing (..)
import B_Message exposing (..)
import RDF.RDF as RDF


singleString : Model -> List RDF.TriplePattern -> String -> Html Msg
singleString model patterns inputLabel =
    inputWrap inputLabel
        (input
            [ type_ "text"
            , value (valueFromGraph model patterns)
            ]
            []
        )


singleText : String -> Int -> Html Msg
singleText inputLabel numLines =
    inputWrap inputLabel (textarea [ rows numLines ] [])


singleNumber : Model -> List RDF.TriplePattern -> String -> Int -> Html Msg
singleNumber model patterns inputLabel length =
    inputWrap inputLabel
        (input
            [ type_ "text"
            , value (valueFromGraph model patterns)
            , maxlength length
            , style
                [ ( "width", (toString length) ++ "em" )
                ]
            ]
            []
        )


multiSearchSelect : Model -> List RDF.TriplePattern -> String -> Authority -> Html Msg
multiSearchSelect model patterns inputLabel authority =
    let
        values =
            (valuesFromGraph model patterns)
    in
        inputWrap inputLabel
            (div []
                [ input
                    [ type_ "text"
                    , list inputLabel
                    ]
                    []
                , div [ class "input-values" ] (List.map (\v -> inputValue v authority) values)
                , datalist [ id inputLabel ]
                    (List.map (\( uri, label_ ) -> option [ id uri ] [ text label_ ]) (allValuesFor authority))
                ]
            )


inputValue : String -> Authority -> Html Msg
inputValue v authority =
    div []
        [ span [ class "input-values-delete" ] [ text "×" ]
        , text (labelFor v authority)
        ]


inputWrap : String -> Html Msg -> Html Msg
inputWrap inputLabel input_ =
    div [ class "input-group" ]
        [ div [ class "input-group-label" ]
            [ label [] [ text inputLabel ] ]
        , div [ class "input-group-input" ]
            [ input_
            ]
        ]
