module Views.Inputs exposing (..)

import Html exposing (Html, div, text, input, label, textarea, datalist, option)
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


singleSearchSelect : String -> List String -> Html Msg
singleSearchSelect inputLabel options =
    inputWrap inputLabel
        (div []
            [ input
                [ type_ "text"
                , list inputLabel
                ]
                []
            , datalist [ id inputLabel ]
                (List.map (\o -> option [] [ text o ]) options)
            ]
        )


inputWrap : String -> Html Msg -> Html Msg
inputWrap inputLabel input_ =
    div [ class "input-group" ]
        [ div [ class "input-group-label" ]
            [ label [] [ text inputLabel ] ]
        , div [ class "input-group-input" ]
            [ input_
            ]
        ]
