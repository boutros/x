module B_Message exposing (..)

import C_Data as Data
import Http


type Msg
    = Search String
    | GetResults (Result Http.Error Data.SearchResults)
    | Paginate Int
