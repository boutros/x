module App exposing (..)

import Html exposing (Html, text, div, a, p, input, h2, h3, ul, li, pre, nav, main_, strong, span, label)
import Html.Attributes exposing (type_, value, attribute, class, id, for)
import Html.Events exposing (onInput, onClick)
import Http
import Json.Decode exposing (Decoder, field, at, string, int, float, dict, list, nullable)
import Json.Decode.Pipeline exposing (decode, required, requiredAt, optional)


-- MODEL


type alias Model =
    { error : String
    , query : String
    , results : Maybe SearchResults
    }


init : ( Model, Cmd Msg )
init =
    ( { error = "", query = "", results = Nothing }, Cmd.none )


type alias SearchResults =
    { offset : Int
    , totalHits : Int
    , hits : List SearchHit
    }


type alias SearchHit =
    { id : String
    , label : String
    , authorityType : String
    , abstract : String
    }



-- UPDATE


type Msg
    = Search String
    | GetResults (Result Http.Error SearchResults)
    | Paginate Int


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Search query ->
            ( { model | query = query }, doSearch query 0 )

        GetResults (Ok newResults) ->
            ( Model "" model.query (Just newResults), Cmd.none )

        GetResults (Err err) ->
            ( Model (stringFromHttpError err) "" Nothing, Cmd.none )

        Paginate offset ->
            ( model, doSearch model.query offset )



-- VIEW


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


viewSearchResults : SearchResults -> Html Msg
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


viewSearchHit : SearchHit -> Html never
viewSearchHit a =
    div
        [ class "search-hit" ]
        [ p []
            [ span [ class "search-hit-authority-type" ]
                [ text a.authorityType ]
            , strong [ class "search-hit-title" ]
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



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- HTTP


doSearch : String -> Int -> Cmd Msg
doSearch query offset =
    let
        url =
            "http://localhost:8008/search?q=" ++ query ++ "&from=" ++ (toString offset)
    in
        Http.send GetResults (Http.get url decodeResults)


decodeResults : Json.Decode.Decoder SearchResults
decodeResults =
    decode SearchResults
        |> required "Offset" int
        |> required "TotalHits" int
        |> optional "Hits" (list decodeSearchHit) []


decodeSearchHit : Json.Decode.Decoder SearchHit
decodeSearchHit =
    decode SearchHit
        |> required "ID" string
        |> required "Label" string
        |> required "Type" string
        |> required "Abstract" string


stringFromHttpError : Http.Error -> String
stringFromHttpError e =
    case e of
        Http.BadUrl msg ->
            "Bad URL" ++ msg

        Http.Timeout ->
            "Timeout"

        Http.NetworkError ->
            "Network Error"

        Http.BadPayload msg response ->
            "Bad Payload: " ++ msg

        Http.BadStatus response ->
            "Bad Reponse: " ++ response.status.message ++ " : " ++ response.body
