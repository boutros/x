module D_Command exposing (..)

import B_Message exposing (..)
import C_Data exposing (decodeResults)
import Http


-- HTTP CALLS


doSearch : String -> Int -> Cmd Msg
doSearch query offset =
    let
        url =
            "http://localhost:8008/search?q=" ++ query ++ "&from=" ++ (toString offset)
    in
        Http.send GetResults (Http.get url decodeResults)
