module RDF.Graph exposing (fromString)

import RDF.RDF exposing (TriplePattern)
import RDF.Parser exposing (parseTriples)
import Parser


fromString : String -> Result Parser.Error (List TriplePattern)
fromString ntriples =
    Parser.run (parseTriples ntriples) ntriples



-- TODO fromString : String -> Result (List TriplePattern)
-- with serialized error message
