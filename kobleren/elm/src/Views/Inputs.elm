module Views.Inputs exposing (..)

import Html exposing (Html, div, text, input, label, textarea, datalist, option)
import Html.Attributes exposing (class, value, type_, maxlength, style, rows, list, id)
import B_Message exposing (..)
import C_Data exposing (..)
import RDF.RDF as RDF


singleString : List RDF.TriplePattern -> String -> Html Msg
singleString patterns inputLabel =
    inputWrap inputLabel (input [ type_ "text" ] [])


singleText : String -> Int -> Html Msg
singleText inputLabel numLines =
    inputWrap inputLabel (textarea [ rows numLines ] [])


singleNumber : List RDF.TriplePattern -> String -> Int -> Html Msg
singleNumber patterns inputLabel length =
    inputWrap inputLabel
        (input
            [ type_ "text"
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
