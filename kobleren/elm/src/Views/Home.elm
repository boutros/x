module Views.Home exposing (view)

import Html exposing (Html, Attribute, text, div, a, p, input, h2, h3, ul, li, pre, nav, main_, strong, span, label)
import Html.Attributes exposing (type_, value, attribute, class, id, for)
import Html.Events exposing (onInput, onClick)
import A_Model exposing (Model)
import B_Message exposing (..)
import C_Data as Data
import Views.Helpers exposing (link)


view : Model -> Html Msg
view model =
    let
        content =
            case model.results of
                Just searchResults ->
                    viewSearchResults searchResults

                Nothing ->
                    div [] []
    in
        div []
            [ nav []
                [ h2 [] [ text "Vedlikehold av autoriteter" ]
                , input
                    [ class "authority-search-box"
                    , attribute "size" "14"
                    , type_ "search"
                    , onInput Search
                    ]
                    []
                , h2 [] [ text "Nytt materiale" ]
                , ul [ class "linkify" ]
                    [ li []
                        [ a [] [ text "Bok" ]
                        ]
                    , li []
                        [ a [] [ text "E-bok" ]
                        ]
                    , li []
                        [ a [] [ text "Lydbok" ]
                        ]
                    , li []
                        [ a [] [ text "Tegneserie" ]
                        ]
                    , li []
                        [ a [] [ text "Film" ]
                        ]
                    , li []
                        [ a [] [ text "Musikkopptak" ]
                        ]
                    , li []
                        [ a [] [ text "Musikknoter" ]
                        ]
                    , li []
                        [ a [] [ text "Spill" ]
                        ]
                    , li []
                        [ a [] [ text "Språkkurs" ]
                        ]
                    ]
                ]
            , main_
                []
                [ p
                    []
                    [ text model.error ]
                , content
                ]
            ]


viewSearchResults : Data.SearchResults -> Html Msg
viewSearchResults results =
    div [ class "search-results" ]
        [ h3 []
            [ text ((toString results.totalHits) ++ " treff") ]
        , (viewSearchPagination results.offset results.totalHits)
        , div
            [ class "clearfix" ]
            (List.map viewSearchHit results.hits)
        , div [ class "clearfix" ] []
        ]


viewSearchHitAbstract : String -> String -> Html never
viewSearchHitAbstract abstract uri =
    let
        lines =
            (String.split "\n" abstract)

        length =
            (List.length lines)
    in
        if length == 0 then
            div [] []
        else if length <= 5 then
            div [] [ text abstract ]
        else
            div [ class "relative search-hit-abstract-more" ]
                [ input [ id uri, type_ "checkbox" ] []
                , label [ for uri ] []
                , div [ class "search-hit-abstract-text" ] [ text abstract ]
                ]


viewSearchHit : Data.SearchHit -> Html Msg
viewSearchHit a =
    div
        [ class "search-hit" ]
        [ p []
            [ span [ class "search-hit-authority-type" ]
                [ text a.authorityType ]
            , link ("/edit?uri=" ++ a.id ++ "&type=" ++ a.authorityType)
                [ class "search-hit-title"
                ]
                [ text a.label ]
            ]
        , div [ class "search-hit-abstract" ]
            [ (viewSearchHitAbstract a.abstract a.id) ]
        ]


viewSearchPagination : Int -> Int -> Html Msg
viewSearchPagination offset total =
    let
        hitsPerPage =
            10

        numPages =
            if total == 0 then
                0
            else
                min (ceiling ((toFloat total) / hitsPerPage)) 10

        activePage =
            (offset // hitsPerPage) + 1
    in
        if total <= hitsPerPage then
            div [] []
            -- TODO figure out how to do express this without this dummy else-branch
        else
            div
                [ class "search-pagination" ]
                [ ul []
                    (List.map
                        (\i ->
                            if i == activePage then
                                li [ class "search-pagination-active-page" ]
                                    [ strong
                                        []
                                        [ text (toString i) ]
                                    ]
                            else
                                li
                                    [ class "search-pagination-page linkify"
                                    , onClick
                                        (Paginate ((i - 1) * hitsPerPage))
                                    ]
                                    [ text
                                        (toString i)
                                    ]
                        )
                        (List.range 1 (numPages))
                    )
                ]
