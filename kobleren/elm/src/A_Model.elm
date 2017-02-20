module A_Model exposing (Model)

import C_Data as Data


type alias Model =
    { error : String
    , query : String
    , results : Maybe Data.SearchResults
    }
