module C_Data exposing (..)

import Json.Decode exposing (Decoder, field, at, string, int, float, dict, list, nullable)
import Json.Decode.Pipeline exposing (decode, required, requiredAt, optional)


-- TYPES


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



-- DECODERS


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
