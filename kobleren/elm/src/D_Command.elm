module D_Command exposing (..)

import B_Message exposing (..)
import C_Data exposing (decodeResults, decodeResource)
import Http


-- HTTP CALLS


doSearch : String -> Int -> Cmd Msg
doSearch query offset =
    let
        url =
            "http://localhost:8008/search?q=" ++ query ++ "&from=" ++ (toString offset)
    in
        Http.send GetResults (Http.get url decodeResults)


loadResource : String -> Cmd Msg
loadResource uri =
    let
        id =
            removePrefix "http://data.deichman.no" uri

        url =
            "http://localhost:8008" ++ id
    in
        Http.send GetResource (Http.getString url)



-- Helper functions (TODO move out)


removePrefix : String -> String -> String
removePrefix prefix s =
    if (String.startsWith prefix s) then
        String.dropLeft (String.length prefix) s
    else
        s
