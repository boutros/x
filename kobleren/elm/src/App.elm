module App exposing (..)

import Html exposing (Html, text, div, p, input, h2, h3, ul, li, pre, nav, main_, strong, span)
import Html.Attributes exposing (type_, value, attribute, class)
import Html.Events exposing (onInput)
import Http
import Json.Decode exposing (Decoder, field, at, string, int, float, dict, list, nullable)
import Json.Decode.Pipeline exposing (decode, required, requiredAt, optional)


-- MODEL


type alias Model =
    { error : String
    , results : Maybe SearchResults
    }


init : ( Model, Cmd Msg )
init =
    ( { error = "", results = Nothing }, Cmd.none )


type alias SearchResults =
    { took : Int
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


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Search query ->
            ( model, doSearch query )

        GetResults (Ok newResults) ->
            ( Model "" (Just newResults), Cmd.none )

        GetResults (Err err) ->
            ( Model (stringFromHttpError err) Nothing, Cmd.none )



-- VIEW


view : Model -> Html Msg
view model =
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
            ]
        , main_
            []
            [ div [ class "search-results" ]
                [ p
                    []
                    [ text model.error ]
                , h3 []
                    [ case model.results of
                        Just res ->
                            text (toString res.totalHits ++ " treff")

                        Nothing ->
                            text ""
                    ]
                , div
                    []
                    [ viewSearchResults model.results ]
                ]
            ]
        ]


viewSearchResults : Maybe SearchResults -> Html never
viewSearchResults results =
    case results of
        Just searchResults ->
            div [] (List.map viewSearchHit searchResults.hits)

        Nothing ->
            text ""


viewSearchHit : SearchHit -> Html never
viewSearchHit a =
    div
        [ class "search-hit" ]
        [ p []
            [ span [ class "search-hit-authority-type" ]
                [ text a.authorityType ]
            , strong []
                [ text a.label ]
            , p [ class "search-hit-abstract" ]
                [ text a.abstract ]
            ]
        ]



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- HTTP


doSearch : String -> Cmd Msg
doSearch query =
    let
        url =
            "http://localhost:8008/dummysearch?q=" ++ query
    in
        Http.send GetResults (Http.get url decodeResults)


decodeResults : Json.Decode.Decoder SearchResults
decodeResults =
    decode SearchResults
        |> required "Took" int
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
