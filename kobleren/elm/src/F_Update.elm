module F_Update exposing (update)

import A_Model exposing (Model)
import B_Message exposing (..)
import D_Command as Command
import Http


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Search query ->
            ( { model | query = query }, Command.doSearch query 0 )

        GetResults (Ok newResults) ->
            ( Model "" model.query (Just newResults), Cmd.none )

        GetResults (Err err) ->
            ( Model (stringFromHttpError err) "" Nothing, Cmd.none )

        Paginate offset ->
            ( model, Command.doSearch model.query offset )



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
