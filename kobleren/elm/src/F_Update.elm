module F_Update exposing (update, parseLocation)

import A_Model exposing (Model)
import B_Message exposing (..)
import D_Command as Command
import RDF.Graph as Graph
import Http
import Navigation
import Vendor.UrlParser exposing (..)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Search query ->
            if (String.trim query) == "" then
                ( Model HomeRoute "" "" Nothing [], Cmd.none )
            else
                ( { model | query = query }, Command.doSearch query 0 )

        GetResults (Ok newResults) ->
            ( Model HomeRoute "" model.query (Just newResults) [], Cmd.none )

        GetResults (Err err) ->
            ( Model HomeRoute (stringFromHttpError err) "" Nothing [], Cmd.none )

        Paginate offset ->
            ( model, Command.doSearch model.query offset )

        LoadResource uri ->
            ( model, Command.loadResource uri )

        GetResource (Err err) ->
            ( { model | error = (stringFromHttpError err) }, Cmd.none )

        GetResource (Ok resource) ->
            case Graph.fromString resource of
                Err err ->
                    ( { model | error = err }, Cmd.none )

                Ok triples ->
                    ( { model | graph = triples }, Cmd.none )

        NavigateTo url ->
            ( model, Navigation.newUrl url )

        LocationChange location ->
            let
                newRoute =
                    parseLocation location
            in
                ( { model | route = newRoute }, routeAction newRoute )


routeAction : Route -> Cmd Msg
routeAction route =
    case route of
        HomeRoute ->
            Cmd.none

        EditAuthorityRoute uri type_ ->
            Command.loadResource uri

        NotFoundRoute ->
            Cmd.none


matchers : Parser (Route -> a) a
matchers =
    oneOf
        [ map HomeRoute top
        , map EditAuthorityRoute (s "edit" <?> requiredStringParam "uri" <?> requiredStringParam "type")
        ]


parseLocation : Navigation.Location -> Route
parseLocation location =
    case (parsePath matchers location) of
        Just route ->
            route

        Nothing ->
            NotFoundRoute



-- TODO find another place for stringFromHttpError


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
